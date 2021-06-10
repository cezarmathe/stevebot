package steve

import (
	"context"
)

// rconClient is the interface between steve and the Minecraft Server (via
// RCON). An rcon client is NOT thread safe!
type rconClient interface {
	// SendCommand sends a command via RCON.
	SendCommand(context.Context, rconCommandInput) rconCommandOutput
}
