package defense

import (
	"testing"
	"time"

	"typer/internal/model"
)

func TestBuildSessionResult(t *testing.T) {
	started := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	ended := started.Add(90 * time.Second)
	r := Result{
		Score:     12,
		Lives:     0,
		Elapsed:   90 * time.Second,
		StartedAt: started,
		EndedAt:   ended,
		Over:      true,
	}
	got := BuildSessionResult(r)
	if got.Prompt.Mode != model.ModeDefense {
		t.Fatalf("mode=%q", got.Prompt.Mode)
	}
	if got.ElapsedMS != 90000 {
		t.Fatalf("elapsed=%d", got.ElapsedMS)
	}
	if got.Aborted {
		t.Fatal("should not be aborted")
	}
}
