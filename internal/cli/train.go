package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"typer/internal/model"
	"typer/internal/session"
	"typer/internal/storage"
	"typer/internal/text"
	"typer/internal/train"
)

var newProfileStore = train.NewProfileStore

const noProfileMessage = "No training profile yet. Run `typer train -e` to take the placement test."

func runTrain(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) > 0 {
		switch args[0] {
		case "status":
			return runTrainStatus(args[1:], stdout)
		case "reset":
			return runTrainReset(args[1:], stdin, stdout)
		case "--help", "-h", "help":
			printTrainHelp(stdout)
			return nil
		}
	}

	fs := flag.NewFlagSet("train", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	evaluate := fs.Bool("evaluate", false, "Run placement test and create/update profile.")
	evalShort := fs.Bool("e", false, "Shorthand for --evaluate.")
	listLessons := fs.Bool("list", false, "List curriculum and completion status.")
	lessonID := fs.String("lesson", "", "Jump to a specific lesson (if unlocked).")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printTrainHelp(stdout)
			return nil
		}
		return usageErrf("%v", err)
	}
	if err := rejectExtraArgs("train", fs.Args()); err != nil {
		return err
	}

	if *listLessons && (*evaluate || *evalShort) {
		return usageErrf("--list cannot be combined with --evaluate")
	}

	if *listLessons {
		return runTrainList(stdout)
	}

	if *evaluate || *evalShort {
		return runTrainEvaluate(ctx, stdin, stdout)
	}

	return runTrainSession(ctx, stdin, stdout, strings.TrimSpace(*lessonID))
}

func runTrainStatus(args []string, stdout io.Writer) error {
	if err := rejectExtraArgs("train status", args); err != nil {
		return err
	}
	store, err := newProfileStore()
	if err != nil {
		return err
	}
	p, err := store.Load()
	if err != nil {
		if errors.Is(err, train.ErrNoProfile) {
			fmt.Fprintln(stdout, noProfileMessage)
			return nil
		}
		return err
	}
	curriculum, err := train.LoadCurriculum()
	if err != nil {
		return err
	}
	printTrainProfile(stdout, p, curriculum)
	return nil
}

func runTrainReset(args []string, stdin io.Reader, stdout io.Writer) error {
	if err := rejectExtraArgs("train reset", args); err != nil {
		return err
	}
	fmt.Fprint(stdout, "Reset training profile? Session history is kept. [y/N]: ")
	ok, err := readResetConfirm(stdin)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(stdout, "Reset cancelled.")
		return nil
	}
	store, err := newProfileStore()
	if err != nil {
		return err
	}
	if err := store.Reset(); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Training profile cleared.")
	return nil
}

func runTrainList(stdout io.Writer) error {
	store, err := newProfileStore()
	if err != nil {
		return err
	}
	p, err := store.Load()
	hasProfile := err == nil
	if err != nil && !errors.Is(err, train.ErrNoProfile) {
		return err
	}
	curriculum, err := train.LoadCurriculum()
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, "Training curriculum:")
	current := ""
	completed := map[string]bool{}
	if hasProfile {
		current = p.Progress.CurrentLesson
		for _, id := range p.Progress.CompletedLessons {
			completed[id] = true
		}
	}
	var lastTier string
	for i, l := range curriculum.AllLessons() {
		if l.Tier != lastTier {
			lastTier = l.Tier
			fmt.Fprintf(stdout, "\n%s\n", curriculum.FormatTierName(l.Tier))
		}
		marker := " "
		switch {
		case completed[l.ID]:
			marker = "✓"
		case l.ID == current:
			marker = "→"
		case hasProfile && !curriculum.IsLessonUnlocked(p, l.ID):
			marker = "·"
		case !hasProfile && i > 0:
			marker = "·"
		}
		fmt.Fprintf(stdout, "  %s %s  %s\n", marker, l.ID, l.Title)
	}
	if !hasProfile {
		fmt.Fprintln(stdout, "\nNo profile yet — run `typer train -e` to start.")
	}
	return nil
}

