package config

import (
	"flag"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/ini.v1"
)

type Config struct {
	Token                  string
	GuildID                string
	Activity               discordgo.Activity
	Status                 discordgo.Status
	GeneralCommandsEnabled bool
	MusicCommandsEnable    bool
}

func Read() (*Config, error) {
	var filePath string
	flag.StringVar(&filePath, "config", "config.ini", "config file path")
	flag.Parse()

	data, err := ini.Load(filePath)
	if err != nil {
		return nil, err
	}

	var cfg Config = Config{
		Token:   data.Section("bot").Key("token").String(),
		GuildID: data.Section("bot").Key("guild_id").String(),
		Activity: discordgo.Activity{
			Name: data.Section("activity").Key("name").String(),
			Type: discordgo.ActivityType(data.Section("activity").Key("type").MustInt()),
			URL:  data.Section("activity").Key("url").String(),
		},
		Status:                 discordgo.Status(data.Section("activity").Key("status").String()),
		GeneralCommandsEnabled: data.Section("general_commands").Key("enabled").MustBool(),
		MusicCommandsEnable:    data.Section("music_commands").Key("enabled").MustBool(),
	}
	return &cfg, nil
}
