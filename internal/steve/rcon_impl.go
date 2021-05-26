package steve

import (
	"context"
	"fmt"

	"github.com/bearbin/mcgorcon"
)

// rconClientDummy is a dummy rcon client.
type rconClientDummy struct{}

func newDummyRconClient() rconClient {
	return new(rconClientDummy)
}

func (c *rconClientDummy) IsDummy() bool {
	return true
}

func (c *rconClientDummy) SendCommand(ctx context.Context,
	input RconCommandInput) RconCommandOutput {

	err := fmt.Errorf("stevebot is not connected to the server via rcon")
	return newRconCommandOutput("", err)
}

// rconClientImpl is an actual implementation of a rcon client.
type rconClientImpl struct {
	inner *mcgorcon.Client
}

func newRconClientImpl(ctx context.Context) (rconClient, error) {
	type out struct {
		client *mcgorcon.Client
		err    error
	}

	outChan := make(chan out, 1)
	go func() {
		client, err := mcgorcon.Dial(rconHost, rconPort, rconPassword)
		outChan <- out{&client, err}
	}()

	select {
	case <-ctx.Done():
		errMsg := "context canceled before rcon client could be created"
		return nil, fmt.Errorf(errMsg)
	case out := <-outChan:
		if out.err != nil {
			return nil, out.err
		}
		return &rconClientImpl{out.client}, nil
	}
}

func (c *rconClientImpl) IsDummy() bool {
	return false
}

func (c *rconClientImpl) SendCommand(ctx context.Context,
	input RconCommandInput) RconCommandOutput {

	outChan := make(chan RconCommandOutput, 1)

	go func() {
		out, err := c.inner.SendCommand(input.Command())
		outChan <- newRconCommandOutput(out, err)
	}()

	select {
	case <-ctx.Done():
		err := fmt.Errorf("server did not respond to the rcon command in time")
		return newRconCommandOutput("", err)
	case out := <-outChan:
		return out
	}
}
