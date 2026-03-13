package box

type Box struct {
	id       string
	name     string
	root     string
	hostname string
	image    string
	ports    map[string]string
	pram     map[string]string
}

type BoxService interface {
	run(id string) error
	stop(id string) error
	restart(id string) error
	remove(id string) error
}
