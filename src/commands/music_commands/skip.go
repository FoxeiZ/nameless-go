package music_commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/foxeiz/namelessgo/src/commands"
)

var SkipCommand = &discordgo.ApplicationCommand{
	Name:        "skip",
	Description: "Skip current track",
}

func Skip(
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	options commands.CommandOptionsType,
) {
	if player := GetPlayer(interaction.GuildID); player != nil {
		player.Stop()
		session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
			Content: "Skipped. Next track should be playing",
		})
	} else {
		session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
			Content: "Not connected to voice channel",
		})
	}
}
