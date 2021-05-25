package steve

import (
	"fmt"
	"os"
	"strconv"

	"github.com/cezarmathe/stevebot/internal/common"
	"go.uber.org/zap"
)

var (
	rconHostKey     = fmt.Sprintf("%s_RCON_HOST", common.EnvVarKeyPrefix)
	rconPortKey     = fmt.Sprintf("%s_RCON_PORT", common.EnvVarKeyPrefix)
	rconPasswordKey = fmt.Sprintf("%s_RCON_PASSWORD", common.EnvVarKeyPrefix)
)

var (
	log *zap.SugaredLogger

	rconHost     string
	rconPort     int
	rconPassword string
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

	if shouldExit {
		log.Fatal("cannot continue")
	}
}
