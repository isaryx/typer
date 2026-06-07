package defense

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadWordPool_Builtin(t *testing.T) {
	pool, err := LoadWordPool("")
	if err != nil {
		t.Fatalf("load pool: %v", err)
	}
	if pool.Len() == 0 {
		t.Fatal("expected non-empty defense word pool from builtin list")
	}
}

func TestLoadWordPool_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(path, []byte("\n\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, err := LoadWordPool(path)
	if err == nil {
		t.Fatal("expected error for empty word list")
	}
}

func TestLoadWordPool_NoDefenseWords(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "short.txt")
	if err := os.WriteFile(path, []byte("a\nb\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, err := LoadWordPool(path)
	if err == nil {
		t.Fatal("expected error when no words match defense length bounds")
	}
	if !strings.Contains(err.Error(), "no words between") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadWordPool_CustomValid(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "words.txt")
	if err := os.WriteFile(path, []byte("cat\ncode\nterminal\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	pool, err := LoadWordPool(path)
	if err != nil {
		t.Fatalf("load pool: %v", err)
	}
	if pool.Len() != 3 {
		t.Fatalf("pool len = %d, want 3", pool.Len())
	}
}
