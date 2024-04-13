package music_commands

import (
	"github.com/foxeiz/namelessgo/src/commands"
)

func Add(commandHandlers *commands.CommandType) {
	commandHandlers.Handlers["connect"] = Connect
	commandHandlers.Commands = append(commandHandlers.Commands, ConnectCommand)

	commandHandlers.Handlers["disconnect"] = Disconnect
	commandHandlers.Commands = append(commandHandlers.Commands, DisconnectCommand)

	commandHandlers.Handlers["play"] = Play
	commandHandlers.Commands = append(commandHandlers.Commands, PlayCommand)
}
