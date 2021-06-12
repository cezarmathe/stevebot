package steve

import (
	"context"
)

// Steve is a component that handles the communication between stevebot and the
// Minecraft Server (via RCON).
type Steve interface {
	// Start starts steve.
	//
	// This function must return an error if any errors are encountered during
	// the start process.
	Start(context.Context) error

	// SubmitCommand submits a command and returns after the command is started.
	//
	// This function starts a goroutine that actually handles the command and
	// waits for it to be "ready" - that means it got access to an rcon client
	// and it is ready to send the command.
	SubmitCommand(context.Context, []string) SteveCommandOutput

	// getRconClient returns a rcon client.
	//
	// This function must:
	// * lock the mutex
	// * if the client is nil, attempt to get a new client
	getRconClient(ctx context.Context) (rconClient, error)
}

// NewSteve creates a new steve instance.
//
// This function must return an error if:
// * a steve instance has already been initialized
// * other errors are encountered when initializing steve
func NewSteve() error {
	return newSteve()
}

// Get returns a Steve instance.
func Get() Steve {
	return steve
}
