package general_commands

import (
	"github.com/foxeiz/namelessgo/src/commands"
)

func Add(commandHandlers *commands.CommandType) {
	commandHandlers.Handlers["ping"] = Ping
	commandHandlers.Commands = append(commandHandlers.Commands, PingCommand)
}
