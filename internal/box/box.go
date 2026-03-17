package box

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Ox03bb/boxy/internal/config"
)

type Box struct {
	ID       string
	Name     string
	Root     string
	Hostname string
	Image    Image
	Pty      string
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

func NewBox(imageName Image) *Box {
	box := &Box{
		Image:  Image{},
		Ports:  make(map[string]string),
		Params: make(map[string]string),
		Env:    make(map[string]string),
	}

	box.GenerateID()
	box.GenerateName()
	box.SetHostname("")
	box.SetRoot("")

	return box
}

func (b *Box) GenerateID() string {
	if b.ID != "" {
		return b.ID
	}

	now := strconv.FormatInt(time.Now().UnixNano(), 10)

	random := make([]byte, 25)
	_, err := rand.Read(random)
	if err != nil {
		panic(err)
	}
	randomPart := hex.EncodeToString(random)

	hash := sha256.Sum256([]byte(now + randomPart))

	b.ID = hex.EncodeToString(hash[:])

	return b.ID
}

func (b *Box) GenerateName() string {
	if b.Name != "" {
		return b.Name
	}
	if b.ID == "" {
		b.GenerateID()
	}
	b.Name = "box-" + b.ID[:5]
	return b.Name
}

func (b *Box) SetHostname(hostname string) {
	if hostname == "" {
		b.Hostname = b.Name
		return
	}
	b.Hostname = hostname
}

func (b *Box) SetRoot(root string) (string, error) {
	if root == "" {
		envPath := os.ExpandEnv(config.EnvPath)

		if envPath == "" {
			return "", errors.New("EnvPath not set in config")
		}

		if b.ID == "" {
			b.GenerateID()
		}

		b.Root = filepath.Join(envPath, b.ID, "rootfs")

	}

	return b.Root, nil
}

func WriteBoxJSON(box *Box) error {
	envPath := os.ExpandEnv(config.EnvPath)

	filepath := filepath.Join(envPath, box.ID, "box.json")

	data, err := json.MarshalIndent(box, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal box: %w", err)
	}
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write box json %s: %w", filepath, err)
	}
	return nil
}

// Leadbox reads the box.json for a given box id and returns a Box object.
func Leadbox(id string) (*Box, error) {
	envPath := os.ExpandEnv(config.EnvPath)
	jsonPath := filepath.Join(envPath, id, "box.json")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed reading box json %s: %w", jsonPath, err)
	}

	var b Box
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("failed parsing box json %s: %w", jsonPath, err)
	}

	return &b, nil
}
