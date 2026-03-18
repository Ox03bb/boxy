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
	"strings"
	"time"

	"github.com/Ox03bb/boxy/internal/config"
)

const ( //Box Status
	Created = "created"
	Running = "running"
	Stopped = "stopped"
	Exited  = "exited"
)

type Box struct {
	ID         string
	Name       string
	Root       string
	Hostname   string
	Image      Image
	Status     string
	Created_at time.Time
	Pty        string
	PIDs       []int
	Ports      map[string]string
	Params     map[string]string
	Env        map[string]string
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

func UpdateStatus(containerID, newStatus string) error {
	if containerID == "" {
		return fmt.Errorf("containerID is required")
	}

	b, err := Loadbox(containerID)
	if err != nil {
		return fmt.Errorf("failed to load box by ID %s: %w", containerID, err)

	}
	b.Status = newStatus
	err = WriteBoxJSON(b)
	if err != nil {
		return fmt.Errorf("failed to write box json: %w", err)
	}
	return nil

}

func Loadbox(id string) (*Box, error) {
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

// LoadAllBoxes scans the EnvPath directory and loads all valid box.json files.
func LoadAllBoxes() ([]*Box, error) {
	envPath := os.ExpandEnv(config.EnvPath)

	entries, err := os.ReadDir(envPath)
	if err != nil {
		return nil, err
	}

	var boxes []*Box
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		b, err := Loadbox(e.Name())
		if err != nil {
			// skip invalid or unreadable boxes
			continue
		}
		boxes = append(boxes, b)
	}

	return boxes, nil
}

// FindBoxByName finds a box by its name.
func FindBoxByName(name string) (*Box, error) {
	boxes, err := LoadAllBoxes()
	if err != nil {
		return nil, err
	}
	for _, b := range boxes {
		if b.Name == name {
			return b, nil
		}
	}
	return nil, fmt.Errorf("box with name %s not found", name)
}

// FindBoxesByPrefix returns all boxes whose ID starts with the given prefix.
func FindBoxesByPrefix(prefix string, boxes []*Box) ([]*Box, error) {
	if boxes == nil {
		var err error
		boxes, err = LoadAllBoxes()
		if err != nil {
			return nil, err
		}
	}

	var matches []*Box
	for _, b := range boxes {
		if strings.HasPrefix(b.ID, prefix) {
			matches = append(matches, b)
		}
	}
	return matches, nil
}

// ResolveBoxIdentifier resolves an identifier which may be a full ID, a
// 3+ character prefix of an ID, or a name (when isName is true). It returns
// the matching Box or an error describing the problem (not found or
// ambiguous).
func ResolveBoxIdentifier(idOrName string, isName bool) (*Box, error) {
	if isName {
		return FindBoxByName(idOrName)
	}

	// try exact lookup by ID first (Loadbox expects full ID)
	if b, err := Loadbox(idOrName); err == nil {
		return b, nil
	}

	if len(idOrName) < 3 {
		return nil, fmt.Errorf("provide at least 3 characters to match partial ID")
	}

	matches, err := FindBoxesByPrefix(idOrName, nil)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no box starts with %s", idOrName)
	}
	if len(matches) > 1 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("multiple boxes start with %s:", idOrName))
		for _, m := range matches {
			short := m.ID
			if len(short) > 12 {
				short = short[:12]
			}
			sb.WriteString(fmt.Sprintf("\n - %s (%s)", short, m.Name))
		}
		return nil, fmt.Errorf("%s", sb.String())
	}
	return matches[0], nil
}
