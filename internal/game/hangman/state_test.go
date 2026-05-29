package hangman

import "testing"

func TestRecordMistakeScaling(t *testing.T) {
	s := NewState(Config{MaxStrikes: 6, MistakesPerStrike: 3})
	for i := 1; i <= 2; i++ {
		if lost := s.RecordMistake(); lost {
			t.Fatalf("mistake %d: unexpected loss", i)
		}
		if s.Stage() != 0 {
			t.Fatalf("mistake %d: stage = %d, want 0", i, s.Stage())
		}
	}
	if lost := s.RecordMistake(); lost {
		t.Fatal("3rd mistake should advance stage, not lose")
	}
	if s.Stage() != 1 {
		t.Fatalf("after 3 mistakes: stage = %d, want 1", s.Stage())
	}
}

func TestRecordMistakeLossAtMaxStage(t *testing.T) {
	s := NewState(Config{MaxStrikes: 6, MistakesPerStrike: 1})
	for i := 1; i <= 5; i++ {
		if lost := s.RecordMistake(); lost {
			t.Fatalf("mistake %d: unexpected loss at stage %d", i, s.Stage())
		}
	}
	if !s.RecordMistake() {
		t.Fatal("6th mistake should lose")
	}
	if s.Stage() != 6 {
		t.Fatalf("stage = %d, want 6", s.Stage())
	}
}

func TestValidateConfigRejectsNonSixStrikes(t *testing.T) {
	if err := ValidateConfig(Config{MaxStrikes: 5, MistakesPerStrike: 1}); err == nil {
		t.Fatal("expected error for strikes != 6")
	}
}
