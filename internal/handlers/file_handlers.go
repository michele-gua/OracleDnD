package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dnd-dm/internal/campaign"
)

// CampaignFileHandler serves a markdown file from a campaign directory.
// GET /api/campaign/{name}/file?path=world/factions.md
// Returns: { "content": "# Fazioni\n\n..." }
func CampaignFileHandler(cm *campaign.Manager, dataDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("name")
		rel  := r.URL.Query().Get("path")

		if rel == "" {
			writeError(w, http.StatusBadRequest, "path query param required")
			return
		}

		// Security: reject any path traversal
		rel = filepath.Clean(rel)
		if strings.Contains(rel, "..") {
			writeError(w, http.StatusBadRequest, "invalid path")
			return
		}

		fullPath := filepath.Join(dataDir, "campaigns", slug, rel)

		data, err := os.ReadFile(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				writeError(w, http.StatusNotFound, "file not found")
			} else {
				writeError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"content": string(data)})
	}
}

// CampaignFilesListHandler lists all markdown files for a campaign.
// GET /api/campaign/{name}/files
func CampaignFilesListHandler(cm *campaign.Manager, dataDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug    := r.PathValue("name")
		baseDir := filepath.Join(dataDir, "campaigns", slug)

		var files []map[string]string
		err := walkMarkdown(baseDir, baseDir, &files)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"files": files})
	}
}

func walkMarkdown(baseDir, dir string, out *[]map[string]string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		fullPath := filepath.Join(dir, e.Name())
		if e.IsDir() {
			// Skip sessions dir (too much data)
			if e.Name() == "sessions" {
				continue
			}
			_ = walkMarkdown(baseDir, fullPath, out)
			continue
		}
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		rel, _ := filepath.Rel(baseDir, fullPath)
		*out = append(*out, map[string]string{
			"path": rel,
			"name": e.Name(),
		})
	}
	return nil
}
