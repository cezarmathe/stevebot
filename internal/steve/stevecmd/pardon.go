package stevecmd

type Pardon struct {
	dyn string
}

func NewPardon(playerName string) *Pardon {
	w := new(Pardon)
	w.dyn = "pardon " + playerName
	return w
}

func (w *Pardon) Command() string {
	return w.dyn
}

func (w *Pardon) Help() string {
	return "how to use pardon: pardon <player_name>"
}
