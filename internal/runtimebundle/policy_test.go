package runtimebundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyPolicyUsesProcessTempDir(t *testing.T) {
	dir, err := ApplyPolicy(false)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if filepath.Base(dir) == "ImageMagick-7" {
		t.Fatalf("ApplyPolicy wrote to shared config directory: %q", dir)
	}
	b, err := os.ReadFile(filepath.Join(dir, "policy.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `pattern="PDF"`) {
		t.Fatalf("policy.xml does not contain safe policy: %s", b)
	}
}
