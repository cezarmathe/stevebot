package botv2i

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	stevev2i "github.com/cezarmathe/stevebot/internal/steve/v2"
	"go.uber.org/zap"
)

type Config struct {
	CommandPrefix string `env:"COMMAND_PREFIX"`
}

type Service struct {
	config *Config
	logger *zap.Logger

	steve stevev2i.SteveV2
}

func New(config *Config, logger *zap.Logger, steve stevev2i.SteveV2) Service {
	return Service{
		config: config,
		logger: logger,

		steve: steve,
	}
}

func (svc *Service) HandleCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		svc.logger.Debug("message sent by bot user")
		return
	}
	command := m.Content
	if !strings.HasPrefix(command, svc.config.CommandPrefix) {
		svc.logger.Debug("message is not command")
		return
	}
	command = strings.TrimPrefix(command, svc.config.CommandPrefix)
	argv := strings.Fields(command)
	svc.logger.Debug("handle command", zap.Strings("argv", argv))
	fmsg, err := s.ChannelMessageSend(m.ChannelID, "Working on it..")
	if err != nil {
		svc.logger.Error("send feedback message", zap.Error(err))
	}
	out, err := svc.steve.Execute(ctx, strings.Join(argv, " "))
	if err != nil {
		_, err = s.ChannelMessageEdit(fmsg.ChannelID, fmsg.ID, fmt.Sprintf("Error: %s", err.Error()))
	} else {
		_, err = s.ChannelMessageEdit(fmsg.ChannelID, fmsg.ID, out)
	}
	if err != nil {
		svc.logger.Error("edit feedback message", zap.Error(err))
	}
}
