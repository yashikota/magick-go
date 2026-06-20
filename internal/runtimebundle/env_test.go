package runtimebundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigurePathPrependsProcessConfig(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "runtime")
	configDir := filepath.Join(string(filepath.Separator), "policy")

	got := configurePath(root, configDir)
	want := filepath.Join(root, "etc", "ImageMagick-7") + string(os.PathListSeparator) + configDir
	if got != want {
		t.Fatalf("configurePath() = %q, want %q", got, want)
	}
}

func TestConfigurePathIncludesModuleConfig(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(string(filepath.Separator), "policy")
	moduleConfig := filepath.Join(root, "lib", "ImageMagick", "config-Q16HDRI")
	if err := os.MkdirAll(moduleConfig, 0o755); err != nil {
		t.Fatal(err)
	}

	got := configurePath(root, configDir)
	want := moduleConfig +
		string(os.PathListSeparator) + filepath.Join(root, "etc", "ImageMagick-7") +
		string(os.PathListSeparator) + configDir
	if got != want {
		t.Fatalf("configurePath() = %q, want %q", got, want)
	}
}
