package session

import (
	"testing"
	"time"

	"typer/internal/model"
)

func TestTimedSessionExpires(t *testing.T) {
	t.Parallel()

	start := time.Unix(1000, 0)
	now := start
	m := newTypingSessionModel(
		model.Prompt{Content: "one two three four five six seven eight nine ten"},
		false,
		func() time.Time { return now },
		false, nil, false, false, false, false,
		model.DefaultInputPlacement(),
		nil, nil, 500,
	)

	m.appendRunes([]rune("one"))
	if m.startedAt.IsZero() {
		t.Fatal("expected session clock to start")
	}

	now = start.Add(600 * time.Millisecond)
	updated, _ := m.Update(sessionTickMsg{})
	fm := updated.(*typingSessionModel)
	if !fm.timerExpired {
		t.Fatal("expected timer to expire after duration")
	}
	res := fm.result()
	if res.Aborted {
		t.Fatal("timer expiry should not count as abort")
	}
	if !res.Completed {
		t.Fatal("expected completed on timer expiry")
	}
}
