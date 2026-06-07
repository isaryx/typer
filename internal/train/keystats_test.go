package train

import (
	"testing"

	"typer/internal/model"
)

func TestReplayTraceKeyStatsStrictWrongKey(t *testing.T) {
	t.Parallel()

	trace := []model.ReplayEvent{
		{AtMS: 0, Kind: model.ReplayEventKey, Rune: "x"},
		{AtMS: 50, Kind: model.ReplayEventKey, Rune: "c"},
		{AtMS: 100, Kind: model.ReplayEventKey, Rune: "a"},
		{AtMS: 150, Kind: model.ReplayEventKey, Rune: "t"},
		{AtMS: 200, Kind: model.ReplayEventCommit},
	}
	attempts, errors := replayTraceKeyStats([]string{"cat"}, true, trace)
	if attempts["c"] != 2 || errors["c"] != 1 {
		t.Fatalf("c: attempts=%d errors=%d, want 2/1", attempts["c"], errors["c"])
	}
	if attempts["a"] != 1 || errors["a"] != 0 {
		t.Fatalf("a: attempts=%d errors=%d, want 1/0", attempts["a"], errors["a"])
	}
}

func TestMergeSessionStrictTraceWeakKeys(t *testing.T) {
	t.Parallel()

	p := NewProfile(PlacementResult{AssignedTier: TierFoundation, AssignedLesson: "1.1"})
	opts := model.SessionOptionsSnapshot{Strict: true, Mode: model.ModeTrain}

	for i := 0; i < 25; i++ {
		MergeSession(&p, model.SessionResult{
			Options:   opts,
			Prompt:    model.Prompt{Content: "vvvv"},
			TypedText: "vvvv",
			TypingTrace: strictTraceWithWrongBeforeEach("vvvv", 'b'),
		})
	}
	if len(p.WeakKeys) == 0 || p.WeakKeys[0] != "v" {
		t.Fatalf("WeakKeys = %v, want v first", p.WeakKeys)
	}
}

func TestMergeSessionStrictPerfectTraceNoWeakKeys(t *testing.T) {
	t.Parallel()

	p := NewProfile(PlacementResult{AssignedTier: TierFoundation, AssignedLesson: "1.1"})
	opts := model.SessionOptionsSnapshot{Strict: true, Mode: model.ModeTrain}

	for i := 0; i < 25; i++ {
		MergeSession(&p, model.SessionResult{
			Options:   opts,
			Prompt:    model.Prompt{Content: "vvvv"},
			TypedText: "vvvv",
			TypingTrace: strictTracePerfectWord("vvvv"),
		})
	}
	if len(p.WeakKeys) != 0 {
		t.Fatalf("WeakKeys = %v, want none for perfect strict typing", p.WeakKeys)
	}
}

func TestTopSessionErrorCharsStrictTrace(t *testing.T) {
	t.Parallel()

	top := TopSessionErrorChars(model.SessionResult{
		Options:   model.SessionOptionsSnapshot{Strict: true},
		Prompt:    model.Prompt{Content: "cat"},
		TypedText: "cat",
		TypingTrace: []model.ReplayEvent{
			{AtMS: 0, Kind: model.ReplayEventKey, Rune: "x"},
			{AtMS: 50, Kind: model.ReplayEventKey, Rune: "c"},
			{AtMS: 100, Kind: model.ReplayEventKey, Rune: "a"},
			{AtMS: 150, Kind: model.ReplayEventKey, Rune: "t"},
			{AtMS: 200, Kind: model.ReplayEventCommit},
		},
	}, 1)
	if len(top) != 1 || top[0].Key != "c" || top[0].Count != 1 {
		t.Fatalf("top = %+v, want c:1", top)
	}
}

func strictTracePerfectWord(word string) []model.ReplayEvent {
	var trace []model.ReplayEvent
	var at int64
	for _, r := range word {
		trace = append(trace, model.ReplayEvent{AtMS: at, Kind: model.ReplayEventKey, Rune: string(r)})
		at += 50
	}
	trace = append(trace, model.ReplayEvent{AtMS: at, Kind: model.ReplayEventCommit})
	return trace
}

func strictTraceWithWrongBeforeEach(word string, wrong rune) []model.ReplayEvent {
	var trace []model.ReplayEvent
	var at int64
	for _, r := range word {
		trace = append(trace, model.ReplayEvent{AtMS: at, Kind: model.ReplayEventKey, Rune: string(wrong)})
		at += 50
		trace = append(trace, model.ReplayEvent{AtMS: at, Kind: model.ReplayEventKey, Rune: string(r)})
		at += 50
	}
	trace = append(trace, model.ReplayEvent{AtMS: at, Kind: model.ReplayEventCommit})
	return trace
}
