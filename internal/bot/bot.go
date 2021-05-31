package bot

import (
	"context"
	"sync"
)

// Bot is a component that handles the interaction with stevebot via Discord.
type Bot interface {
	// Start starts bot.
	//
	// This function must return an error if:
	// * the bot instance has not been initialized yet
	// * any other errors are encountered during the start process
	Start() error

	// gracefulDisconnect is a goroutine that gracefully disconects the bot from
	// Discord.
	gracefulDisconnect()

	// handleCommand handles a command received.
	handleCommand(command []string)
}

// NewBot creates a new bot instance.
//
// This function must return an error if:
// * a steve instance has already been initialized
// * other errors are encountered when initializing steve
func NewBot(ctx context.Context, wg *sync.WaitGroup) error {
	return newBot(ctx, wg)
}

// Get returns a Bot instance.
func Get() Bot {
	return bot
}
