package stevev2i

import "context"

type SteveV2 interface {
	// Execute an RCON command.
	Execute(context.Context, string) (string, error)
}