func runTrainEvaluate(ctx context.Context, stdin io.Reader, stdout io.Writer) error {
	store, err := newProfileStore()
	if err != nil {
		return err
	}
	if _, err := store.Load(); err == nil {
		fmt.Fprint(stdout, "Re-run placement test? This replaces your training profile. [y/N]: ")
		ok, err := readResetConfirm(stdin)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(stdout, "Placement test cancelled.")
			return nil
		}
		fmt.Fprintln(stdout)
	} else if err != nil && !errors.Is(err, train.ErrNoProfile) {
		return err
	}

	curriculum, err := train.LoadCurriculum()
	if err != nil {
		return err
	}
	settingsStore, err := storage.NewSettingsStore()
	if err != nil {
		return err
	}
	settings, err := settingsStore.Load()
	if err != nil {
		return err
	}
	corpus, err := text.LoadWordCorpus(settings.WordsFile)
	if err != nil {
		return err
	}
	filter := train.NewWordFilter(corpus.Words)

	fmt.Fprintln(stdout, "Placement test — 3 segments (~2 min total).")
	fmt.Fprint(stdout, "Type until the timer ends or you finish the text.\n\n")

	var segments []model.SessionResult
	for i, seg := range curriculum.PlacementSegments() {
		fmt.Fprintf(stdout, "Segment %d/%d (%s)…\n", i+1, len(curriculum.PlacementSegments()), seg.ID)
		content, err := placementSegmentContent(seg, filter, settings.PassagesFile)
		if err != nil {
			return err
		}
		provider := text.NewStaticProvider(model.Prompt{
			ID:      "placement-" + seg.ID,
			Content: content,
			Source:  "placement",
			Mode:    model.ModeTrain,
		})
		opts := model.SessionOptions{
			Mode:       model.ModeTrain,
			DurationMS: seg.DurationMS,
			FingerHint: true,
			Strict:     seg.PlacementStrict(),
		}
		applySessionDisplayFromSettings(&opts, settings)

		runner := session.NewRunner(provider)
		var historyStore *storage.HistoryStore
		result, err := runSessionAndPersist(ctx, runner, opts, stdin, stdout, nil, &historyStore)
		if err != nil {
			return err
		}
		if result.Aborted {
			fmt.Fprintln(stdout, "\nPlacement test aborted.")
			return nil
		}
		segments = append(segments, result)
		fmt.Fprintf(stdout, "Segment %s: %.0f net WPM, %.0f%% accuracy\n\n", seg.ID, result.Metrics.NetWPM, train.SegmentAccuracy(result))
	}

	placement := train.AssignPlacement(train.PlacementInput{Segments: segments})
	profile := train.NewProfile(placement)
	for _, s := range segments {
		train.MergeSession(&profile, s)
	}
	train.FinalizePlacementProfile(&profile)

	if err := store.Save(profile); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "Placement complete.")
	printPlacementResults(stdout, curriculum, placement, profile, segments)
	fmt.Fprintln(stdout, "\nProfile saved. Run `typer train` when you're ready to start.")
	return nil
}

func printPlacementResults(out io.Writer, curriculum *train.Curriculum, placement train.PlacementResult, profile train.Profile, segments []model.SessionResult) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Segment results:")
	for i, s := range segments {
		id := "?"
		if i < len(curriculum.PlacementSegments()) {
			id = curriculum.PlacementSegments()[i].ID
		}
		fmt.Fprintf(out, "  %s: %.0f net WPM, %.0f%% accuracy\n", id, s.Metrics.NetWPM, train.SegmentAccuracy(s))
	}
	fmt.Fprintln(out)
	fmt.Fprintf(out, "  Level:     %s\n", curriculum.FormatTierName(placement.AssignedTier))
	if l, ok := curriculum.Lesson(placement.AssignedLesson); ok {
		fmt.Fprintf(out, "  Start at:  %s\n", train.LessonLabel(l))
	} else {
		fmt.Fprintf(out, "  Start at:  lesson %s\n", placement.AssignedLesson)
	}
	fmt.Fprintf(out, "  Net WPM:   %.0f\n", placement.NetWPM)
	fmt.Fprintf(out, "  Accuracy:  %.0f%%\n", placement.Accuracy)
	if len(profile.WeakKeys) > 0 {
		fmt.Fprintf(out, "  Weak keys: %s\n", strings.Join(profile.WeakKeys, ", "))
	}
}

