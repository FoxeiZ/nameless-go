package music_commands

var playerMapper = make(map[string]*Player)

func GetPlayer(guildID string) *Player {
	return playerMapper[guildID]
}

func AddPlayer(guildID string, player *Player) {
	playerMapper[guildID] = player
}

func RemovePlayer(guildID string) {
	delete(playerMapper, guildID)
}
