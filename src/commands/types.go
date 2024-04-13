package commands

import (
	"github.com/bwmarrin/discordgo"
)

type CommandOptionsType map[string]*discordgo.ApplicationCommandInteractionDataOption

type CommandType struct {
	Handlers map[string]func(
		session *discordgo.Session,
		interaction *discordgo.InteractionCreate,
		options CommandOptionsType,
	)
	Commands []*discordgo.ApplicationCommand
}
