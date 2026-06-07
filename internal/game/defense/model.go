package defense

import (
	"fmt"
	"io"
	"math/rand/v2"
	"time"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"typer/internal/session"
)

type tickMsg struct {
	now time.Time
}

// Result is the outcome after the TUI exits.
type Result struct {
	Score     int
	Lives     int
	Elapsed   time.Duration
	StartedAt time.Time
	EndedAt   time.Time
	Over      bool
	Aborted   bool
}

type defenseModel struct {
	cfg    Config
	pool   WordPool
	words  []Word
	styles session.Styles
	rng    *rand.Rand
	width  int
	height int
	tooSmall bool

	lives  int
	score  int
	lockID int
	typed  string
	nextID int

	lastSpawnedText string

	startedAt time.Time
	lastTick  time.Time
	spawnWait float64

	over    bool
	aborted bool
	endedAt time.Time
	bellOut io.Writer

	baseHitFlashUntil time.Time
}

func newDefenseModel(pool WordPool, cfg Config, bellOut io.Writer, seed uint64) *defenseModel {
	wp := pool
	if len(wp.all) == 0 {
		wp = NewWordPool([]string{"code", "type", "word", "cat", "byte", "the"})
	}
	seed = resolveSeed(seed)
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	wp.Shuffle(rng)
	return &defenseModel{
		cfg:     cfg,
		pool:    wp,
		styles:  session.DefaultStyles(),
		rng:     rng,
		lives:   cfg.Lives,
		lockID:  0,
		bellOut: bellOut,
		width:   80,
		height:  24,
	}
}

func (m *defenseModel) Init() tea.Cmd {
	now := time.Now()
	m.startedAt = now
	m.lastTick = now
	return m.scheduleTick()
}

func (m *defenseModel) scheduleTick() tea.Cmd {
	return tea.Tick(TickInterval, func(t time.Time) tea.Msg {
		return tickMsg{now: t}
	})
}

func (m *defenseModel) elapsed(now time.Time) time.Duration {
	if m.startedAt.IsZero() {
		return 0
	}
	return now.Sub(m.startedAt)
}

func (m *defenseModel) clampWordsToWidth(innerWidth int) {
	for i := range m.words {
		wlen := utf8.RuneCountInString(m.words[i].Text)
		m.words[i].Col = clampWordCol(m.words[i].Col, wlen, innerWidth)
	}
}

func (m *defenseModel) applyTick(now time.Time) {
	if m.over || m.tooSmall {
		return
	}

	dt := now.Sub(m.lastTick).Seconds()
	if dt <= 0 {
		dt = TickInterval.Seconds()
	}
	m.lastTick = now

	elapsed := m.elapsed(now)
	speed := EffectiveFallSpeed(m.cfg.BaseFallSpeed, elapsed)
	interval := EffectiveSpawnInterval(m.cfg.BaseSpawnSeconds, elapsed)

	lifeLostThisTick := false
	var survivors []Word
	for _, w := range m.words {
		w.Row += speed * dt
		if w.Row >= float64(ShieldRow()) {
			if !lifeLostThisTick {
				m.loseLife(now)
				lifeLostThisTick = true
			}
			if w.ID == m.lockID {
				m.clearLock()
			}
			continue
		}
		survivors = append(survivors, w)
	}
	m.words = survivors

	if m.over {
		return
	}

	m.spawnWait += dt
	for m.spawnWait >= interval && len(m.words) < MaxConcurrentWords && !m.over {
		inner := m.innerWidth()
		w, ok := spawnWord(m.pool, inner, m.rng, m.nextID+1, m.words, m.score, m.lastSpawnedText)
		if !ok {
			break
		}
		m.spawnWait -= interval
		m.nextID = w.ID
		m.lastSpawnedText = w.Text
		m.words = append(m.words, w)
		interval = EffectiveSpawnInterval(m.cfg.BaseSpawnSeconds, m.elapsed(now))
	}
}

func (m *defenseModel) baseHitFlashing(now time.Time) bool {
	return !m.baseHitFlashUntil.IsZero() && now.Before(m.baseHitFlashUntil)
}

func (m *defenseModel) loseLife(now time.Time) {
	if m.lives <= 0 {
		return
	}
	m.baseHitFlashUntil = now.Add(BaseHitFlashDuration)
	m.lives--
	if m.lives <= 0 {
		m.over = true
		m.endedAt = now
	}
}

