package train

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"typer/assets"
)

type TierConfig struct {
	Label    string  `json:"label"`
	WPMFloor float64 `json:"wpm_floor"`
	AccFloor float64 `json:"acc_floor"`
}

type Lesson struct {
	ID          string   `json:"id"`
	Tier        string   `json:"tier"`
	Title       string   `json:"title"`
	Keys        []string `json:"keys,omitempty"`
	Prompts     []string `json:"prompts,omitempty"`
	PromptWords int      `json:"prompt_words,omitempty"`
	Strict      bool     `json:"strict"`
	FingerHint  bool     `json:"finger_hint"`
	TimedMS      int      `json:"timed_ms,omitempty"`
	Rounds       int      `json:"rounds,omitempty"`
	Assessment   bool     `json:"assessment,omitempty"`
	ReviewOnFail string  `json:"review_on_fail,omitempty"`
}

type PlacementSegment struct {
	ID          string `json:"id"`
	DurationMS  int    `json:"duration_ms"`
	Content     string `json:"content,omitempty"`
	PromptWords int    `json:"prompt_words,omitempty"`
	UsePassage  bool   `json:"use_passage,omitempty"`
	Strict      bool   `json:"strict,omitempty"`
}

// PlacementStrict reports whether trace-based weak-key stats are collected during placement.
func (s PlacementSegment) PlacementStrict() bool {
	return s.Strict
}

type CurriculumFile struct {
	Tiers       map[string]TierConfig `json:"tiers"`
	Lessons     []Lesson              `json:"lessons"`
	Placement   PlacementConfig       `json:"placement"`
	Assessments []string              `json:"assessments"`
}

type PlacementConfig struct {
	Segments []PlacementSegment `json:"segments"`
}

type Curriculum struct {
	file CurriculumFile
	byID map[string]Lesson
	order []string
}

func LoadCurriculum() (*Curriculum, error) {
	var file CurriculumFile
	if err := json.Unmarshal(assets.Lessons, &file); err != nil {
		return nil, fmt.Errorf("parse lessons.json: %w", err)
	}
	c := &Curriculum{
		file:  file,
		byID:  make(map[string]Lesson, len(file.Lessons)),
		order: make([]string, 0, len(file.Lessons)),
	}
	for _, l := range file.Lessons {
		if _, ok := c.byID[l.ID]; ok {
			return nil, fmt.Errorf("duplicate lesson id %q", l.ID)
		}
		c.byID[l.ID] = l
		c.order = append(c.order, l.ID)
	}
	return c, nil
}

func (c *Curriculum) Tier(id string) (TierConfig, bool) {
	t, ok := c.file.Tiers[id]
	return t, ok
}

func (c *Curriculum) Lesson(id string) (Lesson, bool) {
	l, ok := c.byID[id]
	return l, ok
}

func (c *Curriculum) PlacementSegments() []PlacementSegment {
	return c.file.Placement.Segments
}

func (c *Curriculum) FirstIncompleteLesson(p Profile) string {
	for _, id := range c.order {
		l := c.byID[id]
		if l.Tier == TierAdaptive {
			continue
		}
		if !slices.Contains(p.Progress.CompletedLessons, id) {
			return id
		}
	}
	return ""
}

func (c *Curriculum) NextLessonID(current string) string {
	for i, id := range c.order {
		if id == current && i+1 < len(c.order) {
			next := c.byID[c.order[i+1]]
			if next.Tier == TierAdaptive {
				return ""
			}
			return c.order[i+1]
		}
	}
	return ""
}

func (c *Curriculum) AllLessons() []Lesson {
	out := make([]Lesson, 0, len(c.order))
	for _, id := range c.order {
		out = append(out, c.byID[id])
	}
	return out
}

func (c *Curriculum) IsLessonUnlocked(p Profile, lessonID string) bool {
	if lessonID == p.Progress.CurrentLesson {
		return true
	}
	if slices.Contains(p.Progress.CompletedLessons, lessonID) {
		return true
	}
	idx := slices.Index(c.order, lessonID)
	if idx <= 0 {
		return lessonID == c.order[0]
	}
	prev := c.order[idx-1]
	return slices.Contains(p.Progress.CompletedLessons, prev)
}

func (c *Curriculum) TierComplete(p Profile, tier string) bool {
	for _, l := range c.AllLessons() {
		if l.Tier != tier {
			continue
		}
		if !slices.Contains(p.Progress.CompletedLessons, l.ID) {
			return false
		}
	}
	return true
}

func (c *Curriculum) CurriculumComplete(p Profile) bool {
	for _, l := range c.AllLessons() {
		if l.Tier == TierAdaptive {
			continue
		}
		if !slices.Contains(p.Progress.CompletedLessons, l.ID) {
			return false
		}
	}
	return len(c.AllLessons()) > 0
}

// LessonLabel returns a display string for a lesson.
func LessonLabel(l Lesson) string {
	return fmt.Sprintf("%s — %s", l.ID, l.Title)
}

// FormatTierName returns the human label for a tier id.
func (c *Curriculum) FormatTierName(tier string) string {
	if t, ok := c.file.Tiers[tier]; ok && t.Label != "" {
		return t.Label
	}
	if tier == "" {
		return tier
	}
	return strings.ToUpper(tier[:1]) + tier[1:]
}
