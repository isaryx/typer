package train

import (
	"strings"

	"typer/internal/analytics"
	"typer/internal/model"
)

const minKeyAttempts = 20

// MinKeyAttemptsPlacement is the lower sample threshold used right after typer train -e.
const MinKeyAttemptsPlacement = 10

// MergeSession updates per-key stats from a completed session and recomputes weak keys.
func MergeSession(p *Profile, result model.SessionResult) {
	if p.Keys == nil {
		p.Keys = make(map[string]KeyStat)
	}
	mergeKeyErrors(p, result)
	p.WeakKeys = computeWeakKeysMin(p.Keys, minKeyAttempts)
}

// FinalizePlacementProfile recomputes weak keys with a lower attempt threshold suited to the short placement test.
func FinalizePlacementProfile(p *Profile) {
	p.WeakKeys = computeWeakKeysMin(p.Keys, MinKeyAttemptsPlacement)
}

func mergeKeyErrors(p *Profile, result model.SessionResult) {
	words := strings.Fields(result.Prompt.Content)
	if result.Options.Strict && len(result.TypingTrace) > 0 && len(words) > 0 {
		mergeTraceKeyStats(p, words, true, result.TypingTrace)
		return
	}

	targetRunes := []rune(strings.Join(words, " "))
	typedRunes := []rune(result.TypedText)
	errorMap := map[string]int{}
	missingMap := map[string]int{}
	extraMap := map[string]int{}
	collectCharErrorsForTrain(targetRunes, typedRunes, errorMap, missingMap, extraMap)
	for key, count := range extraMap {
		errorMap[key] += count
	}

	attemptMap := countTargetChars(targetRunes)
	for key, attempts := range attemptMap {
		stat := p.Keys[key]
		stat.Attempts += attempts
		stat.Errors += errorMap[key]
		p.Keys[key] = stat
	}
	for key, count := range missingMap {
		stat := p.Keys[key]
		stat.Attempts += count
		stat.Errors += count
		p.Keys[key] = stat
	}
}

func mergeTraceKeyStats(p *Profile, words []string, strict bool, trace []model.ReplayEvent) {
	attempts, errors := replayTraceKeyStats(words, strict, trace)
	for key, n := range attempts {
		stat := p.Keys[key]
		stat.Attempts += n
		stat.Errors += errors[key]
		p.Keys[key] = stat
	}
}

// replayTraceKeyStats walks a typing trace and returns per-target-key attempts and errors.
func replayTraceKeyStats(words []string, strict bool, trace []model.ReplayEvent) (attempts, errors map[string]int) {
	if len(words) == 0 || len(trace) == 0 {
		return nil, nil
	}

	wordRunes := make([][]rune, len(words))
	for i, w := range words {
		wordRunes[i] = []rune(w)
	}

	attempts = map[string]int{}
	errors = map[string]int{}
	wordIdx := 0
	var current []rune

	for _, ev := range trace {
		switch ev.Kind {
		case model.ReplayEventKey:
			rs := []rune(ev.Rune)
			if len(rs) != 1 || wordIdx >= len(wordRunes) {
				continue
			}
			target := wordRunes[wordIdx]
			pos := len(current)
			pressed := rs[0]

			if pos < len(target) {
				expected := target[pos]
				key := normalizeTrainChar(expected)
				attempts[key]++
				matched := pressed == expected
				prefixBroken := !strict && !runesArePrefix(current, target)
				if !matched || prefixBroken {
					errors[key]++
				}
			}

			if strict && (pos >= len(target) || pressed != target[pos]) {
				continue
			}
			current = append(current, pressed)

		case model.ReplayEventBackspace:
			if len(current) > 0 {
				current = current[:len(current)-1]
			}

		case model.ReplayEventCommit:
			if wordIdx < len(wordRunes) {
				wordIdx++
				current = nil
			}
		}
	}
	return attempts, errors
}

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

func collectCharErrorsForTrain(target, typed []rune, wrong, missing, extra map[string]int) {
	n := min(len(target), len(typed))
	for i := 0; i < n; i++ {
		if target[i] != typed[i] {
			wrong[normalizeTrainChar(target[i])]++
		}
	}
	if len(target) > n {
		for _, ch := range target[n:] {
			missing[normalizeTrainChar(ch)]++
		}
	}
	if len(typed) > n {
		for _, ch := range typed[n:] {
			extra[normalizeTrainChar(ch)]++
		}
	}
}

func normalizeTrainChar(ch rune) string {
	switch ch {
	case ' ':
		return "<space>"
	case '\t':
		return "<tab>"
	case '\n':
		return "<newline>"
	default:
		return strings.ToLower(string(ch))
	}
}

func countTargetChars(target []rune) map[string]int {
	out := map[string]int{}
	for _, ch := range target {
		out[normalizeTrainChar(ch)]++
	}
	return out
}

func computeWeakKeysMin(keys map[string]KeyStat, minAttempts int) []string {
	type scored struct {
		key   string
		rate  float64
		count int
	}
	var items []scored
	for key, stat := range keys {
		if key == "<space>" || key == "<tab>" || key == "<newline>" {
			continue
		}
		if stat.Attempts < minAttempts {
			continue
		}
		rate := float64(stat.Errors) / float64(stat.Attempts)
		items = append(items, scored{key: key, rate: rate, count: stat.Errors})
	}
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].rate > items[i].rate ||
				(items[j].rate == items[i].rate && items[j].count > items[i].count) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	const topN = 5
	out := make([]string, 0, topN)
	for i := 0; i < len(items) && i < topN; i++ {
		if items[i].rate > 0 {
			out = append(out, items[i].key)
		}
	}
	return out
}

// TopSessionErrorChars returns the most mistyped target characters from one session.
func TopSessionErrorChars(result model.SessionResult, topN int) []analytics.CharCount {
	words := strings.Fields(result.Prompt.Content)
	if result.Options.Strict && len(result.TypingTrace) > 0 && len(words) > 0 {
		_, errorMap := replayTraceKeyStats(words, true, result.TypingTrace)
		return topCountsTrain(errorMap, topN)
	}

	targetRunes := []rune(strings.Join(words, " "))
	typedRunes := []rune(result.TypedText)
	errorMap := map[string]int{}
	missingMap := map[string]int{}
	extraMap := map[string]int{}
	collectCharErrorsForTrain(targetRunes, typedRunes, errorMap, missingMap, extraMap)
	for key, count := range extraMap {
		errorMap[key] += count
	}
	for key, count := range missingMap {
		errorMap[key] += count
	}
	return topCountsTrain(errorMap, topN)
}

func topCountsTrain(src map[string]int, n int) []analytics.CharCount {
	items := make([]analytics.CharCount, 0, len(src))
	for k, v := range src {
		items = append(items, analytics.CharCount{Key: k, Count: v})
	}
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Count > items[i].Count ||
				(items[j].Count == items[i].Count && items[j].Key < items[i].Key) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	if len(items) > n {
		items = items[:n]
	}
	return items
}
