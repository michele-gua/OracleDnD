package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dnd-dm/internal/ai"
	"github.com/dnd-dm/internal/config"
	"github.com/dnd-dm/internal/dice"
	"github.com/dnd-dm/internal/session"
)

// Deps bundles all dependencies the handlers need.
type Deps struct {
	Cfg      *config.Config
	AI       *ai.Client
	Sessions *session.Manager
	Dice     *dice.Roller
}

// ---- helpers ----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

// ---- /api/models ------------------------------------------------------------

// ModelsHandler returns the list of models available in Ollama.
func ModelsHandler(d *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		models, err := d.AI.ListModels(ctx)
		if err != nil {
			writeError(w, http.StatusBadGateway, "cannot reach Ollama: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"models": models})
	}
}

// ---- /api/model/select ------------------------------------------------------

// ModelSelectHandler changes the active model for a session.
func ModelSelectHandler(d *Deps) http.HandlerFunc {
	type req struct {
		SessionID string `json:"session_id"`
		Model     string `json:"model"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var body req
		if err := decodeJSON(r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if body.SessionID == "" || body.Model == "" {
			writeError(w, http.StatusBadRequest, "session_id and model required")
			return
		}

		s, err := d.Sessions.Get(body.SessionID)
		if err != nil {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		s.SetModel(body.Model)
		writeJSON(w, http.StatusOK, map[string]string{"model": body.Model})
	}
}

// ---- /api/roll --------------------------------------------------------------

// RollHandler evaluates a dice expression and returns the result.
func RollHandler(d *Deps) http.HandlerFunc {
	type req struct {
		Expression string `json:"expression"` // e.g. "2d6+3"
		Context    string `json:"context"`    // e.g. "Attack roll"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var body req
		if err := decodeJSON(r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if body.Expression == "" {
			writeError(w, http.StatusBadRequest, "expression required")
			return
		}

		result, err := d.Dice.Roll(body.Expression, body.Context)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

// ---- /api/roll/suggest ------------------------------------------------------

// SuggestRollHandler returns contextually appropriate roll suggestions
// for the current session (placeholder — will be AI-driven in Part 3).
func SuggestRollHandler(d *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		suggestions := []map[string]string{
			{"expression": "1d20", "label": "Tiro abilità generico"},
			{"expression": "1d20", "label": "Tiro attacco"},
			{"expression": "1d8+3", "label": "Danno spada lunga"},
			{"expression": "1d4+1", "label": "Danno pugnale"},
		}
		writeJSON(w, http.StatusOK, map[string]any{"suggestions": suggestions})
	}
}

// ---- /api/chat (SSE streaming) ---------------------------------------------

// dmSystemPrompt is the base system prompt injected into every chat request.
const dmSystemPrompt = `Sei il Dungeon Master di una campagna di D&D. 
Il tuo compito è narrare in modo vivido, coinvolgente e coerente. 
Puoi emettere tag speciali nel testo per triggerare azioni nel sistema:
- [ROLL:espressione] — chiedi al giocatore di tirare un dado, es. [ROLL:1d20]
- [MAP_UPDATE:json] — aggiorna la mappa, es. [MAP_UPDATE:{"room":"taverna","token":"player","x":3,"y":2}]
- [LORE:testo] — aggiungi una voce al lore journal
- [SUGGEST_ROLL:espressione|motivo] — suggerisci un tiro, es. [SUGGEST_ROLL:1d20|Percezione]
- [XP:quantità] — assegna punti esperienza, es. [XP:150]
- [LEVEL_UP:livello] — notifica level up, es. [LEVEL_UP:3]
Usa questi tag con parsimonia e solo quando narrativamente appropriato.
Non mostrare i tag raw al giocatore — il sistema li filtrerà.
Rispondi sempre in italiano salvo che il giocatore scriva in un'altra lingua.`

// ChatHandler handles POST /api/chat with SSE streaming back to the client.
// The client receives a stream of text/event-stream events.
// Each event has the form:
//
//	data: {"token":"...", "done":false}\n\n
//
// When streaming ends:
//
//	data: {"token":"", "done":true, "tags":[...]}\n\n
func ChatHandler(d *Deps) http.HandlerFunc {
	type req struct {
		SessionID  string `json:"session_id"`
		Message    string `json:"message"`
		CampaignName string `json:"campaign_name"`
		CharacterID  string `json:"character_id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// --- SSE headers ---
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		sendEvent := func(data string) {
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}

		sendErr := func(msg string) {
			b, _ := json.Marshal(map[string]string{"error": msg})
			sendEvent(string(b))
		}

		// --- parse body ---
		var body req
		if err := decodeJSON(r, &body); err != nil {
			sendErr("invalid JSON")
			return
		}
		if body.Message == "" {
			sendErr("message required")
			return
		}
		if body.SessionID == "" {
			body.SessionID = fmt.Sprintf("session_%d", time.Now().UnixMilli())
		}

		// --- get or create session ---
		s := d.Sessions.GetOrNew(
			body.SessionID,
			body.CampaignName,
			body.CharacterID,
			d.Cfg.DefaultModel,
			"dnd5e",
		)

		// --- append user message ---
		s.AppendMessage(ai.Message{Role: "user", Content: body.Message})

		// --- build messages for Ollama ---
		messages := s.Messages(dmSystemPrompt)

		// --- check if history needs compression (every maxMsgs messages) ---
		historyLen := len(s.History)
		needsSummary := historyLen > 0 && historyLen%d.Cfg.MaxHistoryMessages == 0

		// --- stream from Ollama ---
		var fullResponse strings.Builder
		model := s.Model
		if model == "" {
			model = d.Cfg.DefaultModel
		}

		opts := &ai.Options{
			NumCtx: int(float64(d.Cfg.ContextWindow(model)) * d.Cfg.ContextReserveRatio),
		}

		err := d.AI.StreamChat(r.Context(), model, messages, opts, func(token string) error {
			fullResponse.WriteString(token)
			b, _ := json.Marshal(map[string]any{"token": token, "done": false})
			sendEvent(string(b))
			return nil
		})

		if err != nil && r.Context().Err() == nil {
			sendErr("AI error: " + err.Error())
			return
		}

		// --- parse AI tags from full response ---
		rawResponse := fullResponse.String()
		narrative, tags := ai.ParseTags(rawResponse)

		// --- append assistant message (clean narrative) ---
		s.AppendMessage(ai.Message{Role: "assistant", Content: narrative})

		// --- persist session ---
		_ = d.Sessions.Save(s)

		// --- auto-summarise if needed ---
		if needsSummary {
			go func() {
				summariseSession(d, s)
			}()
		}

		// --- send final SSE event ---
		tagList := make([]map[string]string, len(tags))
		for i, t := range tags {
			tagList[i] = map[string]string{"type": t.Type, "payload": t.Payload}
		}
		b, _ := json.Marshal(map[string]any{
			"token": "",
			"done":  true,
			"tags":  tagList,
		})
		sendEvent(string(b))
	}
}

// summariseSession asks the AI to compress the session history.
func summariseSession(d *Deps, s *session.State) {
	msgs := s.Messages("Sei un assistente che riassume conversazioni di D&D in modo molto conciso.")
	msgs = append(msgs, ai.Message{
		Role: "user",
		Content: fmt.Sprintf(
			"Riassumi questa sessione di D&D in massimo %d token, mantenendo: "+
				"eventi chiave, decisioni del giocatore, NPC incontrati, oggetti trovati. "+
				"Sii conciso e narrativo.",
			d.Cfg.SummaryMaxTokens,
		),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	summary, err := d.AI.Chat(ctx, s.Model, msgs, nil)
	if err != nil {
		return
	}

	s.SetSummary(summary, d.Cfg.MaxHistoryMessages)
	_ = d.Sessions.Save(s)
}
