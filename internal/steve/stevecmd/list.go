package stevecmd

type List struct{}

func (l *List) Command() string {
	return "list"
}

func (l *List) Help() string {
	return "list: list the number of players connected and their names"
}
