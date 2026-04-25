package text

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// loadCorpusOrBuiltin returns (content, sourceID). When path is empty/blank,
// it returns the built-in asset labeled builtinLabel. Otherwise it resolves
// the path to an absolute location, reads the file, and labels the source
// as "file:<abs-path>".
func loadCorpusOrBuiltin(path, builtin, builtinLabel string) (string, string, error) {
	if strings.TrimSpace(path) == "" {
		return builtin, "builtin:" + builtinLabel, nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("resolve corpus path: %w", err)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return "", "", fmt.Errorf("read corpus file %q: %w", abs, err)
	}
	return string(data), "file:" + abs, nil
}
