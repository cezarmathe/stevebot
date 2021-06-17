package steve

import (
	"fmt"
	"os"

	"github.com/cezarmathe/stevebot/internal/common"
	"go.uber.org/zap"
)

var (
	rconHostKey          = fmt.Sprintf("%s_RCON_HOST", common.EnvVarKeyPrefix)
	rconPortKey          = fmt.Sprintf("%s_RCON_PORT", common.EnvVarKeyPrefix)
	rconPasswordKey      = fmt.Sprintf("%s_RCON_PASSWORD", common.EnvVarKeyPrefix)
	allowedCommandsKey   = fmt.Sprintf("%s_ALLOWED_COMMANDS", common.EnvVarKeyPrefix)
	forbiddenCommandsKey = fmt.Sprintf("%s_FORBIDDEN_COMMANDS", common.EnvVarKeyPrefix)
)

var (
	log *zap.SugaredLogger

	rconHost     string
	rconPort     int
	rconPassword string

	allowedCommands   []string
	forbiddenCommands []string
	commandFilter     func(string) error
)

func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to initialize logger")
		os.Exit(1)
	}
	log = logger.Sugar().Named("steve")
}
