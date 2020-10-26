package main

import (
	"github.com/cezarmathe/stevebot/internal/bot"
	"github.com/cezarmathe/stevebot/internal/steve"
	"github.com/cezarmathe/stevebot/internal/sys"

	hclog "github.com/hashicorp/go-hclog"
)

func main() {

	hclog.Default().Info("Starting stevebot")

	componentManager := sys.NewComponentManager()

	componentManager.RegisterComponent(new(bot.DiscordBotComponent))
	componentManager.RegisterComponent(new(steve.RCONComponent))

	componentManager.Start()

	hclog.Default().Info("Shutdown complete")
}
