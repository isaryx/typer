package text

import "testing"

func TestCanonicalMode(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"passage", "passage", false},
		{"Passage", "passage", false},
		{"passages", "passage", false},
		{"p", "passage", false},
		{"words", "words", false},
		{"w", "words", false},
		{"quote", "quote", false},
		{"quotes", "quote", false},
		{"q", "quote", false},
		{"  Q  ", "quote", false},
		{"", "", true},
		{"xyz", "", true},
	}
	for _, tt := range tests {
		got, err := CanonicalMode(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("CanonicalMode(%q) err = nil, want error", tt.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("CanonicalMode(%q): %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("CanonicalMode(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNewProvider_canonicalModes(t *testing.T) {
	if _, err := NewProvider("passage", nil, "", "", QuoteProviderConfig{}); err != nil {
		t.Fatalf("NewProvider(passage): %v", err)
	}
	if _, err := NewProvider("words", nil, "", "", QuoteProviderConfig{}); err != nil {
		t.Fatalf("NewProvider(words): %v", err)
	}
	if _, err := NewProvider("p", nil, "", "", QuoteProviderConfig{}); err == nil {
		t.Fatalf("NewProvider(p) expected error: shorthands must be canonicalized first")
	}
}
