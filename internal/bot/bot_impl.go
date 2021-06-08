package bot

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cezarmathe/stevebot/internal/steve"
)

const (
	COMMAND_TIMEOUT = time.Second * 10
)

var (
	bot *botImpl
)

type botImpl struct {
	ctx context.Context
	wg  *sync.WaitGroup

	mutex *sync.Mutex
	sess  *discordgo.Session
}

func newBot(ctx context.Context, wg *sync.WaitGroup) error {
	if bot != nil {
		return fmt.Errorf("bot has already been created")
	}

	bot = new(botImpl)

	bot.ctx = ctx
	bot.wg = wg

	bot.mutex = new(sync.Mutex)
	bot.sess = nil

	// lock mutex that gives access to the discord session
	// the mutex will only be unlocked if the session is initialized
	bot.mutex.Lock()

	return nil
}

func (b *botImpl) Start() error {
	if b == nil {
		return fmt.Errorf("cannot start bot: uninitialized")
	}

	dg, err := discordgo.New(fmt.Sprintf("Bot %s", discordToken))
	if err != nil {
		return err
	}

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		b.handle(s, m)
	})
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	// if the bot can't open a connection, return right away
	err = dg.Open()
	if err != nil {
		return err
	}
	b.sess = dg
	b.mutex.Unlock()

	b.wg.Add(1)
	go b.gracefulDisconnect()

	return nil
}

func (b *botImpl) handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	// do not process messages sent by the bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !commandStartRegex.Match([]byte(m.Content)) {
		return
	}

	// get command words
	command := strings.Fields(strings.TrimPrefix(m.Content, commandPrefix))

	ctx, cancel := context.WithTimeout(b.ctx, COMMAND_TIMEOUT)

	// todo 24/05/2021: first send a message that says "working" and then update
	// it with the outcome of running the command
	// todo 26/05/2021: get bot update status from out chan
	done := make(chan struct{}, 1)
	go func() {
		steveOut := steve.Get().SubmitCommand(ctx, command)
		if !steveOut.Success() {
			s.ChannelMessageSend(m.ChannelID, steveOut.Error())
			done <- struct{}{}
			cancel()
			return
		}

		rconOut := <-steveOut.OutChan()
		if !rconOut.Success() {
			s.ChannelMessageSend(m.ChannelID, rconOut.Error())
		} else {
			s.ChannelMessageSend(m.ChannelID, rconOut.Out())
		}
		done <- struct{}{}
		cancel()
	}()

	select {
	case <-b.ctx.Done():
		cancel()
		return
	case <-done:
		return
	}
}

// UpdateStatus updates the status of the bot.
func (b *botImpl) UpdateStatus(status string) error {
	b.mutex.Lock()

	errChan := make(chan error, 1)
	go func() {
		err := b.sess.UpdateStatus(0, status)
		errChan <- err
	}()

	select {
	case err := <-errChan:
		b.mutex.Unlock()
		return err
	case <-b.ctx.Done():
		log.Warn("shutting down, session mutex will remain locked")
		return fmt.Errorf("context canceled before UpdateStatus could finish")
	}
}

func (b *botImpl) gracefulDisconnect() {
	locked := make(chan struct{}, 1)

	<-b.ctx.Done()

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

	b.wg.Done()
}

func (b *botImpl) handleCommand(command []string) {}
