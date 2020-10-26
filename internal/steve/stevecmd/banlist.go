package stevecmd

type Banlist struct{}

func (w *Banlist) Command() string {
	return "banlist"
}

func (w *Banlist) Help() string {
	return "banlist: list the banned players"
}
