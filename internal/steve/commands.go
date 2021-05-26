package steve

// SteveCommandInput is the input required by steve for handling commands.
type SteveCommandInput interface {
	// Command returns the command as a list of words.
	Command() []string

	// InChan returns the sending-side of the channel through which the output
	// is passed.
	InChan() chan<- RconCommandOutput
}

// RconCommandInput is the input required to send a rcon command.
type RconCommandInput interface {
	// Command returns the command.
	Command() string
}

// SteveCommandOutput is the output produced by steve after submitting a
// command.
type SteveCommandOutput interface {
	error

	// Success returns whether the command was started successfully or not.
	Success() bool

	// OutChan returns the receiving-side of the channel through which the
	// output is passed.
	OutChan() <-chan RconCommandOutput
}

// RconCommandOutput is the output produced by sending a command via rcon.
type RconCommandOutput interface {
	error

	// Out returns the message sent back by the Minecraft Server via rcon.
	Out() string

	// Success returns whether the command completed successfully or not.
	Success() bool
}