func runTrainSession(ctx context.Context, stdin io.Reader, stdout io.Writer, lessonOverride string) error {
	store, err := newProfileStore()
	if err != nil {
		return err
	}
	profile, err := store.Load()
	if err != nil {
		if errors.Is(err, train.ErrNoProfile) {
			fmt.Fprintln(stdout, noProfileMessage)
			return nil
		}
		return err
	}

	curriculum, err := train.LoadCurriculum()
	if err != nil {
		return err
	}

	settingsStore, err := storage.NewSettingsStore()
	if err != nil {
		return err
	}
	settings, err := settingsStore.Load()
	if err != nil {
		return err
	}
	corpus, err := text.LoadWordCorpus(settings.WordsFile)
	if err != nil {
		return err
	}
	filter := train.NewWordFilter(corpus.Words)

	var anchor progressAnchor
	if lessonOverride != "" {
		if !curriculum.IsLessonUnlocked(profile, lessonOverride) {
			return usageErrf("lesson %q is not unlocked yet", lessonOverride)
		}
		anchor = progressAnchor{
			lesson: profile.Progress.CurrentLesson,
			tier:   profile.Progress.CurrentTier,
			active: true,
		}
		profile.Progress.CurrentLesson = lessonOverride
		if l, ok := curriculum.Lesson(lessonOverride); ok {
			profile.Progress.CurrentTier = l.Tier
		}
	}

	return runTrainLoop(ctx, stdin, stdout, store, profile, curriculum, settings, filter, anchor)
}

type progressAnchor struct {
	lesson string
	tier   string
	active bool
}

func (a progressAnchor) restore(p *train.Profile) {
	if !a.active {
		return
	}
	p.Progress.CurrentLesson = a.lesson
	p.Progress.CurrentTier = a.tier
}

func runTrainLoop(
	ctx context.Context,
	stdin io.Reader,
	stdout io.Writer,
	store *train.ProfileStore,
	profile train.Profile,
	curriculum *train.Curriculum,
	settings storage.AppSettings,
	filter *train.WordFilter,
	anchor progressAnchor,
) error {
	var historyStore *storage.HistoryStore
	var lastHeaderLesson string

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		adaptive := profile.Progress.CurrentTier == train.TierAdaptive || curriculum.CurriculumComplete(profile)

		if adaptive {
			fmt.Fprintln(stdout, "Adaptive training — targeting weak keys.")
			if len(profile.WeakKeys) > 0 {
				fmt.Fprintf(stdout, "Focus: %s\n", strings.Join(profile.WeakKeys, ", "))
			}
			fmt.Fprintln(stdout)

			provider := text.NewAdaptiveProvider(filter, profile.WeakKeys, 20)
			if note := provider.LastNote(); note != "" {
				fmt.Fprintf(stdout, "Note: %s\n", note)
			}
			opts := model.SessionOptions{Mode: model.ModeTrain, LessonID: "adaptive"}
			applySessionDisplayFromSettings(&opts, settings)
			runner := session.NewRunner(provider)

			result, err := runSessionAndPersist(ctx, runner, opts, stdin, stdout, nil, &historyStore)
			if err != nil {
				return err
			}
			if result.Aborted {
				fmt.Fprintln(stdout, "\nTraining stopped.")
				return nil
			}

			train.MergeSession(&profile, result)
			train.UpdateStreak(&profile)
			anchor.restore(&profile)
			if err := store.Save(profile); err != nil {
				return err
			}

			fmt.Fprintf(stdout, "\nAdaptive session — %.0f net WPM, %.0f%% accuracy\n", result.Metrics.NetWPM, result.Metrics.Accuracy)
			if len(profile.WeakKeys) > 0 {
				fmt.Fprintf(stdout, "Weak keys: %s\n", strings.Join(profile.WeakKeys, ", "))
			}

			action, err := promptTrainContinue(stdin, stdout, train.LessonOutcome{}, true)
			if err != nil {
				return err
			}
			if action == trainContinueStop {
				fmt.Fprintln(stdout, "Training stopped.")
				return nil
			}
			continue
		}

		lessonID := profile.Progress.CurrentLesson
		lesson, ok := curriculum.Lesson(lessonID)
		if !ok {
			return usageErrf("unknown lesson %q", lessonID)
		}

		if lesson.ID != lastHeaderLesson {
			printLessonHeader(stdout, lesson, profile, curriculum)
			lastHeaderLesson = lesson.ID
		}

		opts := lessonSessionOptions(lesson)
		applySessionDisplayFromSettings(&opts, settings)
		provider := text.NewTrainProvider(filter, lesson)
		runner := session.NewRunner(provider)

		result, err := runSessionAndPersist(ctx, runner, opts, stdin, stdout, nil, &historyStore)
		if err != nil {
			return err
		}
		if result.Aborted {
			fmt.Fprintln(stdout, "\nTraining stopped.")
			return nil
		}

		finishedLesson := lesson
		train.MergeSession(&profile, result)
		outcome := train.EvaluateLesson(&profile, curriculum, finishedLesson, result)
		anchor.restore(&profile)
		if err := store.Save(profile); err != nil {
			return err
		}

		fmt.Fprintf(stdout, "\n%s\n", outcome.Message)
		if outcome.NextLesson != "" {
			if next, ok := curriculum.Lesson(outcome.NextLesson); ok {
				fmt.Fprintf(stdout, "Next: %s\n", train.LessonLabel(next))
			}
		} else if profile.Progress.CurrentTier == train.TierAdaptive {
			fmt.Fprintln(stdout, "Curriculum complete — adaptive drills unlocked.")
		}
		if outcome.Tip != "" {
			fmt.Fprintf(stdout, "Tip: %s\n", outcome.Tip)
		}

		action, err := promptTrainContinue(stdin, stdout, outcome, false)
		if err != nil {
			return err
		}
		if action == trainContinueStop {
			fmt.Fprintln(stdout, "Training stopped.")
			return nil
		}
		if action == trainContinueRetry {
			profile.Progress.CurrentLesson = finishedLesson.ID
			profile.Progress.CurrentTier = finishedLesson.Tier
			lastHeaderLesson = ""
			continue
		}

		// Enter → next lesson (profile already advanced on pass).
		if outcome.Passed {
			lastHeaderLesson = ""
		}
	}
}