func (m *defenseModel) clearLock() {
	m.lockID = 0
	m.typed = ""
}

func (m *defenseModel) ringBell() {
	if m.bellOut != nil {
		fmt.Fprint(m.bellOut, "\a")
	}
}

func (m *defenseModel) handleRune(r rune) {
	if m.over || m.tooSmall {
		return
	}
	if r >= 'A' && r <= 'Z' {
		r = unicode.ToLower(r)
	}
	if r < 'a' || r > 'z' {
		m.ringBell()
		return
	}

	if m.lockID == 0 {
		id, ok := selectLockCandidate(m.words, r, "")
		if !ok {
			m.ringBell()
			return
		}
		m.lockID = id
		m.typed = string(r)
		m.tryCompleteLocked()
		return
	}

	w, _, ok := wordByID(m.words, m.lockID)
	if !ok {
		m.clearLock()
		return
	}
	newTyped, ok := strictAppendTyped(w.Text, m.typed, r)
	if !ok {
		m.ringBell()
		return
	}
	m.typed = newTyped
	m.tryCompleteLocked()
}

func (m *defenseModel) tryCompleteLocked() {
	w, idx, ok := wordByID(m.words, m.lockID)
	if !ok {
		m.clearLock()
		return
	}
	if m.typed != w.Text {
		return
	}
	m.words = append(m.words[:idx], m.words[idx+1:]...)
	m.score++
	m.clearLock()
}

func (m *defenseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if m.over {
			return m, tea.Quit
		}
		m.applyTick(msg.now)
		if m.over {
			return m, tea.Quit
		}
		return m, m.scheduleTick()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tooSmall = m.width < MinTerminalWidth || m.height < MinTerminalHeight
		if !m.tooSmall {
			m.clampWordsToWidth(m.innerWidth())
		}
		return m, nil
	case tea.PasteMsg:
		m.ringBell()
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			m.aborted = true
			m.endedAt = time.Now()
			return m, tea.Quit
		case "esc", "tab":
			m.clearLock()
			return m, nil
		case "backspace":
			return m, nil
		}
		if len(msg.Text) == 1 {
			r, _ := utf8.DecodeRuneInString(msg.Text)
			m.handleRune(r)
		}
		return m, nil
	}
	return m, nil
}

func (m *defenseModel) View() tea.View {
	return tea.NewView(m.renderView())
}

func (m *defenseModel) result() Result {
	ended := m.endedAt
	if ended.IsZero() {
		ended = time.Now()
	}
	elapsed := m.elapsed(ended)
	if elapsed < 0 {
		elapsed = 0
	}
	return Result{
		Score:     m.score,
		Lives:     m.lives,
		Elapsed:   elapsed,
		StartedAt: m.startedAt,
		EndedAt:   ended,
		Over:      m.over,
		Aborted:   m.aborted,
	}
}

// ValidateConfig checks defense CLI parameters.
func ValidateConfig(c Config) error {
	if c.Lives < 1 {
		return fmt.Errorf("--lives must be at least 1 (got %d)", c.Lives)
	}
	if c.BaseSpawnSeconds <= 0 {
		return fmt.Errorf("--spawn-rate must be positive (got %g)", c.BaseSpawnSeconds)
	}
	if c.BaseFallSpeed <= 0 {
		return fmt.Errorf("--fall-speed must be positive (got %g)", c.BaseFallSpeed)
	}
	return nil
}

func formatDuration(d time.Duration) string {
	sec := int(d.Round(time.Second).Seconds())
	if sec < 60 {
		return fmt.Sprintf("%ds", sec)
	}
	min := sec / 60
	s := sec % 60
	return fmt.Sprintf("%dm %ds", min, s)
}

func formatLivesLabel(lives, max int) string {
	if max < 1 {
		max = DefaultLives
	}
	return fmt.Sprintf("♥  %d/%d", lives, max)
}

func formatCaptionHalves(m *defenseModel) (left, right string) {
	elapsed := m.elapsed(time.Now())
	r := Result{Score: m.score, Elapsed: elapsed}
	display := r.DisplayScore()
	sec := int(elapsed.Round(time.Second).Seconds())
	if sec < 0 {
		sec = 0
	}
	left = fmt.Sprintf("defense · %d (%dw+%ds)", display, m.score, sec)
	right = formatLivesLabel(m.lives, m.cfg.Lives)
	return left, right
}
