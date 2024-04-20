package bot

import (
	"log"

	"github.com/foxeiz/namelessgo/config"
	"github.com/foxeiz/namelessgo/src/commands"
	"github.com/foxeiz/namelessgo/src/commands/general_commands"
	"github.com/foxeiz/namelessgo/src/commands/music_commands"

	"github.com/bwmarrin/discordgo"
)

var botSession *discordgo.Session
var botConfig *config.Config
var botCommand commands.CommandType

func init() {
	botCommand = commands.CommandType{
		Handlers: make(map[string]func(
			session *discordgo.Session,
			interaction *discordgo.InteractionCreate,
			options commands.CommandOptionsType,
		)),
		Commands: []*discordgo.ApplicationCommand{},
	}

	var err error
	botConfig, err = config.Read()
	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
		return
	}
}

func init() {
	if botConfig.GeneralCommandsEnabled {
		log.Println("Adding general commands")
		general_commands.Add(&botCommand)
	}

	if botConfig.MusicCommandsEnable {
		log.Println("Adding music commands")
		music_commands.Add(&botCommand)
	}
}

func init() {
	var err error
	botSession, err = discordgo.New("Bot " + botConfig.Token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
		return
	}

	botSession.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if interaction.Type != discordgo.InteractionApplicationCommand {
			return
		}
		if c, ok := botCommand.Handlers[interaction.ApplicationCommandData().Name]; ok {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})

			options := interaction.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			c(session, interaction, optionMap)
		}
	})

	botSession.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		s.UpdateStatusComplex(discordgo.UpdateStatusData{
			Activities: []*discordgo.Activity{&botConfig.Activity},
			Status:     string(botConfig.Status),
		})
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
}

func Start() {
	err := botSession.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
		return
	}

	log.Println("Syncing commands... Global commands may take up to 1 hour to register.")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(botCommand.Commands))
	for i, v := range botCommand.Commands {
		cmd, err := botSession.ApplicationCommandCreate(botSession.State.User.ID, botConfig.GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")
}

func Close() {
	// Update status to offline
	log.Println("Update bot status to offline.")
	botSession.UpdateStatusComplex(discordgo.UpdateStatusData{
		Status: string(discordgo.StatusOffline),
	})

	// Gracefully close the session
	err := botSession.Close()
	if err != nil {
		log.Fatalf("Cannot close the session: %v", err)
		return
	}
	log.Println("Bot is now closed.")
}
