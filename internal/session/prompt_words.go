package session

import (
	"strings"

	"charm.land/lipgloss/v2"

	"typer/internal/ui"
)

// parsePromptContent splits prompt text into words and optional hard line groups.
// When lineWidth > 0 and content has newlines, each row is expanded or trimmed
// to fit the passage frame without soft-wrapping mid-row.
func parsePromptContent(content string, lineWidth int) (words []string, wordHardLine []int) {
	if !strings.Contains(content, "\n") {
		return strings.Fields(content), nil
	}

	lineIdx := 0
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if lineWidth > 0 {
			line = fitLineToWidth(line, lineWidth)
		}
		if line == "" {
			continue
		}
		for _, w := range strings.Fields(line) {
			words = append(words, w)
			wordHardLine = append(wordHardLine, lineIdx)
		}
		lineIdx++
	}
	return words, wordHardLine
}

// fitLineToWidth repeats a seed row to fill the frame, or trims if the seed alone is too wide.
func fitLineToWidth(seed string, width int) string {
	seed = strings.TrimSpace(seed)
	if seed == "" || width <= 0 {
		return seed
	}
	if lipgloss.Width(seed) > width {
		return truncateLineToWidth(seed, width)
	}
	var parts []string
	cur := 0
	for {
		add := lipgloss.Width(seed)
		if len(parts) > 0 {
			add++
		}
		if len(parts) > 0 && cur+add > width {
			break
		}
		parts = append(parts, seed)
		cur += add
		if cur >= width {
			break
		}
	}
	return strings.Join(parts, " ")
}

func truncateLineToWidth(line string, width int) string {
	words := strings.Fields(line)
	var parts []string
	cur := 0
	for _, w := range words {
		ww := lipgloss.Width(w)
		add := ww
		if len(parts) > 0 {
			add++
		}
		if len(parts) > 0 && cur+add > width {
			break
		}
		parts = append(parts, w)
		cur += add
	}
	return strings.Join(parts, " ")
}

func defaultPromptInnerWidth() int {
	return ui.FrameBodyInnerWidth(80)
}
