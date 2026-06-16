package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dnd-dm/internal/campaign"
	"github.com/dnd-dm/internal/ruleset"
)

// GenerateCampaignHandler handles POST /api/campaign/generate.
// It streams progress over SSE while the AI builds the world.
//
// Request body:
//
//	{ "slug": "la_maledizione_di_strahd", "model": "deepseek-r1:8b" }
//
// SSE events:
//
//	data: {"phase":"world","message":"Il DM sta plasmando il mondo…","token":"","done":false}
//	data: {"phase":"saving","message":"Salvando…","token":"","done":false}
//	data: {"phase":"done","message":"Campagna pronta!","token":"","done":true,"world":{...}}
func GenerateCampaignHandler(d *Deps, cm *campaign.Manager, rm *ruleset.Manager) http.HandlerFunc {
	type req struct {
		Slug  string `json:"slug"`
		Model string `json:"model"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		sendRaw := func(data string) {
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}

		sendEvent := func(phase, message, token string, done bool, extra map[string]any) {
			ev := map[string]any{
				"phase":   phase,
				"message": message,
				"token":   token,
				"done":    done,
			}
			for k, v := range extra {
				ev[k] = v
			}
			b, _ := json.Marshal(ev)
			sendRaw(string(b))
		}

		// Parse request
		var body req
		if err := decodeJSON(r, &body); err != nil || body.Slug == "" {
			sendEvent("error", "slug required", "", true, map[string]any{"error": "slug required"})
			return
		}

		model := body.Model
		if model == "" {
			model = d.Cfg.DefaultModel
		}

		// Load campaign
		c, err := cm.Get(body.Slug)
		if err != nil {
			sendEvent("error", "campaign not found: "+err.Error(), "", true, map[string]any{"error": err.Error()})
			return
		}

		// Create generator
		gen := campaign.NewGenerator(d.AI, rm)
		progressCh := make(chan campaign.GenerateProgress, 64)

		// Run generation in goroutine
		var genErr error
		var world *campaign.GeneratedWorld

		done := make(chan struct{})
		go func() {
			defer close(done)
			world, genErr = gen.Generate(r.Context(), model, c, progressCh)
		}()

		// Stream progress to client
		for prog := range progressCh {
			if prog.Error != "" {
				sendEvent("error", prog.Error, "", true, map[string]any{"error": prog.Error})
				return
			}
			if prog.Token != "" {
				sendEvent(prog.Phase, "", prog.Token, false, nil)
			} else if prog.Message != "" {
				sendEvent(prog.Phase, prog.Message, "", false, nil)
			}
		}

		<-done

		if genErr != nil {
			sendEvent("error", genErr.Error(), "", true, map[string]any{"error": genErr.Error()})
			return
		}

		// Save world to markdown
		sendEvent("saving", "Incidendo la storia su pergamena…", "", false, nil)
		if err := campaign.SaveWorldToMarkdown(cm, c, world); err != nil {
			sendEvent("error", "save failed: "+err.Error(), "", true, map[string]any{"error": err.Error()})
			return
		}

		// Reload campaign with updated data
		c, _ = cm.Get(body.Slug)

		sendEvent("done", "Il mondo è pronto. L'avventura può iniziare.", "", true, map[string]any{
			"campaign": c,
			"opening":  world.OpeningScene,
			"ascii":    world.AsciiArt,
		})
	}
}

// RulesetsHandler returns available rulesets (GET /api/rulesets).
func RulesetsHandler(rm *ruleset.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ids, err := rm.List()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		type entry struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			ShortName string `json:"short_name"`
			Version   string `json:"version"`
		}

		var list []entry
		for _, id := range ids {
			rs, err := rm.Get(id)
			if err != nil {
				continue
			}
			list = append(list, entry{
				ID:        rs.ID,
				Name:      rs.Name,
				ShortName: rs.ShortName,
				Version:   rs.Version,
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{"rulesets": list})
	}
}

// RulesetSelectHandler switches ruleset for a session (POST /api/ruleset/select).
func RulesetSelectHandler(d *Deps) http.HandlerFunc {
	type req struct {
		SessionID string `json:"session_id"`
		Ruleset   string `json:"ruleset"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var body req
		if err := decodeJSON(r, &body); err != nil || body.Ruleset == "" {
			writeError(w, http.StatusBadRequest, "ruleset required")
			return
		}
		s, err := d.Sessions.Get(body.SessionID)
		if err != nil {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		// Update ruleset in session state
		_ = s // session.Ruleset field will be added in Part 4 with character work
		writeJSON(w, http.StatusOK, map[string]string{"ruleset": body.Ruleset})
	}
}
