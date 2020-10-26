package sys

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
)

var (
	log hclog.Logger
)

func init() {
	log = hclog.Default().Named("sys")
}

// Component is a component of the entire bot.
type Component interface {
	// Start starts the component.
	//
	// @interrupt: channel used for receiving interrupts for graceful shutdowns or when an
	//             unrecoverable error occurs and stevebot should do a graceful shutdown
	// @wg       : waitgroup to add self into
	//
	// This should spawn another goroutine that actually runs the component and returns a readiness
	// check channel.
	//
	// Returns a channel which will pass a bool that represents the readiness of the component.
	Start(interrupt chan byte, wg *sync.WaitGroup) <-chan bool

	// Name returns the name of the component.
	Name() string
}

type component struct {
	c         Component
	interrupt chan byte
	ready     <-chan bool
}

func newComponent(c Component) *component {
	comp := new(component)

	comp.c = c
	comp.interrupt = make(chan byte, 1)
	comp.ready = nil

	return comp
}

// ComponentManager is the component manager.
type ComponentManager struct {
	components []*component
	wg         *sync.WaitGroup
}

// NewComponentManager creates a new ComponentManager.
func NewComponentManager() *ComponentManager {
	c := new(ComponentManager)
	c.components = make([]*component, 0, 10)
	c.wg = new(sync.WaitGroup)
	return c
}

// RegisterComponent registers a new Component in the ComponentManager.
func (c *ComponentManager) RegisterComponent(comp Component) {
	c.components = append(c.components, newComponent(comp))
}

// Start starts the ComponentManager. This will block the (main) goroutine.
func (c *ComponentManager) Start() {
	interruptByComponent := make(chan string)

	// start all components
	for _, comp := range c.components {
		comp.ready = comp.c.Start(comp.interrupt, c.wg)

		// start a goroutine that checks the readiness of the component and then forwards
		// interrupts
		go func(_comp *component, _int chan<- string) {
			ready := <-_comp.ready
			if !ready {
				log.Error("Component not ready, sending interrupt signal", "component", _comp.c.Name())
				_int <- _comp.c.Name()
				return
			}
			log.Info("Component ready", "name", _comp.c.Name())
			ok := <-_comp.interrupt
			if ok == 1 {
				_int <- _comp.c.Name()
			}
		}(comp, interruptByComponent)
	}

	// wait for interrupts
	interruptBySignal := make(chan os.Signal, 1)
	signal.Notify(interruptBySignal, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	select {
	case name := <-interruptByComponent:
		log.Info("Received signal from a component, interrupting", "component", name)

		for _, comp := range c.components {
			if comp.c.Name() == name {
				continue
			}
			comp.interrupt <- 1
		}
	case sig := <-interruptBySignal:
		log.Info("Received signal, interrupting", "signal", sig)
		for _, comp := range c.components {
			comp.interrupt <- 1
		}
	}

	// attempt a graceful shutdown
	interruptDone := make(chan byte, 1)
	interruptTimeout := make(chan byte, 1)

	go func() {
		c.wg.Wait()
		interruptDone <- 1
	}()
	go func() {
		time.Sleep(3 * time.Second)
		interruptTimeout <- 1
	}()

	select {
	case _ = <-interruptDone:
		log.Info("Sucessfully stopped all components")
		return
	case _ = <-interruptTimeout:
		log.Error("Interrupting timed out")
		return
	}
}
