package music_commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/foxeiz/namelessgo/src/types"
)

var PlayCommand = &discordgo.ApplicationCommand{
	Name:        "play",
	Description: "Add or search track(s) to queue",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "query",
			Description: "Search query",
			Required:    true,
		},
	},
}

func Play(session *discordgo.Session, interaction *discordgo.InteractionCreate, options types.CommandOptionsType) {
	var query string

	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	_, ok := session.VoiceConnections[interaction.GuildID]
	if !ok {
		session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
			Content: "Not connected to voice channel",
		})
		return
	}

	if q, ok := options["query"]; ok {
		query = q.StringValue()
	}

	trackList, err := GetPlayer(interaction.GuildID).SearchTracks(query)
	if err != nil {
		session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
			Content: err.Error(),
		})
		return
	}

	if len(trackList) == 0 {
		session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
			Content: "No track found",
		})
		return
	}

	if len(trackList) > 1 {
		for _, track := range trackList {
			session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
				Content: track.Title,
			})
		}
	} else {
		session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
			Content: trackList[0].Title,
		})

		GetPlayer(interaction.GuildID).AddTrack(trackList[0])
	}
}
