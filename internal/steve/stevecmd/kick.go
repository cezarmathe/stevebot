package stevecmd

type Kick struct {
	dyn string
}

func NewKick(playerName string) *Kick {
	w := new(Kick)
	w.dyn = "kick " + playerName
	return w
}

func (w *Kick) Command() string {
	return w.dyn
}

func (w *Kick) Help() string {
	return "how to use kick: kick <player_name>"
}
