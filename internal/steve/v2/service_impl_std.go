package stevev2i

import (
	"context"
	"errors"
	"strings"

	"github.com/gorcon/rcon"
	"go.uber.org/zap"
)

var (
	ErrCommandNotAllowed = errors.New("command is not allowed")
)

type StandardServiceConfig struct {
	AllowedCommands []string `env:"ALLOWED_COMMANDS"`
}

type StandardService struct {
	config *StandardServiceConfig
	logger *zap.Logger

	conn       *rcon.Conn                              // rcon connection
	dlqHandler func(cmd string, out string, err error) // dead letter queue handler
}

// Create a new standard steve v2 service.
func NewStandard(config *StandardServiceConfig, logger *zap.Logger, conn *rcon.Conn) StandardService {
	return StandardService{
		config: config,
		logger: logger,

		conn: conn,
		dlqHandler: func(cmd, out string, err error) {
			logger.Warn("dead letter queue", zap.String("out", out), zap.Error(err))
		},
	}
}

var (
	_ SteveV2 = (*StandardService)(nil)
)

func (svc *StandardService) Execute(ctx context.Context, cmd string) (string, error) {
	svc.logger.Debug("execute", zap.Any("ctx", ctx), zap.String("cmd", cmd))
	allowed := false
	for _, c := range svc.config.AllowedCommands {
		if strings.HasPrefix(cmd, c) {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", ErrCommandNotAllowed
	}
	type data struct {
		out string
		err error
	}
	ch := make(chan data, 1)
	go func() {
		defer close(ch)
		out, err := svc.conn.Execute(cmd)
		if ctx.Err() == nil {
			ch <- data{out, err}
		} else {
			svc.dlqHandler(cmd, out, err)
		}
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case val, ok := <-ch:
		if !ok {
			panic("execute channel closed unexpectedly")
		}
		return val.out, val.err
	}
}
