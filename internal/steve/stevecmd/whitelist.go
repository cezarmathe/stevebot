package stevecmd

type Whitelist struct {
	dyn string
}

func NewWhitelistList() *Whitelist {
	w := new(Whitelist)
	w.dyn = "whitelist list"
	return w
}

func NewWhitelistOn() *Whitelist {
	w := new(Whitelist)
	w.dyn = "whitelist on"
	return w
}

func NewWhitelistOff() *Whitelist {
	w := new(Whitelist)
	w.dyn = "whitelist off"
	return w
}

func NewWhitelistAdd(playerName string) *Whitelist {
	w := new(Whitelist)
	w.dyn = "whitelist add " + playerName
	return w
}

func NewWhitelistRemove(playerName string) *Whitelist {
	w := new(Whitelist)
	w.dyn = "whitelist remove " + playerName
	return w
}

func (w *Whitelist) Command() string {
	return w.dyn
}

func (w *Whitelist) Help() string {
	return "whitelist: whitelist add <player_name> / whitelist list / whitelist off / whitelist on / whitelist remove <player_name>"
}
