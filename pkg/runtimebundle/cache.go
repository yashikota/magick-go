package runtimebundle

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Bundle struct {
	Target string
	Hash   string
	Root   string
}

func Target() (string, error) {
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "linux/amd64":
		return "linux-amd64", nil
	case "linux/arm64":
		return "linux-arm64", nil
	case "darwin/arm64":
		return "darwin-arm64", nil
	default:
		return "", fmt.Errorf("unsupported target %s/%s", runtime.GOOS, runtime.GOARCH)
	}
}

func CacheRoot() (string, error) {
	if runtime.GOOS == "darwin" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Caches", "magickgo", "runtime"), nil
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "magickgo", "runtime"), nil
}

func EmbeddedRuntime(target string) ([]byte, string, error) {
	name := "assets/runtime-" + target + ".tar.zst"
	data, err := assets.ReadFile(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, "", fmt.Errorf("embedded runtime %s is missing; build it with scripts/build-runtime-linux.sh or scripts/build-runtime-darwin.sh, or let CI build it", name)
		}
		return nil, "", err
	}
	sum := sha256.Sum256(data)
	return data, hex.EncodeToString(sum[:]), nil
}

func RuntimeDir(target, hash string) (string, error) {
	cache, err := CacheRoot()
	if err != nil {
		return "", err
	}
	shortHash := hash
	if len(shortHash) > 16 {
		shortHash = shortHash[:16]
	}
	return filepath.Join(cache, target+"-"+shortHash), nil
}
