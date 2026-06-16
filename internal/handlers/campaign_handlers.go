package handlers

import (
	"net/http"

	"github.com/dnd-dm/internal/campaign"
)

// CampaignListHandler returns all campaigns as JSON (GET /api/campaign/list).
func CampaignListHandler(cm *campaign.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := cm.List()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"campaigns": list})
	}
}

// CampaignCreateHandler creates a new campaign (POST /api/campaign/create).
func CampaignCreateHandler(cm *campaign.Manager) http.HandlerFunc {
	type req struct {
		Name      string `json:"name"`
		WorldName string `json:"world_name"`
		Ruleset   string `json:"ruleset"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var body req
		if err := decodeJSON(r, &body); err != nil || body.Name == "" {
			writeError(w, http.StatusBadRequest, "name required")
			return
		}
		if body.Ruleset == "" {
			body.Ruleset = "dnd5e"
		}
		c, err := cm.Create(body.Name, body.WorldName, body.Ruleset)
		if err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, c)
	}
}

// CampaignArchiveHandler archives a campaign (POST /api/campaign/{name}/archive).
func CampaignArchiveHandler(cm *campaign.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("name")
		if err := cm.Archive(slug); err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "archived"})
	}
}

// CampaignCompleteHandler marks a campaign as completed (POST /api/campaign/{name}/complete).
func CampaignCompleteHandler(cm *campaign.Manager) http.HandlerFunc {
	type req struct {
		Notes string `json:"notes"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("name")
		var body req
		_ = decodeJSON(r, &body)
		if err := cm.Complete(slug, body.Notes); err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
	}
}

// CampaignDeleteHandler permanently deletes a campaign (DELETE /api/campaign/{name}).
func CampaignDeleteHandler(cm *campaign.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("name")
		if err := cm.Delete(slug); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

// CampaignGetHandler returns a single campaign (GET /api/campaign/{name}).
func CampaignGetHandler(cm *campaign.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("name")
		c, err := cm.Get(slug)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, c)
	}
}

// CampaignUpdateHandler updates campaign metadata (PATCH /api/campaign/{name}).
func CampaignUpdateHandler(cm *campaign.Manager) http.HandlerFunc {
	type req struct {
		Name      string `json:"name"`
		WorldName string `json:"world_name"`
		Notes     string `json:"notes"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("name")
		c, err := cm.Get(slug)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		var body req
		if err := decodeJSON(r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if body.Name != "" {
			c.Name = body.Name
		}
		if body.WorldName != "" {
			c.WorldName = body.WorldName
		}
		if body.Notes != "" {
			c.Notes = body.Notes
		}
		if err := cm.Update(c); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, c)
	}
}
