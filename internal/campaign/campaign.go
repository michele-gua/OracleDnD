package campaign

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Status represents the lifecycle state of a campaign.
type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusAbandoned Status = "abandoned"
	StatusArchived  Status = "archived"
)

// CharacterRef is a lightweight reference to a character in a campaign card.
type CharacterRef struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Class string `json:"class"`
	Race  string `json:"race"`
	Level int    `json:"level"`
}

// Campaign is the full metadata for a campaign (stored in campaign.json).
type Campaign struct {
	Slug           string         `json:"slug"`             // filesystem-safe name
	Name           string         `json:"name"`             // display name
	WorldName      string         `json:"world_name"`
	Ruleset        string         `json:"ruleset"`          // e.g. "dnd5e"
	Status         Status         `json:"status"`
	Characters     []CharacterRef `json:"characters"`
	SessionCount   int            `json:"session_count"`
	TotalMinutes   int            `json:"total_minutes"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	LastSessionAt  *time.Time     `json:"last_session_at,omitempty"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
	LastSummary    string         `json:"last_summary,omitempty"`   // last line of last session summary
	CoverAscii     string         `json:"cover_ascii,omitempty"`    // AI-generated ASCII art cover
	CoverPrompt    string         `json:"cover_prompt,omitempty"`   // Stable Diffusion prompt
	Notes          string         `json:"notes,omitempty"`
}

// Manager handles all campaign persistence.
type Manager struct {
	dataDir string
}

// NewManager creates a campaign manager rooted at dataDir.
func NewManager(dataDir string) *Manager {
	return &Manager{dataDir: dataDir}
}

// campaignDir returns the directory for a given campaign slug.
func (m *Manager) campaignDir(slug string) string {
	return filepath.Join(m.dataDir, "campaigns", slug)
}

// Create initialises a new campaign on disk and returns it.
func (m *Manager) Create(name, worldName, ruleset string) (*Campaign, error) {
	slug := slugify(name)
	dir := m.campaignDir(slug)

	if _, err := os.Stat(dir); err == nil {
		return nil, fmt.Errorf("campaign %q already exists", slug)
	}

	// Create subdirectories.
	for _, sub := range []string{"sessions", "characters", "companions", "world", "npcs", "items"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			return nil, err
		}
	}

	c := &Campaign{
		Slug:      slug,
		Name:      name,
		WorldName: worldName,
		Ruleset:   ruleset,
		Status:    StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := m.save(c); err != nil {
		return nil, err
	}
	return c, nil
}

// Get loads a campaign by slug.
func (m *Manager) Get(slug string) (*Campaign, error) {
	data, err := os.ReadFile(filepath.Join(m.campaignDir(slug), "campaign.json"))
	if err != nil {
		return nil, fmt.Errorf("campaign %q not found", slug)
	}
	var c Campaign
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// List returns all campaigns sorted by last access (most recent first).
func (m *Manager) List() ([]*Campaign, error) {
	base := filepath.Join(m.dataDir, "campaigns")
	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var campaigns []*Campaign
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		c, err := m.Get(e.Name())
		if err != nil {
			continue // skip corrupted entries
		}
		campaigns = append(campaigns, c)
	}

	sort.Slice(campaigns, func(i, j int) bool {
		ti := campaigns[i].UpdatedAt
		tj := campaigns[j].UpdatedAt
		if campaigns[i].LastSessionAt != nil {
			ti = *campaigns[i].LastSessionAt
		}
		if campaigns[j].LastSessionAt != nil {
			tj = *campaigns[j].LastSessionAt
		}
		return ti.After(tj)
	})

	return campaigns, nil
}

// Update saves changes to an existing campaign.
func (m *Manager) Update(c *Campaign) error {
	c.UpdatedAt = time.Now()
	return m.save(c)
}

// Archive hides a campaign from the main grid without deleting it.
func (m *Manager) Archive(slug string) error {
	c, err := m.Get(slug)
	if err != nil {
		return err
	}
	c.Status = StatusArchived
	return m.Update(c)
}

// Complete marks a campaign as completed.
func (m *Manager) Complete(slug, notes string) error {
	c, err := m.Get(slug)
	if err != nil {
		return err
	}
	now := time.Now()
	c.Status = StatusCompleted
	c.CompletedAt = &now
	c.Notes = notes
	return m.Update(c)
}

// Delete permanently removes a campaign from disk.
func (m *Manager) Delete(slug string) error {
	return os.RemoveAll(m.campaignDir(slug))
}

// SaveOverview writes the human-readable campaign overview markdown.
func (m *Manager) SaveOverview(slug, content string) error {
	path := filepath.Join(m.campaignDir(slug), "campaign_overview.md")
	return os.WriteFile(path, []byte(content), 0644)
}

// LoadOverview reads the campaign overview markdown.
func (m *Manager) LoadOverview(slug string) (string, error) {
	data, err := os.ReadFile(filepath.Join(m.campaignDir(slug), "campaign_overview.md"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (m *Manager) save(c *Campaign) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(m.campaignDir(c.Slug), "campaign.json"), data, 0644)
}

// slugify converts a display name into a filesystem-safe slug.
func slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, s)
	// Collapse multiple underscores.
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}
	return strings.Trim(s, "_")
}
