package session

import (
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
)

// Styles is the shared TUI palette used by session and game modes.
type Styles = tuiStyles

// DefaultStyles returns the standard session color styles.
func DefaultStyles() Styles {
	return defaultStyles()
}

// TitleStyle returns the session heading style.
func (s Styles) TitleStyle() lipgloss.Style { return s.title }

// MetaStyle returns auxiliary caption chrome style.
func (s Styles) MetaStyle() lipgloss.Style { return s.meta }

// BorderStyle returns frame border style.
func (s Styles) BorderStyle() lipgloss.Style { return s.border }

func RenderUpcomingText(styles Styles, text string) string {
	return styles.upcoming.Render(text)
}

// RenderMetaText styles auxiliary chrome (brackets, captions).
func RenderMetaText(styles Styles, text string) string {
	return styles.meta.Render(text)
}

// RenderErrorText styles error messages.
func RenderErrorText(styles Styles, text string) string {
	return styles.errorMessage.Render(text)
}

func RenderActiveWordText(styles Styles, target, typed string) string {
	return renderActiveWordWithTyped(styles, target, typed, false)
}

func renderActiveWordWithTyped(styles Styles, target, typed string, isLastWord bool) string {
	targetRunes := []rune(target)
	if len(targetRunes) == 0 {
		return styles.active.Render(target)
	}

	cursor := utf8.RuneCountInString(typed)
	if cursor >= len(targetRunes) {
		var b strings.Builder
		b.WriteString(styles.activeTyped.Render(string(targetRunes)))
		if isLastWord {
			b.WriteString(styles.activePlain.Underline(true).Render(" "))
		}
		return b.String()
	}

	var b strings.Builder
	if cursor > 0 {
		b.WriteString(styles.activeTyped.Render(string(targetRunes[:cursor])))
	}
	for j := cursor; j < len(targetRunes); j++ {
		ch := string(targetRunes[j])
		st := styles.activePlain
		if j == cursor {
			st = st.Underline(true)
		}
		b.WriteString(st.Render(ch))
	}
	return b.String()
}
