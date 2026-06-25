package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// precoCommand define o slash command /preco. O nome usa "preco" (sem cedilha)
// para respeitar as restrições de nome de comando do Discord.
var precoCommand = &discordgo.ApplicationCommand{
	Name:        "preco",
	Description: "Consulta os preços de um jogo de PC",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "jogo",
			Description: "Nome do jogo a consultar",
			Required:    true,
		},
	},
}

// leaksCommand define o slash command /leaks. O termo é opcional: sem ele,
// retorna os vazamentos mais recentes.
var leaksCommand = &discordgo.ApplicationCommand{
	Name:        "leaks",
	Description: "Lista vazamentos e rumores recentes de games",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "termo",
			Description: "Filtra por um termo (ex.: nome de um jogo)",
			Required:    false,
		},
	},
}

// commands reúne todos os slash commands registrados pelo bot.
var commands = []*discordgo.ApplicationCommand{precoCommand, leaksCommand}

// RegisterCommands registra (sobrescrevendo) o conjunto de comandos. Com guildID
// preenchido o registro é por guild (propaga na hora); vazio = global.
func (b *Bot) RegisterCommands(appID, guildID string) error {
	if _, err := b.session.ApplicationCommandBulkOverwrite(appID, guildID, commands); err != nil {
		return fmt.Errorf("registrando comandos: %w", err)
	}
	b.logger.Info("comandos registrados",
		"guild_scoped", guildID != "",
		"total", len(commands),
	)
	return nil
}
