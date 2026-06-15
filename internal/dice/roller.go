package dice

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Result holds the outcome of a single roll expression.
type Result struct {
	Expression string    `json:"expression"` // e.g. "2d6+3"
	Rolls      []int     `json:"rolls"`      // individual die results
	Modifier   int       `json:"modifier"`
	Total      int       `json:"total"`
	Timestamp  time.Time `json:"timestamp"`
	Context    string    `json:"context,omitempty"` // e.g. "Attack roll vs Goblin"
}

// Roller is a thread-safe dice roller with a seeded RNG and roll log.
type Roller struct {
	mu  sync.Mutex
	rng *rand.Rand
	log []Result
}

// New creates a Roller with a random seed.
func New() *Roller {
	return NewWithSeed(time.Now().UnixNano())
}

// NewWithSeed creates a Roller with a deterministic seed (useful for tests/replays).
func NewWithSeed(seed int64) *Roller {
	return &Roller{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Standard D&D die sizes.
var validSides = map[int]bool{4: true, 6: true, 8: true, 10: true, 12: true, 20: true, 100: true}

// expressionRe matches NdM+K / NdM-K / NdM (N dice of M sides with optional modifier).
var expressionRe = regexp.MustCompile(`(?i)^(\d+)d(\d+)([+-]\d+)?$`)

// Roll parses and evaluates a dice expression such as "2d6+3" or "1d20".
// An optional context string is attached to the log entry.
func (r *Roller) Roll(expression, context string) (Result, error) {
	expr := strings.TrimSpace(strings.ToLower(expression))
	m := expressionRe.FindStringSubmatch(expr)
	if m == nil {
		return Result{}, fmt.Errorf("invalid dice expression %q (expected NdM or NdM±K)", expression)
	}

	count, _ := strconv.Atoi(m[1])
	sides, _ := strconv.Atoi(m[2])
	modifier := 0
	if m[3] != "" {
		modifier, _ = strconv.Atoi(m[3])
	}

	if count < 1 || count > 100 {
		return Result{}, fmt.Errorf("dice count must be 1–100, got %d", count)
	}
	if sides < 2 || sides > 100 {
		return Result{}, fmt.Errorf("die sides must be 2–100, got %d", sides)
	}

	r.mu.Lock()
	rolls := make([]int, count)
	total := modifier
	for i := range rolls {
		roll := r.rng.Intn(sides) + 1
		rolls[i] = roll
		total += roll
	}
	r.mu.Unlock()

	res := Result{
		Expression: strings.ToUpper(expression),
		Rolls:      rolls,
		Modifier:   modifier,
		Total:      total,
		Timestamp:  time.Now(),
		Context:    context,
	}

	r.mu.Lock()
	r.log = append(r.log, res)
	r.mu.Unlock()

	return res, nil
}

// D20 is a convenience shortcut for rolling a single d20.
func (r *Roller) D20(context string) (Result, error) {
	return r.Roll("1d20", context)
}

// Log returns all rolls recorded since the Roller was created.
func (r *Roller) Log() []Result {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Result, len(r.log))
	copy(out, r.log)
	return out
}

// ClearLog empties the roll history.
func (r *Roller) ClearLog() {
	r.mu.Lock()
	r.log = nil
	r.mu.Unlock()
}
