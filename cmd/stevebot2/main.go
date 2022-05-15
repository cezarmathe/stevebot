package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/caarlos0/env/v6"
	botv2i "github.com/cezarmathe/stevebot/internal/bot/v2"
	stevev2i "github.com/cezarmathe/stevebot/internal/steve/v2"
	"github.com/gorcon/rcon"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func init() {
	jsonOutput := flag.Bool("json", false, "Enable JSON logging output.")
	debug := flag.Bool("debug", false, "Enable debug logging.")
	flag.Parse()

	var err error

	lc := zap.NewProductionConfig()
	if *jsonOutput {
		lc.Encoding = "json"
	} else {
		lc.Encoding = "console"
	}
	if *debug {
		lc.Level.SetLevel(zap.DebugLevel)
	} else {
		lc.Level.SetLevel(zap.InfoLevel)
	}
	logger, err = lc.Build()
	if err != nil {
		panic(err)
	}
}

type Config struct {
	DiscordToken string                         `env:"DISCORD_TOKEN"`
	RconAddress  string                         `env:"RCON_ADDRESS"`
	RconPassword string                         `env:"RCON_PASSWORD"`
	Bot          botv2i.Config                  `envPrefix:"BOT_"`
	Steve        stevev2i.StandardServiceConfig `envPrefix:"STEVE_"`
}

func main() {
	logger.Info("hello, this is stevebot2")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var mainConfig Config
	if err := env.Parse(&mainConfig); err != nil {
		logger.Panic("load main config", zap.Error(err))
	}
	logger.Debug("main config", zap.Any("value", mainConfig))

	dSess, err := discordgo.New(fmt.Sprintf("Bot %s", mainConfig.DiscordToken))
	if err != nil {
		logger.Panic("create discord session", zap.Error(err))
	}
	defer dSess.Close()

	rc, err := rcon.Dial(mainConfig.RconAddress, mainConfig.RconPassword)
	if err != nil {
		logger.Panic("dial rcon", zap.Error(err))
	}
	defer rc.Close()

	steve := stevev2i.NewStandard(&mainConfig.Steve, logger, rc)
	bot := botv2i.New(&mainConfig.Bot, logger, &steve)

	dSess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		bot.HandleCommand(ctx, s, m)
	})

	logger.Info("running")
	<-ctx.Done()
	logger.Info("shutting down")
	stop()

	if err := dSess.Close(); err != nil {
		logger.Error("close discord session", zap.Error(err))
	}
	if err := rc.Close(); err != nil {
		logger.Error("close rcon client", zap.Error(err))
	}

	logger.Info("bye bye")
}
