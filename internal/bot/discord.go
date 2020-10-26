package bot

import (
	"os"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/cezarmathe/stevebot/internal/steve"
	"github.com/cezarmathe/stevebot/internal/steve/stevecmd"
	"github.com/hashicorp/go-hclog"
)

var (
	// the token used by this bot
	botToken string = "abc"

	// the prefix for commands
	commandPrefix string

	// logger
	log hclog.Logger
)

var (
	// the discord bot component
	discordBot *DiscordBotComponent
)

func init() {
	log = hclog.Default().Named("bot")

	ok := true

	botToken = os.Getenv("STEVEBOT_TOKEN")
	if botToken == "" {
		ok = false
		log.Error("Missing STEVEBOT_TOKEN environment variable")
	}

	commandPrefix = os.Getenv("STEVEBOT_COMMAND_PREFIX")
	if commandPrefix == "" {
		ok = false
		log.Error("Missing STEVEBOT_COMMAND_PREFIX environment variable")
	}

	if !ok {
		os.Exit(1)
	}
}

// DiscordBotComponent is the Discord bot component.
type DiscordBotComponent struct {
	interrupt chan byte
	wg        *sync.WaitGroup
}

// Start starts the discord bot component.
func (d *DiscordBotComponent) Start(interrupt chan byte, wg *sync.WaitGroup) <-chan bool {
	d.interrupt = interrupt
	d.wg = wg
	ready := make(chan bool, 1)

	d.wg.Add(1)
	go d.run(ready)

	discordBot = d
	return ready
}

// Name returns the name of the component.
func (d *DiscordBotComponent) Name() string {
	return "discordbot"
}

