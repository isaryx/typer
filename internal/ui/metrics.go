package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

const metricsSpeedControlHeader = "SPEED                        CONTROL"

// writerTerminalWidth returns the layout width for metrics (same cap as the typing session wrap width).
func writerTerminalWidth(out io.Writer) int {
	const fallback = 80
	f, ok := out.(*os.File)
	if !ok {
		return fallback
	}
	fd := int(f.Fd())
	if !term.IsTerminal(fd) {
		return fallback
	}
	w, _, err := term.GetSize(fd)
	if err != nil || w < 24 {
		return fallback
	}
	if w > MaxContentWidth {
		return MaxContentWidth
	}
	return w
}

// metricsInnerWidth is the inner text width inside the metrics box (between side borders),
// matching session.FrameBodyInnerWidth so the stats panel matches the prompt frame width.
func metricsInnerWidth(layoutWidth int) int {
	return FrameBodyInnerWidth(layoutWidth)
}

// PrintMetricsTable renders session metrics; uses lipgloss on a terminal file, plain ASCII otherwise.
func PrintMetricsTable(out io.Writer, heading string, gross, net, adjusted, acc, cons float64, errCount int, elapsedMS int64, summary bool) {
	tw := writerTerminalWidth(out)
	inner := metricsInnerWidth(tw)
	f, ok := out.(*os.File)
	useLip := ok && term.IsTerminal(int(f.Fd()))
	if useLip {
		printMetricsTableLipgloss(out, heading, gross, net, adjusted, acc, cons, errCount, elapsedMS, summary, inner)
		return
	}
	printMetricsTablePlain(out, heading, gross, net, adjusted, acc, cons, errCount, elapsedMS, summary, inner)
}

func printMetricsTablePlain(out io.Writer, heading string, gross, net, adjusted, acc, cons float64, errCount int, elapsedMS int64, summary bool, boxInnerWidth int) {
	fmt.Fprintln(out)
	horizontalBar := strings.Repeat("─", TopMiddleWidth(boxInnerWidth))
	boxBottomFmt := "╰%s╯\n"
	boxInnerFmt := fmt.Sprintf("│ %%-%ds │\n", boxInnerWidth)
	twoColumnFmt := "%-16s %-10s  %-14s %s"
	fmt.Fprintf(out, "%s", BuildRoundedTopPlain("", heading, boxInnerWidth))
	if summary {
		fmt.Fprintf(out, boxInnerFmt, metricsSpeedControlHeader)
		fmt.Fprintf(out, boxInnerFmt, fmt.Sprintf(twoColumnFmt, "Avg gross WPM", fmt.Sprintf("%.2f", gross), "Avg accuracy", fmt.Sprintf("%.2f%%", acc)))
		fmt.Fprintf(out, boxInnerFmt, fmt.Sprintf(twoColumnFmt, "Avg adjusted WPM", fmt.Sprintf("%.2f", adjusted), "Avg pace stability", fmt.Sprintf("%.2f", cons)))
		fmt.Fprintf(out, boxInnerFmt, fmt.Sprintf(twoColumnFmt, "Avg net WPM", fmt.Sprintf("%.2f", net), "Total errors", fmt.Sprintf("%d", errCount)))
		fmt.Fprintf(out, boxInnerFmt, "")
		fmt.Fprintf(out, boxInnerFmt, fmt.Sprintf("Total time: %s", FormatElapsedMS(elapsedMS)))
	} else {
		fmt.Fprintf(out, boxInnerFmt, metricsSpeedControlHeader)
		fmt.Fprintf(out, boxInnerFmt, fmt.Sprintf(twoColumnFmt, "Gross WPM", fmt.Sprintf("%.2f", gross), "Accuracy", fmt.Sprintf("%.2f%%", acc)))
		fmt.Fprintf(out, boxInnerFmt, fmt.Sprintf(twoColumnFmt, "Adjusted WPM", fmt.Sprintf("%.2f", adjusted), "Pace stability", fmt.Sprintf("%.2f", cons)))
		fmt.Fprintf(out, boxInnerFmt, fmt.Sprintf(twoColumnFmt, "Net WPM", fmt.Sprintf("%.2f", net), "Errors", fmt.Sprintf("%d", errCount)))
		fmt.Fprintf(out, boxInnerFmt, "")
		fmt.Fprintf(out, boxInnerFmt, fmt.Sprintf("Duration: %s", FormatElapsedMS(elapsedMS)))
	}
	fmt.Fprintf(out, boxBottomFmt, horizontalBar)
	fmt.Fprintln(out)
}

