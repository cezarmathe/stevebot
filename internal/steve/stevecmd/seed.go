package stevecmd

type Seed struct{}

func (l *Seed) Command() string {
	return "seed"
}

func (l *Seed) Help() string {
	return "seed: get the world seed"
}
