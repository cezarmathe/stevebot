package stevecmd

type Stop struct{}

func (l *Stop) Command() string {
	return "stop"
}

func (l *Stop) Help() string {
	return "stop: stop the server"
}
