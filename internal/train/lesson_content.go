package train

import (
	"fmt"
	"strings"
)

// EffectiveRounds returns how many drill lines to include in one lesson session.
func (l Lesson) EffectiveRounds() int {
	if l.Rounds > 0 {
		return l.Rounds
	}
	if l.TimedMS > 0 || l.Assessment {
		return 1
	}
	switch l.Tier {
	case TierFoundation:
		return 4
	case TierBuilding:
		return 3
	case TierFluent:
		return 2
	default:
		return 1
	}
}

// PromptLines returns seed drill rows; the session TUI fills each to the frame width.
func (l Lesson) PromptLines() []string {
	if len(l.Prompts) == 0 {
		return nil
	}
	n := l.EffectiveRounds()
	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = strings.TrimSpace(l.Prompts[i%len(l.Prompts)])
	}
	return lines
}

// BuildLessonContent assembles seed lines into one newline-separated prompt.
func BuildLessonContent(l Lesson, filter *WordFilter) (string, error) {
	if lines := l.PromptLines(); len(lines) > 0 {
		return strings.Join(lines, "\n"), nil
	}
	if l.PromptWords > 0 && filter != nil {
		lineCount := max(l.EffectiveRounds(), 1)
		count := l.PromptWords * lineCount
		words, matched := filter.WordsContaining(l.Keys, count)
		if len(words) == 0 {
			return "", fmt.Errorf("no words available for lesson %s", l.ID)
		}
		if len(l.Keys) > 0 && !matched {
			return "", fmt.Errorf("no words containing keys for lesson %s", l.ID)
		}
		return strings.Join(wordLineSeeds(words, lineCount), "\n"), nil
	}
	return "", fmt.Errorf("lesson %s has no prompts", l.ID)
}

func wordLineSeeds(words []string, lineCount int) []string {
	if lineCount <= 1 {
		return []string{strings.Join(words, " ")}
	}
	perLine := (len(words) + lineCount - 1) / lineCount
	lines := make([]string, 0, lineCount)
	for i := 0; i < len(words); i += perLine {
		end := i + perLine
		if end > len(words) {
			end = len(words)
		}
		lines = append(lines, strings.Join(words[i:end], " "))
	}
	return lines
}
