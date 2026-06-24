// Package discord encapsula a conexão com o gateway do Discord e o ciclo de
// vida da sessão (abrir/fechar).
package discord

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

// Bot gerencia a sessão do Discord.
type Bot struct {
	session *discordgo.Session
	logger  *slog.Logger
}

// New cria a sessão com intents mínimos. O bot é pull-only via slash commands,
// então não precisa de intents privilegiados (conteúdo de mensagem, membros).
func New(token string, logger *slog.Logger) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("criando sessão do discord: %w", err)
	}
	session.Identify.Intents = discordgo.IntentsGuilds

	return &Bot{session: session, logger: logger}, nil
}

// Open registra o handler de Ready e abre a conexão com o gateway.
func (b *Bot) Open() error {
	b.session.AddHandler(func(_ *discordgo.Session, r *discordgo.Ready) {
		b.logger.Info("conectado ao gateway do discord",
			"usuario", r.User.Username,
			"guilds", len(r.Guilds),
		)
	})

	if err := b.session.Open(); err != nil {
		return fmt.Errorf("abrindo conexão com o gateway: %w", err)
	}
	return nil
}

// InteractionHandler trata eventos de interação do Discord.
type InteractionHandler interface {
	Handle(*discordgo.Session, *discordgo.InteractionCreate)
}

// AddInteractionHandler registra um handler de interações. Deve ser chamado
// antes de Open para não perder eventos.
func (b *Bot) AddInteractionHandler(h InteractionHandler) {
	b.session.AddHandler(h.Handle)
}

// Close encerra a conexão com o gateway.
func (b *Bot) Close() error {
	if err := b.session.Close(); err != nil {
		return fmt.Errorf("fechando conexão com o gateway: %w", err)
	}
	return nil
}
