package steve

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	// note 25/05/2021: is one minute okay?
	// duration of the watch server watchdog
	watchServerWatchdogDuration = time.Minute

	// note 25/05/2021: are five seconds okay?
	// duration waited by watchServer until considering a server dial failed
	watchServerDialTimeout = time.Second * 5
)

var (
	steve *steveImpl
)

type steveImpl struct {
	ctx context.Context
	wg  *sync.WaitGroup

	clientLock sync.Locker
	client     rconClient

	// channel for skipping a scheduled watch server operation
	// this just resets the watchdog
	skipWatchServerChan chan struct{}

	// channel for scheduling a watch server operation now
	// this bypasses the watchdog
	runWatchServerChan chan struct{}
}

func newSteve(ctx context.Context, wg *sync.WaitGroup) error {
	if steve != nil {
		return fmt.Errorf("steve has already been created")
	}

	steve = new(steveImpl)

	steve.ctx = ctx
	steve.wg = wg

	// note 26/05/2021: mutex was an arbitraty choice, this could have been
	//                  a rwlock
	steve.clientLock = new(sync.Mutex)
	steve.client = nil

	// skip channel is unbuffered
	// this avoids blocking the goroutine that calls skipWatchServer without
	// affecting the watchServer goroutine - it will loop a few times until all
	// channel values are exhausted
	steve.skipWatchServerChan = make(chan struct{})

	// run watch server is unbuffered
	// this avoids blocking the goroutine that calls runWatchServer without
	// affecting the watchServer goroutine - while the operation is running, it
	// spawns another goroutine that exhausts all incoming runWatchServer
	// requests
	steve.runWatchServerChan = make(chan struct{})

	// lock the mutex until the client is initialized
	// unlocking happens in Start() only if starting is successful
	steve.clientLock.Lock()

	return nil
}

func (s *steveImpl) Start() error {
	// idea 23/05/2021: if address is domain, set periodic resolve
	//                  othwerise, pass ip

	client, err := newRconClientImpl(s.ctx)
	if err != nil {
		log.Warnw("first client initialization failed", "err", err)
		client = newDummyRconClient()
	}

	// note 24/05/2021: it is expected for the mutex to be locked here, it was
	//                  locked by the newSteve func
	steve.client = client

	// start the watchServer goroutine
	s.wg.Add(1)
	go steve.watchServer()

	// unlock the mutex that guards rconClient
	steve.clientLock.Unlock()

	return nil
}

func (s *steveImpl) watchServer() {
	log.Debugw("starting", "who", "watchServer")

	// channel for returning errors from the goroutine that dials the rcon
	// server
	errChan := make(chan error, 1)

	for {
		// run watchdog
		select {
		case <-s.ctx.Done():
			log.Debugw("context canceled, shutting down", "who", "watchServer")
			s.wg.Done()
			return
		case <-s.skipWatchServerChan:
			log.Debugw("skipping an operation", "who", "watchServer")
			continue
		case <-s.runWatchServerChan:
			log.Debugw("running an operation now", "who", "watchServer")
			// note 26/05/2021: this break is intentionally used to signal that
			//                  we want to break from select, not from the loop
			break
		case <-time.After(watchServerWatchdogDuration):
			log.Debugw("watchdog timed out, running an operation now",
				"who", "watchServer")
			// note 26/05/2021: this break is intentionally used to signal that
			//                  we want to break from select, not from the loop
			break
		}

		// attempt to dial the server
		go func() {
			_, err := net.Dial("tcp", fmt.Sprintf("%s:%d", rconHost, rconPort))
			errChan <- err
		}()

		// exhaust incoming runWatchServer requests while the operation runs
		done := make(chan struct{}, 1)
		s.wg.Add(1)
		go func() {
			for {
				select {
				case <-s.ctx.Done():
					s.wg.Done()
					return
				case <-done:
					s.wg.Done()
					return
				case <-s.runWatchServerChan:
					log.Debugw("received a runWatchServer request, ignoring",
						"who", "watchServer")
					continue
				}
			}
		}()

		// wait for operation to finish
		select {
		case <-s.ctx.Done():
			s.wg.Done()
			log.Debugw("context canceled before the operation finished",
				"who", "watchServer")
			return
		case <-s.skipWatchServerChan:
			log.Debugw("skipping the operation that runs right now",
				"who", "watchServer")
			continue
		case <-time.After(watchServerDialTimeout):
			log.Warnw("failed to dial rcon server",
				"err", "timeout",
				"who", "watchServer")
			err := s.updateRconClient(newDummyRconClient(), nil)
			if err != nil {
				log.Warnw("failed to update rcon client",
					"err", err,
					"who", "watchServer")
			}
		case err := <-errChan:
			if err != nil {
				log.Warnw("failed to dial rcon server",
					"err", err,
					"who", "watchServer")
				err = s.updateRconClient(newDummyRconClient(), nil)
				if err != nil {
					log.Warnw("failed to update rcon client",
						"err", err,
						"who", "watchServer")
				}
			}
		}
		done <- struct{}{}
	}
}

