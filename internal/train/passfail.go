package train

import (
	"fmt"
	"slices"
	"time"

	"typer/internal/model"
	"typer/internal/scoring"
)

type LessonOutcome struct {
	Passed     bool
	LessonID   string
	NextLesson string
	Message    string
	Tip        string
}

// KeystrokeAccuracy returns correct/total keystrokes as a percentage, or -1 when unavailable.
func KeystrokeAccuracy(result model.SessionResult) float64 {
	if result.TotalKeystrokes <= 0 {
		return -1
	}
	return scoring.Round2(100 * float64(result.CorrectKeystrokes) / float64(result.TotalKeystrokes))
}

// lessonAccuracy picks the accuracy metric used for pass/fail.
// Strict lessons use keystroke accuracy; others use final text accuracy.
func lessonAccuracy(lesson Lesson, result model.SessionResult) float64 {
	if lesson.Strict {
		if ks := KeystrokeAccuracy(result); ks >= 0 {
			return ks
		}
	}
	return result.Metrics.Accuracy
}

// EvaluateLesson checks session metrics against tier thresholds and updates profile progress.
func EvaluateLesson(p *Profile, c *Curriculum, lesson Lesson, result model.SessionResult) LessonOutcome {
	out := LessonOutcome{LessonID: lesson.ID}

	tierCfg, ok := c.Tier(lesson.Tier)
	if !ok {
		tierCfg = TierConfig{WPMFloor: 15, AccFloor: 90}
	}

	wpmFloor := tierCfg.WPMFloor
	accFloor := tierCfg.AccFloor
	if p.Placement.NetWPM > 0 {
		wpmFloor = adjustFloor(wpmFloor, p.Placement.NetWPM)
	}

	netWPM := result.Metrics.NetWPM
	accuracy := lessonAccuracy(lesson, result)
	passed := netWPM >= wpmFloor && accuracy >= accFloor

	if lesson.Assessment {
		passed = netWPM >= wpmFloor*0.9 && accuracy >= accFloor-2
	}

	topErrs := TopSessionErrorChars(result, 1)
	var tip string
	if len(topErrs) > 0 && topErrs[0].Count > 0 {
		tip = fmt.Sprintf("practice %s — %d error(s) this session", topErrs[0].Key, topErrs[0].Count)
	}

	if passed {
		out.Passed = true
		if !slices.Contains(p.Progress.CompletedLessons, lesson.ID) {
			p.Progress.CompletedLessons = append(p.Progress.CompletedLessons, lesson.ID)
		}
		next := c.NextLessonID(lesson.ID)
		lines := lesson.EffectiveRounds()
		if next != "" {
			nextLesson, _ := c.Lesson(next)
			p.Progress.CurrentLesson = next
			p.Progress.CurrentTier = nextLesson.Tier
			out.NextLesson = next
			out.Message = formatPassMessage(lesson, lines, accuracy, netWPM)
		} else if c.CurriculumComplete(*p) || c.TierComplete(*p, TierFluent) {
			p.Progress.CurrentTier = TierAdaptive
			p.Progress.CurrentLesson = ""
			out.Message = formatPassMessage(lesson, lines, accuracy, netWPM) + " — curriculum finished"
		} else if nextID := c.FirstIncompleteLesson(*p); nextID != "" {
			nextLesson, _ := c.Lesson(nextID)
			p.Progress.CurrentLesson = nextID
			p.Progress.CurrentTier = nextLesson.Tier
			out.NextLesson = nextID
			out.Message = formatPassMessage(lesson, lines, accuracy, netWPM)
		} else {
			out.Message = formatPassMessage(lesson, lines, accuracy, netWPM)
		}
		out.Tip = tip
		UpdateStreak(p)
		return out
	}

	out.Passed = false
	out.Message = formatFailMessage(lesson, lesson.EffectiveRounds(), accuracy, netWPM, wpmFloor, accFloor)
	out.Tip = tip

	if lesson.Assessment && lesson.ReviewOnFail != "" {
		p.Progress.CurrentLesson = lesson.ReviewOnFail
		if rl, ok := c.Lesson(lesson.ReviewOnFail); ok {
			p.Progress.CurrentTier = rl.Tier
		}
		out.Message = fmt.Sprintf("Assessment %s not passed — review lesson %s", lesson.ID, lesson.ReviewOnFail)
	}
	return out
}

func accuracyLabel(strict bool) string {
	if strict {
		return "keystroke acc"
	}
	return "acc"
}

func formatPassMessage(lesson Lesson, lines int, accuracy, netWPM float64) string {
	label := accuracyLabel(lesson.Strict)
	if lines <= 1 {
		return fmt.Sprintf("Lesson %s complete — passed (%.0f%% %s, %.0f net WPM)", lesson.ID, accuracy, label, netWPM)
	}
	return fmt.Sprintf("Lesson %s complete — passed (%d lines, %.0f%% %s, %.0f net WPM)", lesson.ID, lines, accuracy, label, netWPM)
}

func formatFailMessage(lesson Lesson, lines int, accuracy, netWPM, wpmFloor, accFloor float64) string {
	label := accuracyLabel(lesson.Strict)
	if lesson.Assessment {
		accFloor -= 2
		wpmFloor *= 0.9
	}
	if lines <= 1 {
		return fmt.Sprintf("Lesson %s — retry needed (%.0f%% %s, %.0f net WPM; need %.0f WPM, %.0f%% %s)", lesson.ID, accuracy, label, netWPM, wpmFloor, accFloor, label)
	}
	return fmt.Sprintf("Lesson %s — retry needed (%d lines, %.0f%% %s, %.0f net WPM; need %.0f WPM, %.0f%% %s)", lesson.ID, lines, accuracy, label, netWPM, wpmFloor, accFloor, label)
}

func adjustFloor(base, placementWPM float64) float64 {
	if placementWPM < base {
		return base * 0.85
	}
	return base
}

func UpdateStreak(p *Profile) {
	today := time.Now().UTC().Format("2006-01-02")
	if p.Progress.LastPracticeDate == today {
		return
	}
	yesterday := time.Now().UTC().Add(-24 * time.Hour).Format("2006-01-02")
	if p.Progress.LastPracticeDate == yesterday {
		p.Progress.StreakDays++
	} else {
		p.Progress.StreakDays = 1
	}
	p.Progress.LastPracticeDate = today
}
