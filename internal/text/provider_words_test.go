package text

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewWordsProvider_DefaultBuiltinList(t *testing.T) {
	p, err := NewWordsProvider("")
	if err != nil {
		t.Fatalf("new words provider: %v", err)
	}
	if !strings.HasPrefix(p.Name(), "builtin:") {
		t.Fatalf("expected builtin source name, got %q", p.Name())
	}

	prompt, err := p.Next(context.Background(), Constraints{Words: 5})
	if err != nil {
		t.Fatalf("next prompt: %v", err)
	}
	if len(strings.Fields(prompt.Content)) != 5 {
		t.Fatalf("expected 5 words in prompt, got %q", prompt.Content)
	}
}

func TestNewWordsProvider_CustomFile(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "my-words.txt")
	if err := os.WriteFile(customPath, []byte("alpha\nbeta\ngamma\n"), 0o644); err != nil {
		t.Fatalf("write custom words file: %v", err)
	}

	p, err := NewWordsProvider(customPath)
	if err != nil {
		t.Fatalf("new words provider with custom file: %v", err)
	}
	if !strings.HasPrefix(p.Name(), "file:") {
		t.Fatalf("expected custom file source name, got %q", p.Name())
	}

	prompt, err := p.Next(context.Background(), Constraints{Words: 20})
	if err != nil {
		t.Fatalf("next prompt: %v", err)
	}
	for _, word := range strings.Fields(prompt.Content) {
		switch word {
		case "alpha", "beta", "gamma":
		default:
			t.Fatalf("unexpected word from custom list: %q", word)
		}
	}
}
