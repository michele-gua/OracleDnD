package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/dnd-dm/internal/ai"
	"github.com/dnd-dm/internal/config"
	"github.com/dnd-dm/internal/dice"
	"github.com/dnd-dm/internal/handlers"
	"github.com/dnd-dm/internal/session"
)

func main() {
	cfgPath := flag.String("config", "config.json", "path to config.json")
	flag.Parse()

	// --- Load config ---
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// --- Wire dependencies ---
	aiClient := ai.New(cfg.OllamaURL)
	sessionMgr := session.NewManager(cfg.DataDir, cfg.MaxHistoryMessages)
	diceRoller := dice.New()

	deps := &handlers.Deps{
		Cfg:      cfg,
		AI:       aiClient,
		Sessions: sessionMgr,
		Dice:     diceRoller,
	}

	// --- Router ---
	mux := http.NewServeMux()

	// Static files & templates (served from web/ — populated in Part 2)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Page routes (placeholder handlers — full templates in Part 2)
	mux.HandleFunc("GET /", pageHandler("web/templates/campaigns.html"))
	mux.HandleFunc("GET /campaigns", pageHandler("web/templates/campaigns.html"))
	mux.HandleFunc("GET /campaign/{name}/manage", pageHandler("web/templates/manage.html"))
	mux.HandleFunc("GET /campaign/{name}/export", pageHandler("web/templates/export.html"))
	mux.HandleFunc("GET /import", pageHandler("web/templates/import.html"))
	mux.HandleFunc("GET /character/new", pageHandler("web/templates/character_new.html"))
	mux.HandleFunc("GET /character/{id}", pageHandler("web/templates/character.html"))
	mux.HandleFunc("GET /character/import", pageHandler("web/templates/character_import.html"))
	mux.HandleFunc("GET /session/new", pageHandler("web/templates/session_new.html"))
	mux.HandleFunc("GET /session/{id}", pageHandler("web/templates/session.html"))

	// AI & chat
	mux.HandleFunc("POST /api/chat", handlers.ChatHandler(deps))
	mux.HandleFunc("GET /api/models", handlers.ModelsHandler(deps))
	mux.HandleFunc("POST /api/model/select", handlers.ModelSelectHandler(deps))

	// Dice
	mux.HandleFunc("POST /api/roll", handlers.RollHandler(deps))
	mux.HandleFunc("GET /api/roll/suggest", handlers.SuggestRollHandler(deps))

	// Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	addr := ":" + cfg.Port
	log.Printf("🎲 D&D DM server running at http://localhost%s", addr)
	log.Printf("   Ollama: %s  |  Default model: %s", cfg.OllamaURL, cfg.DefaultModel)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// pageHandler serves a static HTML file (placeholder for Part 2 templates).
func pageHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}
