package steve

import (
	"os"
	"strconv"
	"sync"

	"github.com/bearbin/mcgorcon"
	"github.com/hashicorp/go-hclog"
)

var (
	// rcon host
	host string
	// rcon port
	port int
	// rcon password
	pass string
	// the rcon client
	client mcgorcon.Client

	// logger
	log hclog.Logger
)

var (
	// the rcon componnent
	rcon *RCONComponent
)

func init() {
	log = hclog.Default().Named("steve")

	ok := true

	host = os.Getenv("STEVEBOT_RCON_HOST")
	if host == "" {
		ok = false
		log.Error("Missing STEVEBOT_RCON_HOST environment variable")
	}

	portString := os.Getenv("STEVEBOT_RCON_PORT")
	if portString == "" {
		ok = false
		log.Error("Missing STEVEBOT_RCON_PORT environment variable")
	} else {
		var err error
		port, err = strconv.Atoi(portString)
		if err != nil {
			ok = false
			log.Error("Invalid port", "err", err)
		}
	}

	pass = os.Getenv("STEVEBOT_RCON_PASS")
	if pass == "" {
		ok = false
		log.Error("Missing STEVEBOT_RCON_PASS environment variable")
	}

	if !ok {
		os.Exit(1)
	}
}

// RCONCommand is an interface to RCON Commands.
type RCONCommand interface {
	// Command returns the command that will be run.
	Command() string
	// Help returns a small help text for the command.
	Help() string
}

// RCONResult represents a result returned by executing a rcon command
type RCONResult struct {
	Out string
	Err error
}

// rconJob is a container for rcon jobs.
type rconJob struct {
	cmd RCONCommand
	out chan<- RCONResult
}

// SubmitRCONCommand submits a rcon command to be ran, returning a channel for a RCONResult.
func SubmitRCONCommand(cmd RCONCommand) <-chan RCONResult {
	out := make(chan RCONResult, 1)
	job := rconJob{
		cmd: cmd,
		out: out,
	}
	rcon.jobs <- job
	return out
}

// RCONComponent is the rcon component of the bot.
type RCONComponent struct {
	interrupt chan byte
	wg        *sync.WaitGroup
	jobs      chan rconJob
}

// Start starts the RCONComponent.
func (c *RCONComponent) Start(interrupt chan byte, wg *sync.WaitGroup) <-chan bool {
	c.interrupt = interrupt
	c.wg = wg
	c.jobs = make(chan rconJob)
	ready := make(chan bool, 1)

	c.wg.Add(1)
	go c.run(ready)

	rcon = c
	return ready
}

// Name returns the name of the component.
func (c *RCONComponent) Name() string {
	return "rcon"
}

func (c *RCONComponent) run(ready chan<- bool) {
	defer c.wg.Done()

	log.Info("Connecting to the RCON provider", "host", host, "port", port)

	var err error
	client, err = mcgorcon.Dial(host, port, pass)
	if err != nil {
		log.Error("Failed to connect to the RCON provider", "err", err)
		ready <- false
	}
	_, err = client.SendCommand("help")
	if err != nil {
		log.Error("Failed to connect to the RCON provider", "err", err)
		ready <- false
	}

	log.Info("Connected to the RCON provider", "host", host, "port", port)
	ready <- true

	for {
		select {
		case _ = <-c.interrupt:
			log.Info("Stopping RCON component")
			c.interrupt <- 0
			return
		case job := <-c.jobs:
			log.Debug("Running a job", "job", job)
			go func() {
				out, err := client.SendCommand(job.cmd.Command())
				job.out <- RCONResult{
					Out: out,
					Err: err,
				}
			}()
			continue
		}
	}
}