func printLessonHeader(stdout io.Writer, lesson train.Lesson, profile train.Profile, curriculum *train.Curriculum) {
	fmt.Fprintf(stdout, "Training: %s [%s]\n", train.LessonLabel(lesson), curriculum.FormatTierName(lesson.Tier))
	if lines := lesson.EffectiveRounds(); lines > 1 {
		fmt.Fprintf(stdout, "Lines: %d (one session)\n", lines)
	}
	if profile.Progress.StreakDays > 0 {
		fmt.Fprintf(stdout, "Streak: %d day(s)\n", profile.Progress.StreakDays)
	}
	fmt.Fprintln(stdout)
}

func lessonSessionOptions(lesson train.Lesson) model.SessionOptions {
	opts := model.SessionOptions{
		Mode:       model.ModeTrain,
		LessonID:   lesson.ID,
		Strict:     lesson.Strict,
		FingerHint: lesson.FingerHint,
		DurationMS: lesson.TimedMS,
	}
	return opts
}

func placementSegmentContent(seg train.PlacementSegment, filter *train.WordFilter, passagesFile string) (string, error) {
	if seg.Content != "" {
		return seg.Content, nil
	}
	if seg.PromptWords > 0 {
		words, _ := filter.WordsContaining(nil, seg.PromptWords)
		if len(words) == 0 {
			return "", fmt.Errorf("no words for placement segment %s", seg.ID)
		}
		return strings.Join(words, " "), nil
	}
	if seg.UsePassage {
		provider, err := text.NewLocalProvider(passagesFile)
		if err != nil {
			return "", err
		}
		p, err := provider.Next(context.Background(), text.Constraints{})
		if err != nil {
			return "", err
		}
		return p.Content, nil
	}
	return "", fmt.Errorf("placement segment %s has no content", seg.ID)
}

func printTrainProfile(out io.Writer, p train.Profile, c *train.Curriculum) {
	fmt.Fprintln(out, "Training profile")
	fmt.Fprintf(out, "  Tier:    %s\n", c.FormatTierName(p.Progress.CurrentTier))
	if p.Progress.CurrentLesson != "" {
		if l, ok := c.Lesson(p.Progress.CurrentLesson); ok {
			fmt.Fprintf(out, "  Lesson:  %s\n", train.LessonLabel(l))
		} else {
			fmt.Fprintf(out, "  Lesson:  %s\n", p.Progress.CurrentLesson)
		}
	} else if p.Progress.CurrentTier == train.TierAdaptive {
		fmt.Fprintln(out, "  Mode:    Adaptive drills")
	}
	fmt.Fprintf(out, "  Streak:  %d day(s)\n", p.Progress.StreakDays)
	if !p.EvaluatedAt.IsZero() {
		fmt.Fprintf(out, "  Placed:  %.0f net WPM, %.0f%% accuracy → %s / %s\n",
			p.Placement.NetWPM, p.Placement.Accuracy,
			c.FormatTierName(p.Placement.AssignedTier), p.Placement.AssignedLesson)
	}
	fmt.Fprintf(out, "  Done:    %d lesson(s)\n", len(p.Progress.CompletedLessons))
	if len(p.WeakKeys) > 0 {
		fmt.Fprintf(out, "  Weak keys: %s\n", strings.Join(p.WeakKeys, ", "))
	}
}
