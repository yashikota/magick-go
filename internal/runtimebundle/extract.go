package runtimebundle

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
)

const markerFile = ".magickgo-runtime-ready"

func Ensure() (*Bundle, error) {
	target, err := Target()
	if err != nil {
		return nil, err
	}
	data, hash, err := EmbeddedRuntime(target)
	if err != nil {
		return nil, err
	}
	root, err := RuntimeDir(target, hash)
	if err != nil {
		return nil, err
	}
	if ready(root, hash) {
		return &Bundle{Target: target, Hash: hash, Root: root}, nil
	}
	if err := extractTarZst(data, root, hash); err != nil {
		return nil, err
	}
	return &Bundle{Target: target, Hash: hash, Root: root}, nil
}

func ready(root, hash string) bool {
	b, err := os.ReadFile(filepath.Join(root, markerFile))
	return err == nil && strings.TrimSpace(string(b)) == hash
}

func extractTarZst(data []byte, root, hash string) error {
	parent := filepath.Dir(root)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(parent, filepath.Base(root)+"-*.tmp")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	if err := untarZst(bytes.NewReader(data), tmp); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmp, markerFile), []byte(hash+"\n"), 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, root); err != nil {
		if ready(root, hash) {
			return nil
		}
		return err
	}
	return nil
}

func untarZst(r io.Reader, dest string) error {
	zr, err := zstd.NewReader(r)
	if err != nil {
		return err
	}
	defer zr.Close()

	tr := tar.NewReader(zr)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if h.Name == "." || h.Name == "./" {
			continue
		}
		clean, err := safeJoin(dest, h.Name)
		if err != nil {
			return err
		}
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(clean, os.FileMode(h.Mode)&0o777); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(clean, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(h.Mode)&0o777)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(f, tr)
			closeErr := f.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		case tar.TypeSymlink:
			if err := validateLinkTarget(filepath.Dir(clean), h.Linkname, dest); err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
				return err
			}
			if err := os.Symlink(h.Linkname, clean); err != nil {
				return err
			}
		case tar.TypeLink:
			link, err := safeJoin(dest, h.Linkname)
			if err != nil {
				return err
			}
			if err := os.Link(link, clean); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported tar entry %s type %d", h.Name, h.Typeflag)
		}
	}
}

func safeJoin(root, name string) (string, error) {
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("archive contains absolute path %q", name)
	}
	clean := filepath.Clean(name)
	if clean == "." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return "", fmt.Errorf("archive contains unsafe path %q", name)
	}
	full := filepath.Join(root, clean)
	rel, err := filepath.Rel(root, full)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("archive escapes destination: %q", name)
	}
	return full, nil
}

func validateLinkTarget(linkDir, target, root string) error {
	if target == "" {
		return fmt.Errorf("archive contains empty symlink target")
	}
	var full string
	if filepath.IsAbs(target) {
		full = filepath.Clean(target)
	} else {
		full = filepath.Join(linkDir, target)
	}
	rel, err := filepath.Rel(root, full)
	if err != nil {
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("archive symlink escapes destination: %q", target)
	}
	return nil
}
