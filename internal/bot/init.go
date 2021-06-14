package bot

import (
	"fmt"
	"os"
	"regexp"

	"github.com/cezarmathe/stevebot/internal/common"
	"go.uber.org/zap"
)

var (
	discordTokenKey  = fmt.Sprintf("%s_DISCORD_TOKEN", common.EnvVarKeyPrefix)
	commandPrefixKey = fmt.Sprintf("%s_COMMAND_PREFIX", common.EnvVarKeyPrefix)
)

var (
	log *zap.SugaredLogger

	discordToken  string
	commandPrefix string

	// this regex is used to check whether a message starts like a command
	commandStartRegex *regexp.Regexp
)

func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to initialize logger")
		os.Exit(1)
	}
	log = logger.Sugar().Named("bot")

	commandStartRegexString := fmt.Sprintf("^%s\\w", commandPrefix)
	commandStartRegex, err = regexp.Compile(commandStartRegexString)
	if err != nil {
		log.Fatal("failed to compile command start regex", "err", err)
	}
}
