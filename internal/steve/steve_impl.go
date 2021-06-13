package steve

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	steve *steveImpl
)

var (
	getInitialRconClientTimeout = time.Second * 5
)

type steveImpl struct {
	clientLock sync.Locker
	client     rconClient
}

func newSteve() error {
	if steve != nil {
		return fmt.Errorf("steve: new: already created")
	}

	steve = new(steveImpl)

	// note 26/05/2021: mutex was an arbitraty choice, this could have been
	//                  a rwlock
	steve.clientLock = new(sync.Mutex)
	steve.client = nil

	// lock the mutex until the client is initialized
	// unlocking happens in Start() only if starting is successful
	steve.clientLock.Lock()

	return nil
}

func (s *steveImpl) Start(ctx context.Context) error {
	log.Info("hello, this is steve")
	// idea 23/05/2021: if address is domain, set periodic resolve
	//                  othwerise, pass ip

	ctx, cancel := context.WithTimeout(ctx, getInitialRconClientTimeout)
	client, err := newRconClientImpl(ctx)
	if err != nil {
		log.Warnf("steve: start: failed to get initial rcon client: %w", err)
	} else {
		// note 24/05/2021: it is expected for the mutex to be locked here, it was
		//                  locked by the newSteve func
		steve.client = client
	}
	cancel()

	// unlock the mutex that guards rconClient
	steve.clientLock.Unlock()

	return nil
}

func (s *steveImpl) getRconClient(ctx context.Context) (rconClient, error) {
	locked := make(chan struct{}, 1)

	// lock the rcon client
	// note 11/06/2021: the lock will be unlocked either by `handleCommand` or
	//                  by a goroutine that gets started if the context is
	//                  canceled
	go func() {
		s.clientLock.Lock()
		locked <- struct{}{}
	}()

	// wait for the lock or return on timeout
	select {
	case <-ctx.Done():
		// note 11/06/2021: create a goroutine that unlocks the client if the
		//                  context gets canceled
		go func() {
			<-locked
			s.clientLock.Unlock()
			log.Warn("steve: get rcon client: unlocked after context canceled")
		}()
		err := errors.New("timed out waiting for an available rcon client")
		return nil, err
	case <-locked:
		break
	}

	// if there is a client, return it
	if s.client != nil {
		return s.client, nil
	}

	// otherwise, create a new rcon client
	client, err := newRconClientImpl(ctx)
	s.client = client
	if err != nil {
		log.Warnf("steve: get rcon client: %w", err)
		errMsg := "failed to get an rcon client for the mineraft server"
		return nil, errors.New(errMsg)
	}
	return client, nil
}

func (s *steveImpl) SubmitCommand(ctx context.Context,
	command []string) SteveCommandOutput {

	// if this command does not pass the filter, return an error
	if err := commandFilter(command[0]); err != nil {
		return newSteveCommandOutput(err)
	}

	// create a channel for getting the steve output from the goroutine
	outChan := make(chan SteveCommandOutput)

	// start the command handler
	go func() {
		// get an rcon client
		client, err := s.getRconClient(ctx)
		if err != nil {
			outChan <- newSteveCommandOutput(err)
			return
		}

		// create the steve command input and output
		steveIn, steveOut := newSteveCommand(command)
		// return the steve command output to SubmitCommand so that it can
		// return it to bot
		outChan <- steveOut

		// build rcon command input
		rconIn := newRconCommandInput(strings.Join(steveIn.Command(), " "))

		// sent the command and get it's output
		rconOut := client.SendCommand(ctx, rconIn)

		// note 11/06/2021: if the command is unsuccessful, reset the rcon client
		//                  this should ideally happen only when steve can't reach
		//                  the minecraft server
		if !rconOut.Success() {
			s.client = nil
		}

		// unlock client mutex previously locked by getRconClient
		s.clientLock.Unlock()

		// send result
		steveIn.inChan() <- rconOut
	}()

	return <-outChan
}
