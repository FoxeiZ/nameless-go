package general_commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/foxeiz/namelessgo/src/commands"
)

var PingCommand = &discordgo.ApplicationCommand{
	Name:        "ping",
	Description: "Pong?",
}

func Ping(session *discordgo.Session, interaction *discordgo.InteractionCreate, _ commands.CommandOptionsType) {
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Pong! %s", session.HeartbeatLatency().String()),
		},
	})
}
