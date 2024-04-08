package music_commands

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/foxeiz/namelessgo/src/types"
)

var ConnectCommand = &discordgo.ApplicationCommand{
	Name:        "connect",
	Description: "Connect to voice channel",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionChannel,
			Name:        "channel",
			Description: "Channel for the bot to connect to",
			Required:    false,
			ChannelTypes: []discordgo.ChannelType{
				discordgo.ChannelTypeGuildVoice,
				discordgo.ChannelTypeGuildStageVoice,
			},
		},
	},
}

func findUserVoiceState(session *discordgo.Session, userid string) (*discordgo.VoiceState, error) {
	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userid {
				return vs, nil
			}
		}
	}
	return nil, errors.New("could not find user's voice state")
}

func joinUserVoiceChannel(session *discordgo.Session, userId string) (*discordgo.VoiceConnection, error) {
	vs, err := findUserVoiceState(session, userId)
	if err != nil {
		return nil, err
	}

	return session.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, true)
}

func Connect(session *discordgo.Session, interaction *discordgo.InteractionCreate, options types.CommandOptionsType) {
	var err error

	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	var vc *discordgo.VoiceConnection
	if channel, ok := options["channel"]; ok {
		vc, err = session.ChannelVoiceJoin(
			interaction.GuildID,
			channel.ChannelValue(nil).ID,
			false,
			true,
		)
	} else {
		vc, err = joinUserVoiceChannel(session, interaction.Member.User.ID)
	}

	var content string
	if err != nil {
		content = "Could not connect to voice channel"
	} else {
		content = "Connected to voice channel"
		AddPlayer(interaction.GuildID, NewPlayer(
			interaction.ChannelID,
			vc,
			func(p *Player, err error) {
				if err != nil {
					session.ChannelMessageSend(p.channelID, err.Error())
				}
			},
			func(p *Player) {
				p.Next()
			},
		))
	}

	session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
}
