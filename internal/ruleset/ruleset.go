package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AbilityScore describes a core ability (STR, DEX, etc.).
type AbilityScore struct {
	Name  string `json:"name"`
	Abbr  string `json:"abbr"`
	Desc  string `json:"desc"`
}

// Class is a playable class in the ruleset.
type Class struct {
	Name        string   `json:"name"`
	HitDie      int      `json:"hit_die"`
	PrimaryAbil []string `json:"primary_ability"`
	Desc        string   `json:"desc"`
}

// Race is a playable race/species.
type Race struct {
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Bonus string `json:"ability_bonus,omitempty"`
}

// Alignment is an available alignment.
type Alignment struct {
	Name string `json:"name"`
	Abbr string `json:"abbr"`
}

// SkillDef defines a skill and its governing ability.
type SkillDef struct {
	Name    string `json:"name"`
	Ability string `json:"ability"`
}

// Ruleset is the full rule definition for a game system.
type Ruleset struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	ShortName      string         `json:"short_name"`
	Version        string         `json:"version"`
	AbilityScores  []AbilityScore `json:"ability_scores"`
	Classes        []Class        `json:"classes"`
	Races          []Race         `json:"races"`
	Alignments     []Alignment    `json:"alignments"`
	Skills         []SkillDef     `json:"skills"`
	MaxLevel       int            `json:"max_level"`
	CurrencyNames  []string       `json:"currency_names"`   // e.g. ["PP","PO","PE","PA","PR"]
	SpellLevels    int            `json:"spell_levels"`      // 0 = no spells
	DMPromptHints  string         `json:"dm_prompt_hints"`   // injected into AI system prompt
}

// Manager loads and caches rulesets from the rulesets/ directory.
type Manager struct {
	dir      string
	cache    map[string]*Ruleset
}

// NewManager creates a ruleset manager pointed at dir.
func NewManager(dir string) *Manager {
	return &Manager{dir: dir, cache: make(map[string]*Ruleset)}
}

// Get returns a ruleset by ID, loading from disk if needed.
func (m *Manager) Get(id string) (*Ruleset, error) {
	if r, ok := m.cache[id]; ok {
		return r, nil
	}
	return m.load(id)
}

// List returns all available ruleset IDs.
func (m *Manager) List() ([]string, error) {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			ids = append(ids, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	return ids, nil
}

func (m *Manager) load(id string) (*Ruleset, error) {
	path := filepath.Join(m.dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ruleset %q not found at %s", id, path)
	}
	var r Ruleset
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("ruleset %q parse error: %w", id, err)
	}
	r.ID = id
	m.cache[id] = &r
	return &r, nil
}
