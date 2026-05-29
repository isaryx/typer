package defense

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"

	"typer/internal/model"
	"typer/internal/version"
)

// DisplayScore returns words destroyed plus whole seconds survived.
func (r Result) DisplayScore() int {
	sec := int(r.Elapsed.Round(time.Second).Seconds())
	if sec < 0 {
		sec = 0
	}
	return r.Score + sec
}

// BuildSessionResult maps a defense run into a lightweight history entry.
func BuildSessionResult(r Result) model.SessionResult {
	started := r.StartedAt.UTC()
	ended := r.EndedAt.UTC()
	if ended.IsZero() {
		ended = started.Add(r.Elapsed)
	}
	elapsedMS := r.Elapsed.Milliseconds()
	if elapsedMS < 0 {
		elapsedMS = 0
	}
	return model.SessionResult{
		ID:        newDefenseSessionID(started),
		StartedAt: started,
		EndedAt:   ended,
		ElapsedMS: elapsedMS,
		Prompt: model.Prompt{
			ID:      started.Format(time.RFC3339Nano),
			Content: fmt.Sprintf("%d words destroyed", r.Score),
			Mode:    model.ModeDefense,
		},
		Aborted:      r.Aborted,
		ResultSchema: model.SessionResultSchema,
		TyperVersion: version.Version,
		Options:      model.OptionsSnapshot(model.SessionOptions{Mode: model.ModeDefense}),
	}
}

func newDefenseSessionID(startedAt time.Time) string {
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}
	id, err := ulid.New(ulid.Timestamp(startedAt), ulid.Monotonic(rand.Reader, 0))
	if err != nil {
		return startedAt.UTC().Format(time.RFC3339Nano)
	}
	return id.String()
}
