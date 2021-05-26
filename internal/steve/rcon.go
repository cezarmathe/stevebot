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
