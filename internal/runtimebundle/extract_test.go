package runtimebundle

import "testing"

func TestSafeJoinRejectsTraversal(t *testing.T) {
	if _, err := safeJoin("/tmp/root", "../escape"); err == nil {
		t.Fatal("safeJoin accepted traversal")
	}
	if _, err := safeJoin("/tmp/root", "/absolute"); err == nil {
		t.Fatal("safeJoin accepted absolute path")
	}
}

func TestSafeJoinAcceptsRelativePath(t *testing.T) {
	got, err := safeJoin("/tmp/root", "lib/libMagickWand.so")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/root/lib/libMagickWand.so" {
		t.Fatalf("safeJoin() = %q", got)
	}
}

func TestValidateLinkTargetRejectsEscapes(t *testing.T) {
	for _, target := range []string{"../../escape", "/tmp/escape"} {
		if err := validateLinkTarget("/tmp/root/lib", target, "/tmp/root"); err == nil {
			t.Fatalf("validateLinkTarget accepted %q", target)
		}
	}
}

func TestValidateLinkTargetAcceptsInternalTarget(t *testing.T) {
	if err := validateLinkTarget("/tmp/root/lib", "../share/fonts", "/tmp/root"); err != nil {
		t.Fatal(err)
	}
}
