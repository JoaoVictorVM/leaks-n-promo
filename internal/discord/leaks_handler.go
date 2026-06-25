package discord

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/JoaoVictorVM/leaks-n-promo/internal/cache"
	"github.com/JoaoVictorVM/leaks-n-promo/internal/leaks"
)

const maxEmbedLeaks = 10

// LeaksHandler atende o slash command /leaks.
type LeaksHandler struct {
	source leaks.LeakSource
	cache  *cache.Cache[string, []leaks.Leak]
	logger *slog.Logger
}

// NewLeaksHandler cria o handler com suas dependências.
func NewLeaksHandler(source leaks.LeakSource, c *cache.Cache[string, []leaks.Leak], logger *slog.Logger) *LeaksHandler {
	return &LeaksHandler{source: source, cache: c, logger: logger}
}

// Handle roteia a interação: ignora o que não for o /leaks e responde às demais.
func (h *LeaksHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	data := i.ApplicationCommandData()
	if data.Name != leaksCommand.Name {
		return
	}

	term := strings.TrimSpace(optionString(data.Options, "termo"))

	if !deferThinking(s, i, h.logger) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
	defer cancel()

	items, err := h.lookup(ctx, term)
	switch {
	case err != nil:
		h.logger.Error("falha na busca de leaks", "termo", term, "erro", err)
		editContent(s, i, h.logger, "Não consegui buscar os vazamentos agora. Tente novamente em instantes.")
	case len(items) == 0:
		editContent(s, i, h.logger, noLeaksMessage(term))
	default:
		editEmbed(s, i, h.logger, buildLeaksEmbed(term, items))
	}
}

// lookup consulta o cache e, em miss, a fonte agregada — memorizando o resultado.
// A chave é o termo normalizado (vazio = mais recentes).
func (h *LeaksHandler) lookup(ctx context.Context, term string) ([]leaks.Leak, error) {
	key := strings.ToLower(term)
	if cached, ok := h.cache.Get(key); ok {
		return cached, nil
	}

	items, err := h.source.Fetch(ctx, term)
	if err != nil {
		return nil, err
	}

	h.cache.Set(key, items)
	return items, nil
}

func noLeaksMessage(term string) string {
	if term == "" {
		return "Nenhum vazamento recente encontrado."
	}
	return fmt.Sprintf("Nenhum vazamento encontrado para **%s**.", term)
}

func buildLeaksEmbed(term string, items []leaks.Leak) *discordgo.MessageEmbed {
	title := "Vazamentos recentes"
	if term != "" {
		title = "Vazamentos: " + term
	}

	shown := items
	if len(shown) > maxEmbedLeaks {
		shown = shown[:maxEmbedLeaks]
	}

	var b strings.Builder
	for _, it := range shown {
		text := it.Title
		if text == "" {
			text = it.URL
		}
		fmt.Fprintf(&b, "• [%s](%s)", text, it.URL)
		if it.Source != "" {
			fmt.Fprintf(&b, " — *%s*", it.Source)
		}
		if !it.Published.IsZero() {
			fmt.Fprintf(&b, " · %s", it.Published.Format("02/01/2006"))
		}
		b.WriteByte('\n')
	}

	return &discordgo.MessageEmbed{
		Title:       title,
		Description: b.String(),
		Color:       embedColor,
		Footer:      &discordgo.MessageEmbedFooter{Text: "via RSS e Reddit"},
	}
}
