package discord

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

// deferThinking adia a resposta (o Discord mostra "pensando..."). Retorna false
// se falhar, registrando o erro.
func deferThinking(s *discordgo.Session, i *discordgo.InteractionCreate, logger *slog.Logger) bool {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		logger.Error("falha ao adiar resposta", "erro", err)
		return false
	}
	return true
}

// replyEphemeral responde imediatamente com uma mensagem visível só ao autor.
func replyEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, logger *slog.Logger, msg string) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg, Flags: discordgo.MessageFlagsEphemeral},
	}); err != nil {
		logger.Error("falha ao responder interação", "erro", err)
	}
}

// editContent edita a resposta adiada com um texto simples.
func editContent(s *discordgo.Session, i *discordgo.InteractionCreate, logger *slog.Logger, msg string) {
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg}); err != nil {
		logger.Error("falha ao editar resposta", "erro", err)
	}
}

// editEmbed edita a resposta adiada com um embed.
func editEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, logger *slog.Logger, embed *discordgo.MessageEmbed) {
	embeds := []*discordgo.MessageEmbed{embed}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &embeds}); err != nil {
		logger.Error("falha ao editar resposta com embed", "erro", err)
	}
}

func optionString(opts []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	for _, o := range opts {
		if o.Name == name {
			return o.StringValue()
		}
	}
	return ""
}
