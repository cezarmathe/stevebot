package stevecmd

type Ban struct {
	dyn string
}

func NewBan(playerName string) *Ban {
	w := new(Ban)
	w.dyn = "ban " + playerName
	return w
}

func (w *Ban) Command() string {
	return w.dyn
}

func (w *Ban) Help() string {
	return "how to use ban: ban <player_name>"
}
