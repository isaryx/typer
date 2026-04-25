package text

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// maxCorpusBytes caps how much of a user-supplied corpus file we load into
// memory. 10 MiB is ~1.5M English words or ~20× the embedded passage bundle,
// so it is effectively unlimited for realistic use but protects against
// accidentally pointing --words-file at /dev/zero or a huge binary.
const maxCorpusBytes = 10 << 20 // 10 MiB

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
	f, err := os.Open(abs)
	if err != nil {
		return "", "", fmt.Errorf("read corpus file %q: %w", abs, err)
	}
	defer f.Close()
	// Read one extra byte so we can distinguish "exactly at limit" from "over".
	data, err := io.ReadAll(io.LimitReader(f, maxCorpusBytes+1))
	if err != nil {
		return "", "", fmt.Errorf("read corpus file %q: %w", abs, err)
	}
	if len(data) > maxCorpusBytes {
		return "", "", fmt.Errorf("corpus file %q exceeds %d-byte limit", abs, maxCorpusBytes)
	}
	return string(data), "file:" + abs, nil
}
