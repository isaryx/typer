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
	wordRunes         [][]rune
	wordLens          []int
	typedWords        []string
	wordMatches       []bool
	wordIndex         int
	current           string
	currentRunes      []rune
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
	wordRunes := make([][]rune, n)
	wordLens := make([]int, n)
	traceCap := n * 6
	if traceCap < 16 {
		traceCap = 16
	}
	for i, w := range words {
		wordRunes[i] = []rune(w)
		wordLens[i] = len(wordRunes[i])
	}
	return &typingState{
		words:       words,
		wordRunes:   wordRunes,
		wordLens:    wordLens,
		strict:      strict,
		now:         now,
		typedWords:  make([]string, 0, n),
		wordMatches: make([]bool, 0, n),
		wpmSamples:  make([]float64, 0, n),
		typingTrace: make([]model.ReplayEvent, 0, traceCap),
	}
}

func (s *typingState) replaceWords(words []string) {
	n := len(words)
	wordRunes := make([][]rune, n)
	wordLens := make([]int, n)
	for i, w := range words {
		wordRunes[i] = []rune(w)
		wordLens[i] = len(wordRunes[i])
	}
	s.words = words
	s.wordRunes = wordRunes
	s.wordLens = wordLens
}

func (s *typingState) syncCurrentString() {
	s.current = string(s.currentRunes)
}

// reconcileCurrentRunes rebuilds currentRunes from current when they diverge (e.g. tests set current directly).
func (s *typingState) reconcileCurrentRunes() {
	if s.current == "" {
		s.currentRunes = nil
		return
	}
	if string(s.currentRunes) != s.current {
		s.currentRunes = []rune(s.current)
	}
}

// runesArePrefix reports whether cur equals the first len(cur) runes of tgt.
func runesArePrefix(cur, tgt []rune) bool {
	if len(cur) > len(tgt) {
		return false
	}
	for i := 0; i < len(cur); i++ {
		if cur[i] != tgt[i] {
			return false
		}
	}
	return true
}

// applyRunes appends input runes for the active word. Returns whether this call started the session clock
// (first user input in the run), and whether any keystroke should trigger mistake feedback (bell).
// In non-strict mode, once the buffer is no longer a prefix of the target word, every further keystroke
// counts as mistake feedback until the user fixes with backspace (the word cannot match without that).
func (s *typingState) applyRunes(runes []rune) (sessionClockStarted bool, mistake bool) {
	if len(runes) == 0 || s.wordIndex >= len(s.words) {
		return false, false
	}
	s.reconcileCurrentRunes()
	if s.startedAt.IsZero() {
		s.startedAt = s.now().UTC()
		sessionClockStarted = true
	}

	target := s.wordRunes[s.wordIndex]
	current := s.currentRunes
	for _, r := range runes {
		s.appendTraceKey(r)
		pos := len(current)
		s.totalKeystrokes++
		matched := pos < len(target) && r == target[pos]
		if matched {
			s.correctKeystrokes++
		}
		prefixBroken := !s.strict && !runesArePrefix(current, target)
		if !matched || prefixBroken {
			mistake = true
		}
		if s.strict && !matched {
			continue
		}
		current = append(current, r)
	}
	s.currentRunes = current
	s.syncCurrentString()
	return sessionClockStarted, mistake
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
	s.reconcileCurrentRunes()

	if len(s.currentRunes) == 0 {
		r.emptyCurrent = true
		return r
	}
	currentInput := s.current
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

	_, mismatches := scoring.CompareRunes(s.wordRunes[s.wordIndex], s.currentRunes)
	s.uncorrectedErrors += mismatches

	if s.wordIndex > 0 {
		s.typedCharCount++
	}
	s.typedCharCount += len(s.currentRunes)

	s.typedWords = append(s.typedWords, currentInput)
	s.wordMatches = append(s.wordMatches, matched)
	s.wordIndex++
	s.currentRunes = nil
	s.current = ""
	s.appendWPMSample()
	r.advanced = true
	return r
}

// applyBackspace removes the last rune from the current word buffer and records trace when applicable.
func (s *typingState) applyBackspace() {
	s.reconcileCurrentRunes()
	if len(s.currentRunes) == 0 {
		return
	}
	s.appendTraceBackspace()
	s.currentRunes = s.currentRunes[:len(s.currentRunes)-1]
	s.syncCurrentString()
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
