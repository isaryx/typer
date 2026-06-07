package cli

import (
	"io"
	"strings"
	"testing"

	"typer/internal/train"
)

func TestReadTrainContinueKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want trainContinueAction
	}{
		{"\n", trainContinueNext},
		{"\r\n", trainContinueNext},
		{" ", trainContinueRetry},
		{"\x03", trainContinueStop},
	}
	for _, tc := range cases {
		got, err := readTrainContinueKey(strings.NewReader(tc.in))
		if err != nil {
			t.Fatalf("input %q: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("input %q: got %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestPromptTrainContinueFailMapsEnterToRetry(t *testing.T) {
	t.Parallel()

	outcome := train.LessonOutcome{Passed: false, LessonID: "1.1"}
	got, err := promptTrainContinue(strings.NewReader("\n"), io.Discard, outcome, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != trainContinueRetry {
		t.Fatalf("got %v, want retry", got)
	}
}

func TestPromptTrainContinuePassEnterIsNext(t *testing.T) {
	t.Parallel()

	outcome := train.LessonOutcome{Passed: true, LessonID: "1.1", NextLesson: "1.2"}
	got, err := promptTrainContinue(strings.NewReader("\n"), io.Discard, outcome, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != trainContinueNext {
		t.Fatalf("got %v, want next", got)
	}
}

func TestPromptTrainContinuePassSpaceIsRetry(t *testing.T) {
	t.Parallel()

	outcome := train.LessonOutcome{Passed: true, LessonID: "1.1", NextLesson: "1.2"}
	got, err := promptTrainContinue(strings.NewReader(" "), io.Discard, outcome, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != trainContinueRetry {
		t.Fatalf("got %v, want retry", got)
	}
}
