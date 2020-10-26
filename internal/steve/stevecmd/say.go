package stevecmd

type Say struct {
	dyn string
}

func NewSay(saysmth string) *Say {
	w := new(Say)
	w.dyn = "say stevebot: " + saysmth
	return w
}

func (w *Say) Command() string {
	return w.dyn
}

func (w *Say) Help() string {
	return "how to use say: say <something>"
}
