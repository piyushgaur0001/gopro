package storage

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFilename(t *testing.T) {
	store := NewLocalStore(t.TempDir(), t.TempDir(), t.TempDir())

	name, err := store.NewFilename(".jpg")
	if err != nil {
		t.Fatalf("NewFilename returned error: %v", err)
	}
	if filepath.Ext(name) != ".jpg" {
		t.Fatalf("filename %q does not preserve extension", name)
	}
	if strings.Contains(name, "/") || strings.Contains(name, "..") {
		t.Fatalf("filename %q is not path-safe", name)
	}
}

func TestIsSafeFilename(t *testing.T) {
	tests := map[string]bool{
		"abc.jpg":        true,
		"":               false,
		"../abc.jpg":     false,
		"nested/abc.jpg": false,
		"abc..jpg":       false,
	}

	for name, want := range tests {
		if got := IsSafeFilename(name); got != want {
			t.Fatalf("IsSafeFilename(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestLocalStoreSaveAndOpenCompressed(t *testing.T) {
	root := t.TempDir()
	store := NewLocalStore(filepath.Join(root, "uploads"), filepath.Join(root, "compressed"), filepath.Join(root, "tmp"))
	if err := store.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs returned error: %v", err)
	}

	if err := store.SaveCompressed("image.jpg", []byte("data")); err != nil {
		t.Fatalf("SaveCompressed returned error: %v", err)
	}

	file, info, err := store.OpenCompressed("image.jpg")
	if err != nil {
		t.Fatalf("OpenCompressed returned error: %v", err)
	}
	defer file.Close()

	if info.Size() != 4 {
		t.Fatalf("got file size %d, want 4", info.Size())
	}
}
