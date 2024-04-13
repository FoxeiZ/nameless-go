package music_commands

import p "github.com/foxeiz/namelessgo/src/commands/music_commands/player"

var playerMapper = make(map[string]*p.Player)

func GetPlayer(guildID string) *p.Player {
	return playerMapper[guildID]
}

func AddPlayer(guildID string, player *p.Player) {
	playerMapper[guildID] = player
}

func RemovePlayer(guildID string) {
	delete(playerMapper, guildID)
}
