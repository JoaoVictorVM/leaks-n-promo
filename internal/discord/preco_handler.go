package discord

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/JoaoVictorVM/leaks-n-promo/internal/cache"
	"github.com/JoaoVictorVM/leaks-n-promo/internal/price"
)

const (
	// handlerTimeout limita a duração total da busca após o defer da resposta.
	handlerTimeout = 15 * time.Second
	// maxEmbedOffers limita quantas ofertas são exibidas no embed.
	maxEmbedOffers = 10
	embedColor     = 0x1b2838
)

// PrecoHandler atende o slash command /preco.
type PrecoHandler struct {
	provider price.PriceProvider
	cache    *cache.Cache[string, []price.Offer]
	logger   *slog.Logger
}

// NewPrecoHandler cria o handler com suas dependências.
func NewPrecoHandler(provider price.PriceProvider, c *cache.Cache[string, []price.Offer], logger *slog.Logger) *PrecoHandler {
	return &PrecoHandler{provider: provider, cache: c, logger: logger}
}

// Handle roteia a interação: ignora o que não for o /preco e responde às demais.
func (h *PrecoHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	data := i.ApplicationCommandData()
	if data.Name != precoCommand.Name {
		return
	}

	game := strings.TrimSpace(optionString(data.Options, "jogo"))
	if game == "" {
		replyEphemeral(s, i, h.logger, "Informe o nome de um jogo. Exemplo: `/preco jogo: Celeste`")
		return
	}

	// A busca pode demorar; adiamos a resposta (Discord mostra "pensando...").
	if !deferThinking(s, i, h.logger) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
	defer cancel()

	offers, err := h.lookup(ctx, game)
	switch {
	case err != nil:
		h.logger.Error("falha na busca de preço", "jogo", game, "erro", err)
		editContent(s, i, h.logger, "Não consegui consultar os preços agora. Tente novamente em instantes.")
	case len(offers) == 0:
		editContent(s, i, h.logger, fmt.Sprintf("Nenhum resultado encontrado para **%s**.", game))
	default:
		editEmbed(s, i, h.logger, buildPrecoEmbed(game, offers))
	}
}

// lookup consulta o cache e, em caso de miss, o provider — memorizando o
// resultado. A chave é o termo normalizado.
func (h *PrecoHandler) lookup(ctx context.Context, game string) ([]price.Offer, error) {
	key := strings.ToLower(game)
	if cached, ok := h.cache.Get(key); ok {
		return cached, nil
	}

	offers, err := h.provider.Search(ctx, game)
	if err != nil {
		return nil, err
	}

	h.cache.Set(key, offers)
	return offers, nil
}

func buildPrecoEmbed(game string, offers []price.Offer) *discordgo.MessageEmbed {
	title := game
	if len(offers) > 0 && offers[0].Title != "" {
		title = offers[0].Title
	}

	shown := offers
	if len(shown) > maxEmbedOffers {
		shown = shown[:maxEmbedOffers]
	}

	var b strings.Builder
	for _, o := range shown {
		fmt.Fprintf(&b, "**%s** — $%.2f", o.Store, o.Price)
		if o.Retail > o.Price {
			fmt.Fprintf(&b, " ~~$%.2f~~", o.Retail)
		}
		if o.URL != "" {
			fmt.Fprintf(&b, " — [ver oferta](%s)", o.URL)
		}
		b.WriteByte('\n')
	}

	return &discordgo.MessageEmbed{
		Title:       title,
		Description: b.String(),
		Color:       embedColor,
		Footer:      &discordgo.MessageEmbedFooter{Text: "Preços em USD • via CheapShark"},
	}
}
