package box

type Image struct {
	Id      string
	Name    string
	Version string
	Path    string
}

type ImageService interface {
	pull(name string, version string)
	push()
	remove(id string)
	List() []Image
}
