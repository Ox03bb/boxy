package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func FindImageDir(registry, name string) (string, error) {
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

func FindCompressedRootfs(Dir string) (string, error) {
	files, err := os.ReadDir(Dir)
	if err != nil {
		return "", fmt.Errorf("failed reading image dir %s: %w", Dir, err)
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		n := f.Name()
		if strings.HasSuffix(n, ".tar.gz") || strings.HasSuffix(n, ".tgz") {
			return filepath.Join(Dir, n), nil
		}
	}
	return "", fmt.Errorf("no compressed rootfs found in %s", Dir)
}

func ExtractTarGz(tarPath, target string) error {
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
