package steve

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

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

	var ok bool
	var shouldExit bool = false

	rconHost, ok = os.LookupEnv(rconHostKey)
	if !ok {
		log.Warnw("missing environment variable", "name", rconHostKey)
		shouldExit = true
	}

	var rconPortTmp string
	rconPortTmp, ok = os.LookupEnv(rconPortKey)
	if !ok {
		log.Warnw("missing environment variable", "name", rconPortKey)
		shouldExit = true
	}
	rconPort, err = strconv.Atoi(rconPortTmp)
	if err != nil {
		log.Warnw("environment variable is not an integer",
			"name", rconPortKey,
			"value", rconPortTmp)
		shouldExit = true
	}

	rconPassword, ok = os.LookupEnv(rconPasswordKey)
	if !ok {
		log.Warnw("missing environment variable", "name", rconPasswordKey)
		shouldExit = true
	}

	allowedCommandsStr, ok := os.LookupEnv(allowedCommandsKey)
	if !ok {
		allowedCommands = make([]string, 0)
	} else {
		allowedCommands = strings.Split(allowedCommandsStr, ",")
	}

	forbiddenCommandsStr, ok := os.LookupEnv(forbiddenCommandsKey)
	if !ok {
		forbiddenCommands = make([]string, 0)
	} else {
		forbiddenCommands = strings.Split(forbiddenCommandsStr, ",")
	}

	// allowed commands have a higher priority than forbidden commands
	if len(allowedCommands) > 0 {
		commandFilter = func(command string) error {
			for _, allowedCommand := range allowedCommands {
				if command == allowedCommand {
					return nil
				}
			}
			return errors.New("command not allowed")
		}
	} else if len(forbiddenCommands) > 0 {
		commandFilter = func(command string) error {
			for _, forbiddenCommand := range forbiddenCommands {
				if command == forbiddenCommand {
					return nil
				}
			}
			return errors.New("forbidden command")
		}
	} else {
		commandFilter = func(_ string) error {
			return nil
		}
	}

	if shouldExit {
		log.Fatal("cannot continue")
	}
}
