package music_commands

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/foxeiz/namelessgo/src/commands"
	"github.com/foxeiz/namelessgo/src/extractors"
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

func tracksToOptions(tracks []*extractors.TrackInfo) []discordgo.SelectMenuOption {
	options := make([]discordgo.SelectMenuOption, 0, len(tracks))
	for i, t := range tracks {
		options = append(options, discordgo.SelectMenuOption{
			Label:   t.Title,
			Value:   strconv.Itoa(i),
			Default: false,
		})
	}
	return options
}

// ----- //

type PickTrackHandler struct {
	AfterCallback func()
	Timeout       int
}

func (p *PickTrackHandler) SetAfterCallback(callback func()) {
	p.AfterCallback = callback
}

func (p *PickTrackHandler) Handle(
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	tracks []*extractors.TrackInfo,
) {
	data := interaction.MessageComponentData()
	player := GetPlayer(interaction.GuildID)

	if player == nil {
		return
	}

	if len(data.Values) == 0 {
		return
	}

	for _, v := range data.Values {
		index, err := strconv.Atoi(v)
		if err != nil {
			return
		}
		player.AddTrack(tracks[index])
	}

	go session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Added %d track(s) to queue", len(data.Values)),
		},
	})
	go p.AfterCallback()
}

func NewPickTrack(
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	tracks []*extractors.TrackInfo,
) {
	compID := fmt.Sprintf("track.%s", interaction.ID)
	compHandler := &PickTrackHandler{Timeout: 10}
	removeHander := session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionMessageComponent {
			if i.MessageComponentData().CustomID == compID {
				compHandler.Handle(s, i, tracks)
			}
		}
	})
	compHandler.SetAfterCallback(removeHander)

	_, err := session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
		Content: "Pick a track",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						MenuType:  discordgo.StringSelectMenu,
						CustomID:  compID,
						Options:   tracksToOptions(tracks),
						MaxValues: len(tracks),
					},
				},
			},
		},
	})

	if err != nil {
		panic(err)
	}
}

// ----- //

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
	if len(trackList) == 1 {
		player.AddTrack(trackList[0])
	}
	if trackList[0].Site == "ytsearch" {
		NewPickTrack(session, interaction, trackList)
		return
	}

	for _, track := range trackList {
		session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
			Content: track.Title,
		})
	}

}
