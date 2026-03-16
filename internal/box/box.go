package box

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"time"
)

type Box struct {
	id       string
	name     string
	root     string
	hostname string
	image    string
	pty      *os.File
	ports    map[string]string
	pram     map[string]string
}

type BoxService interface {
	run(id string) error
	stop(id string) error
	restart(id string) error
	remove(id string) error
}

func (b *Box) GenerateID() string {
	now := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(string(now)))

	return hex.EncodeToString(hash[:])
}

func (b *Box) GenerateName(id string) string {
	return "box-" + id[:5]
}

func (b *Box) CreateRootfs() {
	envPath := os.Getenv("EnvPath")
	if envPath == "" {
		return
	}
	boxDir := envPath + string(os.PathSeparator) + b.name
	err := os.MkdirAll(boxDir, 0755)
	if err != nil {
		return
	}
	b.root = boxDir

}
