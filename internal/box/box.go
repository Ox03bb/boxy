package box

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"time"
)

type Box struct {
	ID       string
	Name     string
	Root     string
	Hostname string
	Image    string
	Pty      *os.File
	Ports    map[string]string
	Params   map[string]string
	Env      map[string]string
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

	b.ID = hex.EncodeToString(hash[:])
	return b.ID
}

func (b *Box) GenerateName(id string) string {
	b.Name = "box-" + id[:5]
	return b.Name
}

func (b *Box) SetHostname(hostname string) {
	if hostname == "" {
		b.Hostname = b.Name
		return
	}
	b.Hostname = hostname
}

func (b *Box) CreateRootfs() (string, error) {
	envPath := os.Getenv("EnvPath")
	if envPath == "" {
		return "", errors.New("EnvPath not set")
	}
	boxDir := envPath + string(os.PathSeparator) + b.ID
	err := os.MkdirAll(boxDir, 0755)
	if err != nil {
		return "", err
	}
	b.Root = boxDir

	return b.Root, nil
}
