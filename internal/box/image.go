package box

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ox03bb/boxy/internal/config"
	"github.com/Ox03bb/boxy/internal/utils"
)

type Image struct {
	Name string
	Env  map[string]string
	Cmd  []string
}

type ImageService interface {
	pull(name string)
	push()
	remove(name string)
	List() []Image
}

// InitFs finds the image by name inside config.RegistryPath, locates a compressed
// rootfs (tar.gz / .tgz) and extracts it into the box root directory.
func (i *Image) InitFs(b *Box) error {
	registry := os.ExpandEnv(config.RegistryPath)

	imageDir, err := utils.FindImageDir(registry, i.Name)
	if err != nil {
		log.Printf("InitFs: %v", err)
		return err
	}

	tarPath, err := findCompressedRootfs(imageDir)
	if err != nil {
		log.Printf("InitFs: %v", err)
		return err
	}

	target, err := ensureBoxRoot(b)
	if err != nil {
		log.Printf("InitFs: %v", err)
		return err
	}

	if err := utils.ExtractTarGz(tarPath, target); err != nil {
		log.Printf("InitFs: extraction failed: %v", err)
		return err
	}

	log.Printf("InitFs: extracted %s into %s", tarPath, target)
	return nil
}

func findCompressedRootfs(imageDir string) (string, error) {
	files, err := os.ReadDir(imageDir)
	if err != nil {
		return "", fmt.Errorf("failed reading image dir %s: %w", imageDir, err)
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		n := f.Name()
		if strings.HasSuffix(n, ".tar.gz") || strings.HasSuffix(n, ".tgz") {
			return filepath.Join(imageDir, n), nil
		}
	}
	return "", fmt.Errorf("no compressed rootfs found in %s", imageDir)
}

func ensureBoxRoot(b *Box) (string, error) {
	target := b.Root
	if target == "" {
		envPath := os.ExpandEnv(config.EnvPath)
		boxDir := filepath.Join(envPath, b.ID, "rootfs")
		if err := os.MkdirAll(boxDir, 0755); err != nil {
			return "", fmt.Errorf("failed creating box dir %s: %w", boxDir, err)
		}
		b.Root = boxDir
		return boxDir, nil
	}
	if err := os.MkdirAll(target, 0755); err != nil {
		return "", fmt.Errorf("failed ensuring box root %s: %w", target, err)
	}
	return target, nil
}

// loadImage reads registry/<name>/image.json and returns an Image object.
// The JSON `cmd` field may be either a string or an array of strings.
func loadImage(name string) (*Image, error) {
	registry := os.ExpandEnv(config.RegistryPath)
	jsonPath := filepath.Join(registry, name, "image.json")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed reading image json %s: %w", jsonPath, err)
	}

	var tmp struct {
		Name string            `json:"name"`
		Env  map[string]string `json:"env"`
		Cmd  interface{}       `json:"cmd"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return nil, fmt.Errorf("failed parsing image json %s: %w", jsonPath, err)
	}

	var cmd []string
	switch v := tmp.Cmd.(type) {
	case nil:
		// leave cmd nil
	case string:
		if v != "" {
			cmd = []string{v}
		}
	case []interface{}:
		for _, x := range v {
			s, ok := x.(string)
			if !ok {
				return nil, fmt.Errorf("invalid cmd array element in %s", jsonPath)
			}
			cmd = append(cmd, s)
		}
	default:
		return nil, fmt.Errorf("unsupported cmd type in %s", jsonPath)
	}

	img := &Image{
		Name: tmp.Name,
		Env:  tmp.Env,
		Cmd:  cmd,
	}
	return img, nil
}
