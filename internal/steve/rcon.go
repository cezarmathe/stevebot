package steve

import (
	"context"
)

// rconClient is the interface between steve and the Minecraft Server (via
// RCON).
type rconClient interface {
	// IsDummy returns whether this rcon client is a dummy rcon client or not.
	IsDummy() bool

	// SendCommand sends a command via RCON.
	SendCommand(context.Context, RconCommandInput) RconCommandOutput
}

// RconCommandInput is the input required to send a rcon command.
type RconCommandInput interface {
	// Command returns the actual command.
	Command() string
}

// RconCommandOutput is the output produced by sending a command via rcon.
type RconCommandOutput interface {
	error

	// Out returns the message sent back by the Minecraft Server via rcon.
	Out() string

	// Success returns whether the command completed successfully or not.
	Success() bool
}
