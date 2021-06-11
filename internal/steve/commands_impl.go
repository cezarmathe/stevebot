package steve

type steveCommandInputImpl struct {
	inner     []string
	innerChan chan<- rconCommandOutput
}

type steveCommandOutputImpl struct {
	err       error
	innerChan <-chan rconCommandOutput
}

func newSteveCommand(command []string) (*steveCommandInputImpl,
	*steveCommandOutputImpl) {

	in := new(steveCommandInputImpl)
	in.inner = command

	out := new(steveCommandOutputImpl)
	out.err = nil

	innerChan := make(chan rconCommandOutput, 1)
	in.innerChan = innerChan
	out.innerChan = innerChan

	return in, out
}

func newSteveCommandOutput(err error) *steveCommandOutputImpl {
	return &steveCommandOutputImpl{err, nil}
}

func (c *steveCommandInputImpl) Command() []string {
	return c.inner
}

func (c *steveCommandInputImpl) inChan() chan<- rconCommandOutput {
	return c.innerChan
}

func (c *steveCommandOutputImpl) Error() string {
	return c.err.Error()
}

func (c *steveCommandOutputImpl) Success() bool {
	return c.err == nil
}

func (c *steveCommandOutputImpl) OutChan() <-chan rconCommandOutput {
	return c.innerChan
}

type rconCommandInputImpl struct {
	inner string
}

type rconCommandOutputImpl struct {
	out string
	err error
}

func newRconCommandInput(command string) *rconCommandInputImpl {
	return &rconCommandInputImpl{command}
}

func newRconCommandOutput(out string, err error) *rconCommandOutputImpl {
	return &rconCommandOutputImpl{out, err}
}

func (c *rconCommandInputImpl) Command() string {
	return c.inner
}

func (c *rconCommandOutputImpl) Error() string {
	return c.err.Error()
}

func (c *rconCommandOutputImpl) Out() string {
	return c.out
}

func (c *rconCommandOutputImpl) Success() bool {
	return c.err == nil
}
