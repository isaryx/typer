package model

import "testing"

func TestValidateSessionOptions(t *testing.T) {
	if err := ValidateSessionOptions(SessionOptions{Mode: ModeWords, Words: 0}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSessionOptions(SessionOptions{Mode: ModeWords, Words: 15}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSessionOptions(SessionOptions{Mode: ModeWords, Words: MaxWordsPerPrompt}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSessionOptions(SessionOptions{Mode: ModeWords, Words: -1}); err == nil {
		t.Fatal("expected error for negative words")
	}
	if err := ValidateSessionOptions(SessionOptions{Mode: ModeWords, Words: MaxWordsPerPrompt + 1}); err == nil {
		t.Fatal("expected error for words above max")
	}
}

func TestValidateHistoryLast(t *testing.T) {
	if err := ValidateHistoryLast(1); err != nil {
		t.Fatal(err)
	}
	if err := ValidateHistoryLast(MaxRetainedHistorySessions); err != nil {
		t.Fatal(err)
	}
	if err := ValidateHistoryLast(0); err == nil {
		t.Fatal("expected error")
	}
	if err := ValidateHistoryLast(MaxRetainedHistorySessions + 1); err == nil {
		t.Fatal("expected error")
	}
}
