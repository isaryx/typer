package session

import (
	"time"
	"unicode/utf8"

	"typer/internal/model"
	"typer/internal/scoring"
)

// typingState holds word-lane typing progress, metrics, and replay trace — independent of Bubble Tea.
type typingState struct {
	words             []string
	typedWords        []string
	wordMatches       []bool
	wordIndex         int
	current           string
	strict            bool
	typedCharCount    int
	totalErrors       int
	totalKeystrokes   int
	correctKeystrokes int
	uncorrectedErrors int
	wpmSamples        []float64
	typingTrace       []model.ReplayEvent
	startedAt         time.Time
	now               func() time.Time
}

func newTypingState(words []string, strict bool, now func() time.Time) *typingState {
	if now == nil {
		now = time.Now
	}
	n := len(words)
	return &typingState{
		words:       words,
		strict:      strict,
		now:         now,
		typedWords:  make([]string, 0, n),
		wordMatches: make([]bool, 0, n),
		wpmSamples:  make([]float64, 0, n),
	}
}

// applyRunes appends input runes for the active word. Returns true if this call started the session clock
// (first user input in the run).
func (s *typingState) applyRunes(runes []rune) bool {
	if len(runes) == 0 || s.wordIndex >= len(s.words) {
		return false
	}
	var sessionClockStarted bool
	if s.startedAt.IsZero() {
		s.startedAt = s.now().UTC()
		sessionClockStarted = true
	}

	target := []rune(s.words[s.wordIndex])
	current := []rune(s.current)
	for _, r := range runes {
		s.appendTraceKey(r)
		pos := len(current)
		s.totalKeystrokes++
		matched := pos < len(target) && r == target[pos]
		if matched {
			s.correctKeystrokes++
		}
		if s.strict && !matched {
			continue
		}
		current = append(current, r)
	}
	s.current = string(current)
	return sessionClockStarted
}

type commitWordResult struct {
	advanced            bool
	emptyCurrent        bool
	strictBlocked       bool
	sessionClockStarted bool
}

// applyCommitWord commits the active word (Space/Enter). Caller maps emptyCurrent / strictBlocked to UI status.
func (s *typingState) applyCommitWord() commitWordResult {
	var r commitWordResult
	if s.wordIndex >= len(s.words) {
		return r
	}

	currentInput := s.current
	if currentInput == "" {
		r.emptyCurrent = true
		return r
	}
	targetWord := s.words[s.wordIndex]
	matched := currentInput == targetWord
	if s.strict && !matched {
		r.strictBlocked = true
		return r
	}

	if s.startedAt.IsZero() {
		s.startedAt = s.now().UTC()
		r.sessionClockStarted = true
	}

	s.appendTraceCommit()

	if !matched {
		s.totalErrors++
	}

	_, mismatches := scoring.CompareRunes([]rune(targetWord), []rune(currentInput))
	s.uncorrectedErrors += mismatches

	if s.wordIndex > 0 {
		s.typedCharCount++
	}
	s.typedCharCount += utf8.RuneCountInString(currentInput)

	s.typedWords = append(s.typedWords, currentInput)
	s.wordMatches = append(s.wordMatches, matched)
	s.wordIndex++
	s.current = ""
	s.appendWPMSample()
	r.advanced = true
	return r
}

// applyBackspace removes the last rune from the current word buffer and records trace when applicable.
func (s *typingState) applyBackspace() {
	if s.current == "" {
		return
	}
	s.appendTraceBackspace()
	s.current = removeLastRune(s.current)
}

func (s *typingState) appendWPMSample() {
	elapsedMinutes := s.now().Sub(s.startedAt).Minutes()
	if elapsedMinutes <= 0 {
		return
	}
	gross := (float64(s.typedCharCount) / 5.0) / elapsedMinutes
	s.wpmSamples = append(s.wpmSamples, gross)
}

func (s *typingState) traceAtMS() int64 {
	if s.startedAt.IsZero() {
		return 0
	}
	return s.now().Sub(s.startedAt).Milliseconds()
}

func (s *typingState) appendTraceKey(r rune) {
	s.typingTrace = append(s.typingTrace, model.ReplayEvent{
		AtMS: s.traceAtMS(),
		Kind: model.ReplayEventKey,
		Rune: string(r),
	})
}

func (s *typingState) appendTraceBackspace() {
	s.typingTrace = append(s.typingTrace, model.ReplayEvent{
		AtMS: s.traceAtMS(),
		Kind: model.ReplayEventBackspace,
	})
}

func (s *typingState) appendTraceCommit() {
	s.typingTrace = append(s.typingTrace, model.ReplayEvent{
		AtMS: s.traceAtMS(),
		Kind: model.ReplayEventCommit,
	})
}

func (s *typingState) isDone() bool {
	return s.wordIndex >= len(s.words) && len(s.words) > 0
}

func removeLastRune(str string) string {
	if str == "" {
		return ""
	}
	_, size := utf8.DecodeLastRuneInString(str)
	return str[:len(str)-size]
}