func (s *steveImpl) updateRconClient(newClient rconClient,
	clientLock sync.Locker) error {

	lockChan := make(chan sync.Locker, 1)

	go func() {
		var lock sync.Locker
		if clientLock != nil {
			lock = clientLock
		} else {
			lock = s.clientLock
			lock.Lock()
		}
		lockChan <- lock
	}()

	select {
	case lock := <-lockChan:
		s.client = newClient
		lock.Unlock()
		return nil
	case <-s.ctx.Done():
		log.Warnw("rcon client mutex will not be unlocked")
		return fmt.Errorf("failed to update rcon client: %s",
			"context canceled before mutex could be unlocked")
	}
}

func (s *steveImpl) SubmitCommand(ctx context.Context,
	command []string) SteveCommandOutput {

	// if this command does not pass the filter, return an error
	if err := commandFilter(command[0]); err != nil {
		return newSteveCommandOutput(err)
	}

	locked := make(chan struct{}, 1)
	go func() {
		// note 26/05/2021: the mutex will be unlocked by handleCommand
		s.clientLock.Lock()
		locked <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		// fixme 01/06/2021: client lock will remain locked forever?
		go func() {
			<-locked
			s.clientLock.Unlock()
		}()
		return newSteveCommandOutput(fmt.Errorf("could not get rcon client"))
	case <-locked:
		in, out := newSteveCommand(command)
		go s.handleCommand(ctx, in)
		return out
	}
}

func (s *steveImpl) skipWatchServer() {
	s.skipWatchServerChan <- struct{}{}
}

func (s *steveImpl) runWatchServer() {
	s.runWatchServerChan <- struct{}{}
}

func (s *steveImpl) getRconClient(ctx context.Context) (rconClient, error) {
	client := s.client

	if client.IsDummy() {
		client, err := newRconClientImpl(ctx)
		if err != nil {
			return nil, err
		}
		err = s.updateRconClient(client, s.clientLock)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (s *steveImpl) handleCommand(ctx context.Context,
	input SteveCommandInput) {

	command := newRconCommandInput(strings.Join(input.Command(), " "))

	client, err := s.getRconClient(ctx)
	if err != nil {
		input.InChan() <- newRconCommandOutput("", err)
		s.runWatchServer()
		return
	}
	s.skipWatchServer()

	out := client.SendCommand(ctx, command)
	if !out.Success() {
		err = s.updateRconClient(newDummyRconClient(), s.clientLock)
		if err != nil {
			log.Warnw("failed to update rcon client to dummy", "err", err)
		}
	} else {
		// note 26/05/2021: the mutex was previously locked by SubmitCommand
		s.clientLock.Unlock()
	}
	input.InChan() <- out
}
