package session

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"

	"typer/internal/model"
	"typer/internal/ui"
)

// quoteFetchSplashMessages are shown next to the spinner while remote quotes load (random pick).
var quoteFetchSplashMessages = []string{
	"Borrowing words from the internet…",
	"Convincing APIs to share…",
	"Fetching fresh quotes…",
	"Hang tight — grabbing text…",
	"Hold on — downloading wisdom…",
	"Negotiating with quote servers…",
	"One moment — lining up letters…",
	"Snagging something to type…",
	"Summoning the next passage…",
	"The tubes are warming up…",
	"Tracking down good lines…",
	"Waiting on the wire…",
	"Warming up the quote pipes…",
}

func randomQuoteFetchSplashMessage() string {
	return quoteFetchSplashMessages[rand.IntN(len(quoteFetchSplashMessages))]
}

// runQuoteFetchSplash runs gen (typically Provider.Next) in the background and draws a spinner + message on out until it finishes.
func runQuoteFetchSplash(ctx context.Context, out io.Writer, gen func() (model.Prompt, error)) (model.Prompt, error) {
	type res struct {
		p model.Prompt
		e error
	}
	ch := make(chan res, 1)
	go func() {
		p, e := gen()
		ch <- res{p, e}
	}()

	msg := randomQuoteFetchSplashMessage()
	frames := spinner.Dot.Frames
	spinStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorActiveFg))
	tick := time.NewTicker(spinner.Dot.FPS)
	defer tick.Stop()

	fmt.Fprint(out, "\x1b[?25l")
	defer fmt.Fprint(out, "\x1b[?25h")

	frame := 0
	draw := func() {
		f := frames[frame%len(frames)]
		frame++
		line := fmt.Sprintf("%s %s", spinStyle.Render(f), msg)
		fmt.Fprintf(out, "\r\x1b[2K%s", line)
	}

	draw()

	for {
		select {
		case r := <-ch:
			fmt.Fprint(out, "\r\x1b[2K")
			return r.p, r.e
		case <-ctx.Done():
			fmt.Fprint(out, "\r\x1b[2K")
			return model.Prompt{}, ctx.Err()
		case <-tick.C:
			draw()
		}
	}
}

// NextQuotePrompt loads one quote, optionally showing the remote fetch splash on out.
func NextQuotePrompt(ctx context.Context, out io.Writer, gen func() (model.Prompt, error), remoteSplash bool) (model.Prompt, error) {
	if remoteSplash {
		return runQuoteFetchSplash(ctx, out, gen)
	}
	return gen()
}
