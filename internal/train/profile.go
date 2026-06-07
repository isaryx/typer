package train

import "time"

const ProfileVersion = 1

const (
	TierFoundation = "foundation"
	TierBuilding   = "building"
	TierFluent     = "fluent"
	TierAdaptive   = "adaptive"
)

type KeyStat struct {
	Attempts int `json:"attempts"`
	Errors   int `json:"errors"`
}

type PlacementResult struct {
	NetWPM         float64 `json:"net_wpm"`
	Accuracy       float64 `json:"accuracy"`
	AssignedTier   string  `json:"assigned_tier"`
	AssignedLesson string  `json:"assigned_lesson"`
}

type Progress struct {
	CurrentTier      string   `json:"current_tier"`
	CurrentLesson    string   `json:"current_lesson"`
	CompletedLessons []string `json:"completed_lessons,omitempty"`
	StreakDays       int      `json:"streak_days"`
	LastPracticeDate string   `json:"last_practice_date,omitempty"` // YYYY-MM-DD
}

type Profile struct {
	Version     int                `json:"version"`
	CreatedAt   time.Time          `json:"created_at"`
	EvaluatedAt time.Time          `json:"evaluated_at,omitempty"`
	Placement   PlacementResult    `json:"placement"`
	Progress    Progress           `json:"progress"`
	Keys        map[string]KeyStat `json:"keys,omitempty"`
	WeakKeys    []string           `json:"weak_keys,omitempty"`
}

func NewProfile(placement PlacementResult) Profile {
	now := time.Now().UTC()
	return Profile{
		Version:     ProfileVersion,
		CreatedAt:   now,
		EvaluatedAt: now,
		Placement:   placement,
		Progress: Progress{
			CurrentTier:   placement.AssignedTier,
			CurrentLesson: placement.AssignedLesson,
		},
		Keys: make(map[string]KeyStat),
	}
}

func (p *Profile) IsEmpty() bool {
	return p.Version == 0 && p.CreatedAt.IsZero()
}
