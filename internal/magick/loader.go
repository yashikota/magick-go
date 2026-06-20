package magick

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/ebitengine/purego"
)

type Library struct {
	handle uintptr
	path   string
}

func Load(root string) (*Library, error) {
	path, err := libraryPath(root)
	if err != nil {
		return nil, err
	}
	handle, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return nil, fmt.Errorf("dlopen %s: %w", path, err)
	}
	lib := &Library{handle: handle, path: path}
	if err := register(handle); err != nil {
		_ = purego.Dlclose(handle)
		return nil, err
	}
	magickWandGenesis()
	return lib, nil
}

func (l *Library) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}

func libraryPath(root string) (string, error) {
	var patterns []string
	if runtime.GOOS == "darwin" {
		patterns = []string{
			filepath.Join(root, "lib", "libMagickWand-7.Q16HDRI.dylib"),
			filepath.Join(root, "lib", "libMagickWand-7.Q16HDRI.*.dylib"),
			filepath.Join(root, "lib", "libMagickWand-7.Q16.dylib"),
			filepath.Join(root, "lib", "libMagickWand-7.Q16.*.dylib"),
		}
	} else {
		patterns = []string{
			filepath.Join(root, "lib", "libMagickWand-7.Q16HDRI.so"),
			filepath.Join(root, "lib", "libMagickWand-7.Q16HDRI.so.*"),
			filepath.Join(root, "lib", "libMagickWand-7.Q16.so"),
			filepath.Join(root, "lib", "libMagickWand-7.Q16.so.*"),
		}
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return "", err
		}
		sort.Strings(matches)
		if len(matches) > 0 {
			return matches[0], nil
		}
	}
	return "", fmt.Errorf("libMagickWand was not found under %s/lib", root)
}
