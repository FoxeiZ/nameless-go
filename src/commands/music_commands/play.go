package music_commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/foxeiz/namelessgo/src/commands"
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

func Play(
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	options commands.CommandOptionsType,
) {
	if session == nil || interaction == nil || options == nil {
		return
	}

	query, ok := options["query"]
	if !ok || query == nil {
		return
	}

	CheckVoiceAndMaybeJoin(session, interaction, options, true)

	player := GetPlayer(interaction.GuildID)
	if player == nil {
		return
	}

	trackList, err := player.SearchTracks(query.StringValue())
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

	for _, track := range trackList {
		session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
			Content: track.Title,
		})
	}

	if len(trackList) == 1 {
		player.AddTrack(trackList[0])
	}
}
