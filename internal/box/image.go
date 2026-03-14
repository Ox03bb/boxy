package box

type Image struct {
	Name string
	Cmd  []string
}

type ImageService interface {
	pull(name string)
	push()
	remove(name string)
	List() []Image
}
