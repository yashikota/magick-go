package runtimebundle

import (
	"os"
	"path/filepath"
	"runtime"
)

func ConfigureEnvironment(root, configDir string) {
	coderPath, filterPath := modulePaths(root)
	setenv("MAGICK_HOME", root)
	setenv("MAGICK_CONFIGURE_PATH", configurePath(root, configDir))
	setenv("MAGICK_CODER_MODULE_PATH", coderPath)
	setenv("MAGICK_FILTER_MODULE_PATH", filterPath)
	prependPath("PATH", filepath.Join(root, "bin"))
	if runtime.GOOS == "darwin" {
		prependPath("DYLD_LIBRARY_PATH", filepath.Join(root, "lib"))
	} else {
		prependPath("LD_LIBRARY_PATH", filepath.Join(root, "lib"))
	}
}

func setenv(k, v string) {
	_ = os.Setenv(k, v)
}

func prependPath(k, v string) {
	if cur := os.Getenv(k); cur != "" {
		_ = os.Setenv(k, v+string(os.PathListSeparator)+cur)
		return
	}
	_ = os.Setenv(k, v)
}

func Environment(root, configDir string) map[string]string {
	coderPath, filterPath := modulePaths(root)
	return map[string]string{
		"MAGICK_HOME":               root,
		"MAGICK_CONFIGURE_PATH":     configurePath(root, configDir),
		"MAGICK_CODER_MODULE_PATH":  coderPath,
		"MAGICK_FILTER_MODULE_PATH": filterPath,
	}
}

func configurePath(root, configDir string) string {
	sharedConfig := filepath.Join(root, "etc", "ImageMagick-7")
	if configDir == "" {
		return sharedConfig
	}
	return configDir + string(os.PathListSeparator) + sharedConfig
}

func modulePaths(root string) (string, string) {
	coder := firstGlob(
		filepath.Join(root, "lib", "ImageMagick", "modules-*", "coders"),
		filepath.Join(root, "lib", "ImageMagick-*", "modules-*", "coders"),
	)
	filter := firstGlob(
		filepath.Join(root, "lib", "ImageMagick", "modules-*", "filters"),
		filepath.Join(root, "lib", "ImageMagick-*", "modules-*", "filters"),
	)
	if coder == "" {
		coder = filepath.Join(root, "lib", "ImageMagick", "modules-Q16HDRI", "coders")
	}
	if filter == "" {
		filter = filepath.Join(root, "lib", "ImageMagick", "modules-Q16HDRI", "filters")
	}
	return coder, filter
}

func firstGlob(patterns ...string) string {
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return matches[0]
		}
	}
	return ""
}
