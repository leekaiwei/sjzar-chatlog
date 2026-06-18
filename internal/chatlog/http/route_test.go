package http

import (
	"os"
	"path/filepath"
	"testing"

	chatctx "github.com/sjzar/chatlog/internal/chatlog/ctx"
)

func TestResolveDataPathRejectsTraversal(t *testing.T) {
	base := t.TempDir()
	s := &Service{ctx: &chatctx.Context{DataDir: base}}

	if _, ok := s.resolveDataPath("../secret.txt"); ok {
		t.Fatal("expected traversal outside data dir to be rejected")
	}
}

func TestResolveDataPathAllowsNestedDataFile(t *testing.T) {
	base := t.TempDir()
	s := &Service{ctx: &chatctx.Context{DataDir: base}}

	want := filepath.Join(base, "nested", "media.jpg")
	for _, in := range []string{"nested/media.jpg", "/nested/media.jpg"} {
		got, ok := s.resolveDataPath(in)
		if !ok {
			t.Fatalf("expected %q to be allowed", in)
		}
		if got != want {
			t.Fatalf("resolveDataPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveDataPathAllowsDotDotPrefixFilename(t *testing.T) {
	base := t.TempDir()
	s := &Service{ctx: &chatctx.Context{DataDir: base}}

	got, ok := s.resolveDataPath("..foo")
	if !ok {
		t.Fatal("expected filename beginning with .. to be allowed")
	}
	want := filepath.Join(base, "..foo")
	if got != want {
		t.Fatalf("resolveDataPath() = %q, want %q", got, want)
	}
}

func TestResolveExistingDataPathRejectsSymlinkEscape(t *testing.T) {
	base := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.dat"), []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(base, "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	s := &Service{ctx: &chatctx.Context{DataDir: base}}
	if _, ok := s.resolveExistingDataPath(filepath.Join("link", "secret.dat")); ok {
		t.Fatal("expected symlink escape outside data dir to be rejected")
	}
}
