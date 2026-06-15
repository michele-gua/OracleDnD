package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds all application settings loaded from config.json.
type Config struct {
	Port                string         `json:"port"`
	OllamaURL           string         `json:"ollama_url"`
	DefaultModel        string         `json:"default_model"`
	Language            string         `json:"language"`
	PiperPath           string         `json:"piper_path"`
	TTSEnabled          bool           `json:"tts_enabled"`
	ContextWindows      map[string]int `json:"context_windows"`
	ContextReserveRatio float64        `json:"context_reserve_ratio"`
	MaxHistoryMessages  int            `json:"max_history_messages"`
	SummaryMaxTokens    int            `json:"summary_max_tokens"`
	DataDir             string         `json:"data_dir"`
	RulesetsDir         string         `json:"rulesets_dir"`
}

// Load reads and parses config.json from the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: cannot open %s: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config: parse error: %w", err)
	}

	// Sensible defaults
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.OllamaURL == "" {
		cfg.OllamaURL = "http://localhost:11434"
	}
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = "deepseek-r1:8b"
	}
	if cfg.MaxHistoryMessages == 0 {
		cfg.MaxHistoryMessages = 20
	}
	if cfg.ContextReserveRatio == 0 {
		cfg.ContextReserveRatio = 0.80
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "./data"
	}
	if cfg.RulesetsDir == "" {
		cfg.RulesetsDir = "./rulesets"
	}

	return &cfg, nil
}

// ContextWindow returns the configured context window for a model,
// falling back to 4096 if not specified.
func (c *Config) ContextWindow(model string) int {
	if w, ok := c.ContextWindows[model]; ok {
		return w
	}
	return 4096
}
