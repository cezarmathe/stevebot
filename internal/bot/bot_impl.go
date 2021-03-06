package bot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cezarmathe/stevebot/internal/steve"
)

const (
	// Command timeout - the amount of time bot will wait for before declaring
	// a command as failed.
	COMMAND_TIMEOUT = time.Second * 10

	// Emoji that signifies that a command is in progress.
	CMD_WIP_EMOJI = "🛠 "
	// Emoji that signifies that a command was successful.
	CMD_OK_EMOJI = "✅ "
	// Emoji that signifies that a command failed.
	CMD_ERR_EMOJI = "❌ "
)

var (
	bot *botImpl
)

type botImpl struct {
	mutex *sync.Mutex
	sess  *discordgo.Session
}

func newBot() error {
	if bot != nil {
		return fmt.Errorf("bot has already been created")
	}

	// load configuration from env

	shouldExit := false
	ok := true

	discordToken, ok = os.LookupEnv(discordTokenKey)
	if !ok {
		log.Warnf("new bot: missing environment variable: %s", discordTokenKey)
		shouldExit = true
	}

	commandPrefix, ok = os.LookupEnv(commandPrefixKey)
	if !ok {
		log.Warnf("new bot: missing environment variable: %s", commandPrefixKey)
		shouldExit = true
	}

	if shouldExit {
		return errors.New("new bot: failed to load configuration from env")
	}

	// create the bot object

	bot = new(botImpl)

	bot.mutex = new(sync.Mutex)
	bot.sess = nil

	// lock mutex that gives access to the discord session
	// the mutex will only be unlocked if the session is initialized
	bot.mutex.Lock()

	return nil
}

func (b *botImpl) Start(ctx context.Context, wg *sync.WaitGroup) error {
	log.Info("hello, this is bot")

	dg, err := discordgo.New(fmt.Sprintf("Bot %s", discordToken))
	if err != nil {
		return err
	}

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		b.handleCommand(ctx, s, m)
	})
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	// if the bot can't open a connection, return right away
	err = dg.Open()
	if err != nil {
		return err
	}
	b.sess = dg
	b.mutex.Unlock()

	wg.Add(1)
	go b.gracefulDisconnect(ctx, wg)

	return nil
}

func (b *botImpl) handleCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	// do not process messages sent by the bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	// if the message does not match with how a command starts, return
	if !commandStartRegex.Match([]byte(m.Content)) {
		return
	}

	// get command words
	command := strings.Fields(strings.TrimPrefix(m.Content, commandPrefix))

	ctx, cancel := context.WithTimeout(ctx, COMMAND_TIMEOUT)

	done := make(chan error, 1)
	go func() {
		// send a message to signify that the command was received and we're
		// working on it
		wipMsg := fmt.Sprintf("%s working on it.. (\"%s\" by %s)",
			CMD_WIP_EMOJI, strings.Join(command, " "), m.Author.Mention())
		msg, err := s.ChannelMessageSend(m.ChannelID, wipMsg)
		if err != nil {
			done <- err
			cancel()
			return
		}

		// submit the command
		// if this fails, update the WIP message and exit
		steveOut := steve.Get().SubmitCommand(ctx, command)
		if !steveOut.Success() {
			errMsg := fmt.Sprintf("%s  %s (\"%s\" by %s)",
				CMD_ERR_EMOJI,
				steveOut.Error(),
				strings.Join(command, " "),
				m.Author.Mention())
			_, err = s.ChannelMessageEdit(m.ChannelID, msg.ID, errMsg)
			if err != nil {
				log.Warnf("bot: handle command: %w", err)
				errMsg = fmt.Sprintf("%s  %s (\"%s\" by %s)",
					CMD_ERR_EMOJI,
					"can't update discord message with command output",
					strings.Join(command, " "),
					m.Author.Mention())
				_, _ = s.ChannelMessageEdit(m.ChannelID, msg.ID, errMsg)
			}
			done <- err
			cancel()
			return
		}

		// wait for the command to finish
		rconOut := <-steveOut.OutChan()

		// update wip message with command output
		if rconOut.Success() {
			okMsg := fmt.Sprintf("%s  %s (\"%s\" by %s)",
				CMD_OK_EMOJI,
				rconOut.Out(),
				strings.Join(command, " "),
				m.Author.Mention())
			_, err = s.ChannelMessageEdit(m.ChannelID, msg.ID, okMsg)
			if err != nil {
				log.Warnf("bot: handle command: %w", err)
				errMsg := fmt.Sprintf("%s  %s (\"%s\" by %s)",
					CMD_ERR_EMOJI,
					"can't update discord message with command output",
					strings.Join(command, " "),
					m.Author.Mention())
				_, _ = s.ChannelMessageEdit(m.ChannelID, msg.ID, errMsg)
			}
		} else {
			errMsg := fmt.Sprintf("%s  %s (\"%s\" by %s)",
				CMD_ERR_EMOJI,
				rconOut.Error(),
				strings.Join(command, " "),
				m.Author.Mention())
			_, err = s.ChannelMessageEdit(m.ChannelID, msg.ID, errMsg)
			if err != nil {
				log.Warnf("bot: handle command: %w", err)
				errMsg = fmt.Sprintf("%s  %s (\"%s\" by %s)",
					CMD_ERR_EMOJI,
					"can't update discord message with command output",
					strings.Join(command, " "),
					m.Author.Mention())
				_, _ = s.ChannelMessageEdit(m.ChannelID, msg.ID, errMsg)
			}
		}
		done <- err
		cancel()
	}()

	select {
	case <-ctx.Done():
		cancel()
	case err := <-done:
		if err != nil {
			log.Warnw("command failed",
				"err", err,
				"cmd", strings.Join(command, " "),
				"author", m.Author.String())
		}
	}
}

func (b *botImpl) gracefulDisconnect(ctx context.Context, wg *sync.WaitGroup) {
	locked := make(chan struct{}, 1)

	<-ctx.Done()

	go func() {
		b.mutex.Lock()
		locked <- struct{}{}
	}()

	select {
	case <-locked:
		b.sess.Close()
	case <-time.After(500 * time.Millisecond):
		log.Warnw("failed to gracefully close the connection to discord")
	}

	wg.Done()
}