func printMetricsTableLipgloss(out io.Writer, heading string, gross, net, adjusted, acc, cons float64, errCount int, elapsedMS int64, summary bool, inner int) {
	meta := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMeta))
	label := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTitle)).Bold(true)
	val := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorInputFg))
	border := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBorderHex))

	twoColumnFmt := "%-16s %-10s  %-14s %s"
	twoCol := func(l1, v1, l2, v2 string) string {
		line := fmt.Sprintf(twoColumnFmt, l1, v1, l2, v2)
		return meta.Width(inner).Render(line)
	}

	mw := TopMiddleWidth(inner)
	topLine := RenderRoundedTop("", border, meta, heading, mw)

	var innerLines []string
	innerLines = append(innerLines, label.Width(inner).Render(metricsSpeedControlHeader))
	if summary {
		innerLines = append(innerLines, twoCol("Avg gross WPM", fmt.Sprintf("%.2f", gross), "Avg accuracy", fmt.Sprintf("%.2f%%", acc)))
		innerLines = append(innerLines, twoCol("Avg adjusted WPM", fmt.Sprintf("%.2f", adjusted), "Avg pace stability", fmt.Sprintf("%.2f", cons)))
		innerLines = append(innerLines, twoCol("Avg net WPM", fmt.Sprintf("%.2f", net), "Total errors", fmt.Sprintf("%d", errCount)))
		innerLines = append(innerLines, "")
		innerLines = append(innerLines, lipgloss.JoinHorizontal(lipgloss.Top, meta.Render("Total time: "), val.Render(FormatElapsedMS(elapsedMS))))
	} else {
		innerLines = append(innerLines, twoCol("Gross WPM", fmt.Sprintf("%.2f", gross), "Accuracy", fmt.Sprintf("%.2f%%", acc)))
		innerLines = append(innerLines, twoCol("Adjusted WPM", fmt.Sprintf("%.2f", adjusted), "Pace stability", fmt.Sprintf("%.2f", cons)))
		innerLines = append(innerLines, twoCol("Net WPM", fmt.Sprintf("%.2f", net), "Errors", fmt.Sprintf("%d", errCount)))
		innerLines = append(innerLines, "")
		innerLines = append(innerLines, lipgloss.JoinHorizontal(lipgloss.Top, meta.Render("Duration: "), val.Render(FormatElapsedMS(elapsedMS))))
	}

	var b strings.Builder
	b.WriteString(topLine)
	b.WriteString("\n")
	for _, line := range innerLines {
		b.WriteString(RenderRoundedSide("", border, inner, line))
		b.WriteString("\n")
	}
	b.WriteString(RenderRoundedBottomPlain("", border, mw))

	fmt.Fprintln(out)
	fmt.Fprintln(out, b.String())
	fmt.Fprintln(out)
}

// FormatElapsedMS renders elapsed milliseconds for metrics and replay summaries.
func FormatElapsedMS(ms int64) string {
	if ms < 0 {
		return "0 ms"
	}
	if ms < 1000 {
		return fmt.Sprintf("%d ms", ms)
	}
	secs := ms / 1000
	if secs < 60 {
		if ms < 10_000 {
			return fmt.Sprintf("%.1f s", float64(ms)/1000)
		}
		return fmt.Sprintf("%d s", secs)
	}
	mins := secs / 60
	secs %= 60
	if mins < 60 {
		return fmt.Sprintf("%dm %02ds", mins, secs)
	}
	h := mins / 60
	mins %= 60
	return fmt.Sprintf("%dh %02dm %02ds", h, mins, secs)
}
