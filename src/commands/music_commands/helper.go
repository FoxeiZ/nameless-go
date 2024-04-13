package music_commands

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/foxeiz/namelessgo/src/commands"
	p "github.com/foxeiz/namelessgo/src/commands/music_commands/player"
)

var playerMapper = make(map[string]*p.Player)
var lock sync.Mutex

func CheckVoiceAndMaybeJoin(
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	options commands.CommandOptionsType,
	autoJoin bool,
) bool {
	if player := GetPlayer(interaction.GuildID); player != nil {
		return true
	}

	if autoJoin {
		Connect(session, interaction, options)
		if player := GetPlayer(interaction.GuildID); player != nil {
			return true
		}
		return false
	}

	session.FollowupMessageCreate(interaction.Interaction, true, &discordgo.WebhookParams{
		Content: "Not connected to voice channel",
	})
	return false
}

func GetPlayer(guildID string) *p.Player {
	lock.Lock()
	defer lock.Unlock()
	return playerMapper[guildID]
}

func AddPlayer(guildID string, player *p.Player) {
	lock.Lock()
	defer lock.Unlock()
	playerMapper[guildID] = player
}

func RemovePlayer(guildID string) {
	lock.Lock()
	defer lock.Unlock()
	delete(playerMapper, guildID)
}
