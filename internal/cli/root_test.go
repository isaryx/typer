package cli

import (
	"reflect"
	"testing"
)

func TestExtractPresenceFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantRest     []string
		wantStrict   bool
		wantIndef    bool
		wantErr      bool
	}{
		{"empty", nil, nil, false, false, false},
		{"no flags", []string{"--mode", "words"}, []string{"--mode", "words"}, false, false, false},
		{"bare strict", []string{"--strict"}, nil, true, false, false},
		{"strict then flags", []string{"--strict", "--mode", "words"}, []string{"--mode", "words"}, true, false, false},
		{"strict shorthand", []string{"-s", "-m", "words"}, []string{"-m", "words"}, true, false, false},
		{"bare indefinite", []string{"--indefinite"}, nil, false, true, false},
		{"indef shorthand", []string{"-i", "-m", "w"}, []string{"-m", "w"}, false, true, false},
		{"strict and indefinite", []string{"-s", "-i", "-w", "3"}, []string{"-w", "3"}, true, true, false},
		{"eq strict", []string{"--strict=false"}, nil, false, false, true},
		{"eq indefinite", []string{"-i=true"}, nil, false, false, true},
		{"space true after strict", []string{"--strict", "true"}, nil, false, false, true},
		{"word after strict ok", []string{"--strict", "hello"}, []string{"hello"}, true, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRest, gotStrict, gotIndef, err := extractPresenceFlags(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("extractPresenceFlags: err = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("extractPresenceFlags: %v", err)
			}
			if gotStrict != tt.wantStrict {
				t.Fatalf("strict = %v, want %v", gotStrict, tt.wantStrict)
			}
			if gotIndef != tt.wantIndef {
				t.Fatalf("indefinite = %v, want %v", gotIndef, tt.wantIndef)
			}
			if !reflect.DeepEqual(gotRest, tt.wantRest) {
				t.Fatalf("rest = %#v, want %#v", gotRest, tt.wantRest)
			}
		})
	}
}
