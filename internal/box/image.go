package box

import (
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
		boxDir := filepath.Join(envPath, b.Name)
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
