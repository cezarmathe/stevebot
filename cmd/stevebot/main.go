package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cezarmathe/stevebot/internal/bot"
	"github.com/cezarmathe/stevebot/internal/steve"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const (
	// fixme 23/05/2021: add configurable graceful shutdown timeout
	//
	// This timeout is sensible to the time it takes for discord to gracefully
	// disconnect.
	gracefulShutdownTimeout = time.Second * 2
)

var (
	// Version is the build version of stevebot.
	Version string
)

var (
	log *zap.SugaredLogger
)

func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to initialize logger")
	}
	log = logger.Sugar().Named("main")
}

func main() {
	log.Infow("hello, this is stevebot", "version", Version)

	err := godotenv.Load()
	if err != nil {
		log.Info("not loading variables from .env")
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)

	err = steve.NewSteve()
	if err != nil {
		log.Errorw("failed to create a new steve", "err", err)
		cancel()
		shutdown(wg)
	}
	err = bot.NewBot()
	if err != nil {
		log.Errorw("failed to create a new bot", "err", err)
		cancel()
		shutdown(wg)
	}

	err = steve.Get().Start(ctx)
	if err != nil {
		log.Errorw("failed to start steve", "err", err)
		cancel()
		shutdown(wg)
	}
	err = bot.Get().Start(ctx, wg)
	if err != nil {
		log.Errorw("failed to start bot", "err", err)
		cancel()
		shutdown(wg)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	<-sig
	cancel()

	shutdown(wg)
}

func shutdown(wg *sync.WaitGroup) {
	done := make(chan byte, 1)
	go func() {
		wg.Wait()
		done <- 0x0
	}()

	timeout, cancelTimeout := context.WithTimeout(
		context.Background(),
		gracefulShutdownTimeout)

	select {
	case <-done:
		cancelTimeout()
		log.Info("have a nice day")
		os.Exit(0)
	case <-timeout.Done():
		cancelTimeout()
		log.Errorw("graceful shutdown timed out after, shutting down now",
			"timeout",
			gracefulShutdownTimeout.String())
		os.Exit(1)
	}
}
