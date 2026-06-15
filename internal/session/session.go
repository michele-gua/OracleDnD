package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dnd-dm/internal/ai"
)

// State holds the full in-memory state of an active session.
type State struct {
	mu sync.RWMutex

	ID           string       `json:"id"`
	CampaignName string       `json:"campaign_name"`
	CharacterID  string       `json:"character_id"`
	Model        string       `json:"model"`
	Ruleset      string       `json:"ruleset"`
	History      []ai.Message `json:"history"`
	Summary      string       `json:"summary,omitempty"` // compressed history summary
	StartedAt    time.Time    `json:"started_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// Manager keeps a registry of active sessions keyed by session ID.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*State
	dataDir  string
	maxMsgs  int
}

// NewManager creates a session manager that persists state under dataDir.
func NewManager(dataDir string, maxHistoryMessages int) *Manager {
	return &Manager{
		sessions: make(map[string]*State),
		dataDir:  dataDir,
		maxMsgs:  maxHistoryMessages,
	}
}

// Get returns an existing session by ID (loads from disk if not in memory).
func (m *Manager) Get(id string) (*State, error) {
	m.mu.RLock()
	s, ok := m.sessions[id]
	m.mu.RUnlock()
	if ok {
		return s, nil
	}
	return m.load(id)
}

// GetOrNew returns an existing session or creates a new one.
func (m *Manager) GetOrNew(id, campaign, character, model, ruleset string) *State {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.sessions[id]; ok {
		return s
	}

	s := &State{
		ID:           id,
		CampaignName: campaign,
		CharacterID:  character,
		Model:        model,
		Ruleset:      ruleset,
		StartedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	m.sessions[id] = s
	return s
}

// AppendMessage adds a message to session history.
// Returns the trimmed history length.
func (s *State) AppendMessage(msg ai.Message) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.History = append(s.History, msg)
	s.UpdatedAt = time.Now()
	return len(s.History)
}

// SetSummary stores a compressed summary and clears old messages beyond the
// recent window.
func (s *State) SetSummary(summary string, keepLast int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Summary = summary
	if len(s.History) > keepLast {
		s.History = s.History[len(s.History)-keepLast:]
	}
	s.UpdatedAt = time.Now()
}

// Messages returns the full conversation suitable for sending to Ollama:
// system context (if summary exists) + recent history.
func (s *State) Messages(systemPrompt string) []ai.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var msgs []ai.Message

	// Build system prompt, injecting summary if available.
	sys := systemPrompt
	if s.Summary != "" {
		sys += "\n\n### Riassunto della sessione precedente ###\n" + s.Summary
	}
	msgs = append(msgs, ai.Message{Role: "system", Content: sys})
	msgs = append(msgs, s.History...)
	return msgs
}

// SetModel updates the model for this session.
func (s *State) SetModel(model string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Model = model
	s.UpdatedAt = time.Now()
}

// Save persists the session to disk as JSON.
func (m *Manager) Save(s *State) error {
	dir := m.sessionDir(s.CampaignName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	s.mu.RLock()
	data, err := json.MarshalIndent(s, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}

	path := filepath.Join(dir, s.ID+".json")
	return os.WriteFile(path, data, 0644)
}

func (m *Manager) load(id string) (*State, error) {
	// Search across all campaign dirs for this session ID.
	entries, err := os.ReadDir(filepath.Join(m.dataDir, "campaigns"))
	if err != nil {
		return nil, fmt.Errorf("session %s not found", id)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(m.dataDir, "campaigns", e.Name(), "sessions", id+".json")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var s State
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("session decode: %w", err)
		}
		m.mu.Lock()
		m.sessions[id] = &s
		m.mu.Unlock()
		return &s, nil
	}
	return nil, fmt.Errorf("session %s not found", id)
}

func (m *Manager) sessionDir(campaign string) string {
	return filepath.Join(m.dataDir, "campaigns", campaign, "sessions")
}
