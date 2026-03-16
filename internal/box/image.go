package box

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ox03bb/boxy/internal/config"
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
func (i *Image) InitFs(b *Box) {
	registry := os.ExpandEnv(config.RegistryPath)

	imageDir, err := findImageDir(registry, i.Name)
	if err != nil {
		log.Printf("InitFs: %v", err)
		return
	}

	tarPath, err := findCompressedRootfs(imageDir)
	if err != nil {
		log.Printf("InitFs: %v", err)
		return
	}

	target, err := ensureBoxRoot(b)
	if err != nil {
		log.Printf("InitFs: %v", err)
		return
	}

	if err := extractTarGz(tarPath, target); err != nil {
		log.Printf("InitFs: extraction failed: %v", err)
		return
	}

	log.Printf("InitFs: extracted %s into %s", tarPath, target)
}

func findImageDir(registry, name string) (string, error) {
	entries, err := os.ReadDir(registry)
	if err != nil {
		return "", fmt.Errorf("failed reading registry %s: %w", registry, err)
	}
	for _, e := range entries {
		if e.IsDir() && e.Name() == name {
			return filepath.Join(registry, e.Name()), nil
		}
	}
	return "", fmt.Errorf("image %s not found in registry %s", name, registry)
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

func extractTarGz(tarPath, target string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("failed opening %s: %w", tarPath, err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed creating gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	cleanTarget := filepath.Clean(target)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read error: %w", err)
		}

		name := hdr.Name
		name = strings.TrimPrefix(name, "/")
		dest := filepath.Join(cleanTarget, name)

		// Prevent path traversal
		cleanedDest := filepath.Clean(dest)
		if !strings.HasPrefix(cleanedDest, cleanTarget+string(os.PathSeparator)) && cleanedDest != cleanTarget {
			log.Printf("extractTarGz: skipping invalid path %s", dest)
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(dest, os.FileMode(hdr.Mode)); err != nil {
				log.Printf("extractTarGz: mkdir %s: %v", dest, err)
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
				log.Printf("extractTarGz: mkdir for file %s: %v", dest, err)
				continue
			}
			out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(hdr.Mode))
			if err != nil {
				log.Printf("extractTarGz: open file %s: %v", dest, err)
				continue
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				log.Printf("extractTarGz: write file %s: %v", dest, err)
				continue
			}
			out.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
				log.Printf("extractTarGz: mkdir for symlink %s: %v", dest, err)
				continue
			}
			if err := os.Symlink(hdr.Linkname, dest); err != nil {
				log.Printf("extractTarGz: symlink %s -> %s: %v", dest, hdr.Linkname, err)
			}
		default:
			// ignore other types
		}
	}
	return nil
}
