package text

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewLocalProvider_Builtin(t *testing.T) {
	p, err := NewLocalProvider("")
	if err != nil {
		t.Fatalf("NewLocalProvider: %v", err)
	}
	if !strings.HasPrefix(p.Name(), "builtin:") {
		t.Fatalf("want builtin source, got %q", p.Name())
	}
	got, err := p.Next(context.Background(), Constraints{})
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if got.Content == "" {
		t.Fatalf("empty passage")
	}
}

func TestNewLocalProvider_CustomFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "passages.txt")
	body := "first block line\nanother line\n\nsecond block\n"
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	p, err := NewLocalProvider(tmp)
	if err != nil {
		t.Fatalf("NewLocalProvider: %v", err)
	}
	if !strings.HasPrefix(p.Name(), "file:") {
		t.Fatalf("want file source, got %q", p.Name())
	}
	if len(p.passages) != 2 {
		t.Fatalf("want 2 passages, got %d", len(p.passages))
	}
}

func TestNewLocalProvider_EmptyFileRejected(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "empty.txt")
	if err := os.WriteFile(tmp, []byte("   \n\n  \n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := NewLocalProvider(tmp); err == nil {
		t.Fatalf("expected error for empty passage file")
	}
}

func TestNewLocalProvider_MissingFile(t *testing.T) {
	if _, err := NewLocalProvider(filepath.Join(t.TempDir(), "nope.txt")); err == nil {
		t.Fatalf("expected error for missing file")
	}
}

func TestLoadCorpus_RejectsOversizedFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "huge.txt")
	big := make([]byte, maxCorpusBytes+1)
	for i := range big {
		big[i] = 'a'
	}
	if err := os.WriteFile(tmp, big, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, _, err := loadCorpusOrBuiltin(tmp, "", ""); err == nil {
		t.Fatalf("expected oversize corpus error, got nil")
	}
}