func (d *DiscordBotComponent) run(ready chan<- bool) {
	defer d.wg.Done()

	log.Info("Starting the Discord bot..")

	// create a new discord session using the provided bot token
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Error("Error creating Discord session", "err", err)
		ready <- false
		return
	}

	// register the messageHandler func as a callback for MessageCreate events
	dg.AddHandler(messageHandler)

	// we only care about receiving message events
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	// open a websocket connection to Discord and begin listening
	err = dg.Open()
	if err != nil {
		log.Error("Error opening connection to Discord", "err", err)
		ready <- false
		return
	}

	log.Info("Discord bot is now running")
	ready <- true

	_ = <-d.interrupt
	log.Info("Stopping Discord bot component")
	dg.Close()
	d.interrupt <- 0
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore all messages
	if !strings.HasPrefix(m.Content, commandPrefix) {
		return
	}
	if len(m.Content) == 1 {
		return
	}

	command := strings.Split(strings.TrimPrefix(m.Content, commandPrefix), " ")

	if command[0] == "help" {
		if len(command) == 1 {
			s.ChannelMessageSend(m.ChannelID, "Available commands: "+
				"ban, banlist, kick, list, pardon, say, seed, stop, whitelist. "+
				"Use help <command_name> for more help.")
		} else {
			var help string
			switch command[1] {
			case "ban":
				help = stevecmd.NewBan("").Help()
				break
			case "banlist":
				help = new(stevecmd.Banlist).Help()
				break
			case "kick":
				help = stevecmd.NewKick("").Help()
				break
			case "list":
				help = new(stevecmd.List).Help()
				break
			case "pardon":
				help = stevecmd.NewPardon("").Help()
				break
			case "say":
				help = stevecmd.NewSay("").Help()
				break
			case "seed":
				help = new(stevecmd.Seed).Help()
				break
			case "stop":
				help = new(stevecmd.Stop).Help()
				break
			case "whitelist":
				help = stevecmd.NewWhitelistList().Help()
				break
			default:
				s.ChannelMessageSend(m.ChannelID, "Available commands: "+
					"ban, banlist, kick, list, pardon, say, seed, stop, whitelist. "+
					"Use help <command_name> for more help")
			}
			s.ChannelMessageSend(m.ChannelID, help)
		}
		return
	}

	if command[0] == "ban" {
		var cmd steve.RCONCommand
		if len(command) < 2 {
			s.ChannelMessageSend(m.ChannelID, new(stevecmd.Ban).Help())
			return
		}
		cmd = stevecmd.NewBan(command[1])
		result := <-steve.SubmitRCONCommand(cmd)
		if result.Err != nil {
			s.ChannelMessageSend(m.ChannelID, "ecnountered an error")
			log.Warn("encountered an error when running a rcon command", "cmd", cmd, "err", result.Err)
		} else {
			s.ChannelMessageSend(m.ChannelID, result.Out)
		}
		return
	}

	if command[0] == "banlist" {
		result := <-steve.SubmitRCONCommand(new(stevecmd.Banlist))
		if result.Err != nil {
			s.ChannelMessageSend(m.ChannelID, "ecnountered an error")
			log.Warn("encountered an error when running a rcon command", "cmd", "banlist", "err", result.Err)
		} else {
			s.ChannelMessageSend(m.ChannelID, result.Out)
		}
		return
	}

	if command[0] == "kick" {
		var cmd steve.RCONCommand
		if len(command) < 2 {
			s.ChannelMessageSend(m.ChannelID, new(stevecmd.Kick).Help())
			return
		}
		cmd = stevecmd.NewKick(command[1])
		result := <-steve.SubmitRCONCommand(cmd)
		if result.Err != nil {
			s.ChannelMessageSend(m.ChannelID, "ecnountered an error")
			log.Warn("encountered an error when running a rcon command", "cmd", cmd, "err", result.Err)
		} else {
			s.ChannelMessageSend(m.ChannelID, result.Out)
		}
		return
	}

	if command[0] == "list" {
		result := <-steve.SubmitRCONCommand(new(stevecmd.List))
		if result.Err != nil {
			s.ChannelMessageSend(m.ChannelID, "ecnountered an error")
			log.Warn("encountered an error when running a rcon command", "cmd", "list", "err", result.Err)
		} else {
			s.ChannelMessageSend(m.ChannelID, result.Out)
		}
		return
	}

	if command[0] == "pardon" {
		var cmd steve.RCONCommand
		if len(command) < 2 {
			s.ChannelMessageSend(m.ChannelID, new(stevecmd.Pardon).Help())
			return
		}
		cmd = stevecmd.NewPardon(command[1])
		result := <-steve.SubmitRCONCommand(cmd)
		if result.Err != nil {
			s.ChannelMessageSend(m.ChannelID, "ecnountered an error")
			log.Warn("encountered an error when running a rcon command", "cmd", cmd, "err", result.Err)
		} else {
			s.ChannelMessageSend(m.ChannelID, result.Out)
		}
		return
	}

	if command[0] == "say" {
		var cmd steve.RCONCommand
		if len(command) < 2 {
			s.ChannelMessageSend(m.ChannelID, new(stevecmd.Say).Help())
			return
		}
		cmd = stevecmd.NewSay(strings.Join(command[1:], " "))
		result := <-steve.SubmitRCONCommand(cmd)
		if result.Err != nil {
			s.ChannelMessageSend(m.ChannelID, "ecnountered an error")
			log.Warn("encountered an error when running a rcon command", "cmd", cmd, "err", result.Err)
		} else {
			s.ChannelMessageSend(m.ChannelID, result.Out)
		}
		return
	}

	if command[0] == "seed" {
		result := <-steve.SubmitRCONCommand(new(stevecmd.Seed))
		if result.Err != nil {
			s.ChannelMessageSend(m.ChannelID, "ecnountered an error")
			log.Warn("encountered an error when running a rcon command", "cmd", "seed", "err", result.Err)
		} else {
			s.ChannelMessageSend(m.ChannelID, result.Out)
		}
		return
	}

	if command[0] == "whitelist" {
		var cmd steve.RCONCommand
		switch command[1] {
		case "list":
			cmd = stevecmd.NewWhitelistList()
			break
		case "on":
			cmd = stevecmd.NewWhitelistOn()
			break
		case "off":
			cmd = stevecmd.NewWhitelistOff()
			break
		case "add":
			if len(command) < 3 {
				s.ChannelMessageSend(m.ChannelID, new(stevecmd.Whitelist).Help())
				return
			}
			cmd = stevecmd.NewWhitelistAdd(command[2])
			break
		default:
			s.ChannelMessageSend(m.ChannelID, new(stevecmd.Whitelist).Help())
			return
		}
		result := <-steve.SubmitRCONCommand(cmd)
		if result.Err != nil {
			s.ChannelMessageSend(m.ChannelID, "ecnountered an error")
			log.Warn("encountered an error when running a rcon command", "cmd", cmd, "err", result.Err)
		} else {
			s.ChannelMessageSend(m.ChannelID, result.Out)
		}
		return
	}
}
