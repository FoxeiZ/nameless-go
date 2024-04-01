package general_commands

import (
	"github.com/foxeiz/namelessgo/src/types"
)

func Add(commandHandlers *types.CommandType) {
	commandHandlers.Handlers["ping"] = Ping
	commandHandlers.Commands = append(commandHandlers.Commands, PingCommand)
}
