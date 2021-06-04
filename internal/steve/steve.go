package steve

import (
	"context"
	"sync"
)

// Steve is a component that handles the communication between stevebot and the
// Minecraft Server (via RCON).
type Steve interface {
	// Start starts steve.
	//
	// This function must return an error if any errors are encountered during
	// the start process.
	Start() error

	// SubmitCommand submits a command and returns after the command is started.
	SubmitCommand(context.Context, []string) SteveCommandOutput

	// watchServer is a goroutine that watches the rcon server.
	//
	// The function must
	// * periodically check whether the bot can dial the rcon server
	// * answer to requests to skip an operation
	// * answer to requests to run an operation now
	// * update the client to a dummy if the minecraft server cannot be dialed
	// * send a request to bot to update it's status if steve is not connected
	//   to the server via rcon
	watchServer()

	// skipWatchServer makes the watchServer goroutine skip an operation.
	skipWatchServer()

	// runWatchServer schedules a watch server operation now.
	runWatchServer()

	// updateRconClient updates steve's current rcon client.
	//
	// This is used for switching between a real rcon client and a dummy rcon
	// client.
	//
	// If a non-nil locker is passed, it must be locked.
	updateRconClient(rconClient, sync.Locker) error

	// handleCommand handles a single command.
	//
	// This function must:
	// * get the rcon client
	// * forward the command to the rcon client
	// * schedule a run watch server operation if running the command fails
	handleCommand(context.Context, SteveCommandInput)

	// getRconClient retrieves the rcon client.
	//
	// This function expects the client lock to be acquired!
	//
	// This function must:
	// * return the current rcon client, if it is not a dummy
	// * attempt to obtain a real rcon client, if the current rcon client is a
	//   dummy
	// * skip a watch server operation while attempting to get a new real rcon
	//   client
	// * replace the dummy rcon client if a real rcon client could be obtained
	// * return a dummy rcon client if getting a rcon client fails
	getRconClient(ctx context.Context) (rconClient, error)
}

// NewSteve creates a new steve instance.
//
// This function must return an error if:
// * a steve instance has already been initialized
// * other errors are encountered when initializing steve
func NewSteve(ctx context.Context, wg *sync.WaitGroup) error {
	return newSteve(ctx, wg)
}

// Get returns a Steve instance.
func Get() Steve {
	return steve
}
