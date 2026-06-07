package defense

import (
	"context"
	"fmt"
	"io"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Run starts the defense TUI and returns the final result.
func Run(ctx context.Context, input io.Reader, output io.Writer, pool WordPool, cfg Config, seed uint64) (Result, error) {
	if err := ValidateConfig(cfg); err != nil {
		return Result{}, err
	}

	var bellOut io.Writer
	if !cfg.NoAudible {
		bellOut = output
	}

	m := newDefenseModel(pool, cfg, bellOut, seed)
	fmt.Fprint(output, "\x1b[2J\x1b[H")

	p := tea.NewProgram(
		m,
		tea.WithContext(ctx),
		tea.WithInput(input),
		tea.WithOutput(output),
	)
	finalModel, err := p.Run()
	if err != nil {
		return Result{}, err
	}
	fm, ok := finalModel.(*defenseModel)
	if !ok {
		return Result{}, fmt.Errorf("defense: unexpected model type")
	}
	return fm.result(), nil
}

// PrintOutcome writes the post-game summary to out.
func PrintOutcome(out io.Writer, r Result) {
	fmt.Fprintln(out)
	if r.Aborted {
		fmt.Fprintln(out, "Defense — aborted.")
		return
	}
	fmt.Fprintln(out, "Defense — Game Over")
	sec := int(r.Elapsed.Round(time.Second).Seconds())
	if sec < 0 {
		sec = 0
	}
	fmt.Fprintf(out, "  Score: %d (%d words + %ds survival)\n", r.DisplayScore(), r.Score, sec)
	fmt.Fprintf(out, "  Time:  %s\n", formatDuration(r.Elapsed))
	fmt.Fprintf(out, "  Lives remaining: %d\n", r.Lives)
}
