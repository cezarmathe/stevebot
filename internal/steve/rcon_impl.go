package steve

import (
	"context"
	"fmt"

	"github.com/bearbin/mcgorcon"
)

type rconClientImpl struct {
	inner *mcgorcon.Client
}

func newRconClientImpl(ctx context.Context) (rconClient, error) {
	// data returned after dialing a minecraft server
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
		// create a goroutine that actually waits for the operation to finish
		// this happens in order to log a potential error (or it's absence)
		go func() {
			out := <-outChan
			if out.err != nil {
				log.Warnf("rcon client: new: %w", out.err)
			} else {
				log.Warn("rcon client: new: ok after context canceled")
			}
		}()
		log.Warn("rcon client: new: context canceled")
		return nil, fmt.Errorf("timed out waiting for a new rcon client")
	case out := <-outChan:
		if out.err != nil {
			return nil, out.err
		}
		return &rconClientImpl{out.client}, nil
	}
}

func (c *rconClientImpl) SendCommand(ctx context.Context,
	input rconCommandInput) rconCommandOutput {

	outChan := make(chan rconCommandOutput, 1)

	go func() {
		out, err := c.inner.SendCommand(input.Command())
		outChan <- newRconCommandOutput(out, err)
	}()

	select {
	case <-ctx.Done():
		// create a goroutine that actually waits for the operation to finish
		// this happens in order to log a potential error (or it's absence)
		go func() {
			out := <-outChan
			if out.Success() {
				log.Warnw(
					"rcon client: send command: ok after context canceled",
					"out", out.Out())
			} else {
				log.Warnf("rcon client: send command: %w", out)
			}
		}()
		log.Warn("rcon client: send command: context canceled")
		err := fmt.Errorf("timed out waiting to send a command")
		return newRconCommandOutput("", err)
	case out := <-outChan:
		if !out.Success() {
			log.Warnf("rcon client: send command: %w", out)
		}
		err := fmt.Errorf("encountered an error while sending the command")
		return newRconCommandOutput("", err)
	}
}
