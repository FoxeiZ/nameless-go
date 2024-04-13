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
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	var content string
	vc, ok := session.VoiceConnections[interaction.GuildID]
	if ok {
		vc.Disconnect()
		content = "Bye! :wave:"
	} else {
		content = "Not connected to voice channel"
	}

	session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
}
