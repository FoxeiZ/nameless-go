package music_commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/foxeiz/namelessgo/src/commands"
)

var DisconnectCommand = &discordgo.ApplicationCommand{
	Name:        "disconnect",
	Description: "Disconnect from voice channel",
}

func Disconnect(
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	_ commands.CommandOptionsType,
) {
	var content string
	player := GetPlayer(interaction.GuildID)
	if player != nil {
		player.Cleanup()
		content = "Bye! :wave:"
	} else {
		content = "Not connected to voice channel"
	}

	session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
}
