package campaign

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dnd-dm/internal/ai"
	"github.com/dnd-dm/internal/ruleset"
)

// GenerateProgress is sent over the progress channel during generation.
type GenerateProgress struct {
	Phase   string // "world", "factions", "npcs", "hooks", "saving", "done"
	Message string // human-readable narrative message shown in UI
	Token   string // streamed token (empty for phase transitions)
	Done    bool
	Error   string
}

// GeneratedWorld is the full AI-generated campaign structure before saving.
type GeneratedWorld struct {
	WorldName    string          `json:"world_name"`
	Premise      string          `json:"premise"`
	Atmosphere   string          `json:"atmosphere"`
	MainConflict string          `json:"main_conflict"`
	Factions     []GenFaction    `json:"factions"`
	StartingArea GenArea         `json:"starting_area"`
	MainNPCs     []GenNPC        `json:"main_npcs"`
	Hooks        []GenHook       `json:"hooks"`
	WorldTimeline []GenEvent     `json:"world_timeline"`
	OpeningScene string          `json:"opening_scene"`
	CoverPrompt  string          `json:"cover_prompt"`
	AsciiArt     string          `json:"ascii_art"`
}

type GenFaction struct {
	Name     string `json:"name"`
	Goal     string `json:"goal"`
	Leader   string `json:"leader"`
	Attitude string `json:"attitude"` // "friendly","neutral","hostile","unknown"
	Secret   string `json:"secret"`
}

type GenArea struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "village","city","dungeon","wilderness"
	Description string `json:"description"`
	KeyLocations []string `json:"key_locations"`
}

type GenNPC struct {
	Name        string `json:"name"`
	Role        string `json:"role"`
	Personality string `json:"personality"`
	Secret      string `json:"secret"`
	Attitude    string `json:"attitude"` // toward party
}

type GenHook struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Urgency     string `json:"urgency"` // "immediate","short","long"
}

type GenEvent struct {
	When        string `json:"when"` // relative: "Passato remoto", "50 anni fa", "La settimana scorsa"
	Description string `json:"description"`
}

// Generator handles AI-driven campaign creation.
type Generator struct {
	ai      *ai.Client
	rulesets *ruleset.Manager
}

// NewGenerator creates a campaign generator.
func NewGenerator(aiClient *ai.Client, rm *ruleset.Manager) *Generator {
	return &Generator{ai: aiClient, rulesets: rm}
}

// Generate creates a full campaign world via Ollama and streams progress.
// The caller receives updates on the progress channel and should drain it.
func (g *Generator) Generate(
	ctx context.Context,
	model string,
	c *Campaign,
	progressCh chan<- GenerateProgress,
) (*GeneratedWorld, error) {

	defer close(progressCh)

	rs, err := g.rulesets.Get(c.Ruleset)
	if err != nil {
		// Fallback to dnd5e hints if ruleset file missing
		rs = &ruleset.Ruleset{DMPromptHints: "Usa meccaniche D&D 5e standard."}
	}

	send := func(phase, msg, token string, done bool) {
		select {
		case progressCh <- GenerateProgress{Phase: phase, Message: msg, Token: token, Done: done}:
		case <-ctx.Done():
		}
	}

	// ── Phase 1: World & premise ──────────────────────────────────────────
	send("world", "Il Dungeon Master sta plasmando il mondo…", "", false)

	worldPrompt := buildWorldPrompt(c, rs)
	worldJSON, err := g.callJSON(ctx, model, worldPrompt, progressCh)
	if err != nil {
		return nil, fmt.Errorf("world generation: %w", err)
	}

	var world GeneratedWorld
	if err := json.Unmarshal([]byte(worldJSON), &world); err != nil {
		// Try to extract JSON from markdown fences
		worldJSON = extractJSON(worldJSON)
		if err2 := json.Unmarshal([]byte(worldJSON), &world); err2 != nil {
			return nil, fmt.Errorf("world JSON parse: %w (raw: %.200s)", err, worldJSON)
		}
	}

	// Fill campaign name from AI if world_name generated
	if world.WorldName != "" && c.WorldName == "" {
		c.WorldName = world.WorldName
	}

	// ── Phase 2: ASCII art cover ──────────────────────────────────────────
	send("ascii", "Forgiando la copertina della campagna…", "", false)

	asciiPrompt := fmt.Sprintf(
		`Crea un'illustrazione ASCII art (max 20 righe × 60 colonne) che rappresenti visivamente questa campagna D&D:
Titolo: %s
Mondo: %s
Atmosfera: %s
Conflitto principale: %s

Usa caratteri ASCII per creare un'immagine evocativa (paesaggio, dungeon, simbolo, creatura, ecc.).
Rispondi SOLO con l'ASCII art, nessun testo aggiuntivo.`,
		c.Name, world.WorldName, world.Atmosphere, world.MainConflict,
	)

	var asciiSB strings.Builder
	_ = g.ai.StreamChat(ctx, model,
		[]ai.Message{{Role: "user", Content: asciiPrompt}},
		nil,
		func(token string) error {
			asciiSB.WriteString(token)
			send("ascii", "", token, false)
			return nil
		},
	)
	world.AsciiArt = strings.TrimSpace(asciiSB.String())

	// ── Phase 3: Saving ────────────────────────────────────────────────────
	send("saving", "Salvando il mondo su disco…", "", false)

	return &world, nil
}

// callJSON calls the AI and expects a JSON response, streaming tokens to progressCh.
func (g *Generator) callJSON(
	ctx context.Context,
	model string,
	prompt string,
	progressCh chan<- GenerateProgress,
) (string, error) {

	msgs := []ai.Message{
		{
			Role: "system",
			Content: "Sei un generatore di campagne D&D. Rispondi SEMPRE e SOLO con JSON valido, " +
				"senza markdown, senza testo aggiuntivo, senza spiegazioni. Solo JSON puro.",
		},
		{Role: "user", Content: prompt},
	}

	var sb strings.Builder
	err := g.ai.StreamChat(ctx, model, msgs, &ai.Options{Temperature: 0.85}, func(token string) error {
		sb.WriteString(token)
		select {
		case progressCh <- GenerateProgress{Token: token}:
		default:
		}
		return nil
	})

	return sb.String(), err
}

func buildWorldPrompt(c *Campaign, rs *ruleset.Ruleset) string {
	worldHint := ""
	if c.WorldName != "" {
		worldHint = fmt.Sprintf(`Il nome del mondo è "%s".`, c.WorldName)
	}

	classNames := make([]string, len(rs.Classes))
	for i, cl := range rs.Classes {
		classNames[i] = cl.Name
	}

	return fmt.Sprintf(`Crea una campagna D&D originale e dettagliata in italiano.
Titolo campagna: "%s"
Regolamento: %s
%s
Classi disponibili: %s

Genera un oggetto JSON con questa struttura ESATTA (tutti i campi obbligatori):
{
  "world_name": "nome del mondo fantasy",
  "premise": "descrizione della situazione iniziale (3-4 frasi evocative)",
  "atmosphere": "atmosfera e tono (es: dark gothic, high fantasy, intrigo politico)",
  "main_conflict": "il conflitto principale della campagna (2-3 frasi)",
  "factions": [
    {
      "name": "nome fazione",
      "goal": "obiettivo della fazione",
      "leader": "nome del leader",
      "attitude": "friendly|neutral|hostile|unknown",
      "secret": "segreto oscuro della fazione"
    }
  ],
  "starting_area": {
    "name": "nome dell'area di partenza",
    "type": "village|city|dungeon|wilderness",
    "description": "descrizione vivida dell'area (3-4 frasi)",
    "key_locations": ["locazione1", "locazione2", "locazione3"]
  },
  "main_npcs": [
    {
      "name": "nome NPC",
      "role": "ruolo nella storia",
      "personality": "tratti caratteriali",
      "secret": "segreto dell'NPC",
      "attitude": "come si pone verso i giocatori"
    }
  ],
  "hooks": [
    {
      "title": "titolo della quest hook",
      "description": "descrizione della missione disponibile",
      "urgency": "immediate|short|long"
    }
  ],
  "world_timeline": [
    {
      "when": "periodo temporale (es: 300 anni fa, la scorsa settimana)",
      "description": "evento storico che ha plasmato il mondo"
    }
  ],
  "opening_scene": "scena di apertura narrativa dettagliata che il DM leggerà ai giocatori (5-8 frasi immersive in seconda persona)",
  "cover_prompt": "prompt per Stable Diffusion per generare un'immagine di copertina fantasy"
}

Crea almeno: 3 fazioni, 4 NPC principali, 3 quest hook, 5 eventi nella timeline.
Assicurati che tutto sia coerente, originale e in italiano.`,
		c.Name, rs.Name, worldHint, strings.Join(classNames, ", "),
	)
}

// extractJSON tries to pull a JSON object out of a string that may have markdown fences.
func extractJSON(s string) string {
	// Remove ```json fences
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	// Find first { and last }
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

// SaveWorldToMarkdown converts a GeneratedWorld into human-readable markdown files.
func SaveWorldToMarkdown(cm *Manager, c *Campaign, world *GeneratedWorld) error {
	slug := c.Slug

	// ── campaign_overview.md ──────────────────────────────────────────────
	var overview strings.Builder
	overview.WriteString(fmt.Sprintf("# %s\n\n", c.Name))
	overview.WriteString(fmt.Sprintf("**Mondo:** %s  \n", world.WorldName))
	overview.WriteString(fmt.Sprintf("**Regolamento:** %s  \n", c.Ruleset))
	overview.WriteString(fmt.Sprintf("**Atmosfera:** %s  \n\n", world.Atmosphere))
	overview.WriteString("## Premessa\n\n")
	overview.WriteString(world.Premise + "\n\n")
	overview.WriteString("## Conflitto Principale\n\n")
	overview.WriteString(world.MainConflict + "\n\n")
	overview.WriteString("## Scena di Apertura\n\n")
	overview.WriteString(world.OpeningScene + "\n\n")

	if len(world.Factions) > 0 {
		overview.WriteString("## Fazioni\n\n")
		for _, f := range world.Factions {
			overview.WriteString(fmt.Sprintf("### %s\n", f.Name))
			overview.WriteString(fmt.Sprintf("- **Leader:** %s\n", f.Leader))
			overview.WriteString(fmt.Sprintf("- **Obiettivo:** %s\n", f.Goal))
			overview.WriteString(fmt.Sprintf("- **Atteggiamento:** %s\n\n", f.Attitude))
		}
	}

	if err := cm.SaveOverview(slug, overview.String()); err != nil {
		return fmt.Errorf("save overview: %w", err)
	}

	// ── world/world_timeline.md ───────────────────────────────────────────
	if len(world.WorldTimeline) > 0 {
		var timeline strings.Builder
		timeline.WriteString(fmt.Sprintf("# Cronologia di %s\n\n", world.WorldName))
		for _, ev := range world.WorldTimeline {
			timeline.WriteString(fmt.Sprintf("## %s\n\n%s\n\n", ev.When, ev.Description))
		}
		if err := cm.saveWorldFile(slug, "world_timeline.md", timeline.String()); err != nil {
			return err
		}
	}

	// ── world/side_quests.md ──────────────────────────────────────────────
	if len(world.Hooks) > 0 {
		var quests strings.Builder
		quests.WriteString("# Quest Hooks Disponibili\n\n")
		for _, h := range world.Hooks {
			quests.WriteString(fmt.Sprintf("## %s\n\n", h.Title))
			quests.WriteString(fmt.Sprintf("**Urgenza:** %s\n\n", h.Urgency))
			quests.WriteString(h.Description + "\n\n")
		}
		if err := cm.saveWorldFile(slug, "side_quests.md", quests.String()); err != nil {
			return err
		}
	}

	// ── npcs/{name}.md ────────────────────────────────────────────────────
	for _, npc := range world.MainNPCs {
		var npcMd strings.Builder
		safeName := slugifyName(npc.Name)
		npcMd.WriteString(fmt.Sprintf("# %s\n\n", npc.Name))
		npcMd.WriteString(fmt.Sprintf("**Ruolo:** %s  \n", npc.Role))
		npcMd.WriteString(fmt.Sprintf("**Atteggiamento:** %s  \n\n", npc.Attitude))
		npcMd.WriteString("## Personalità\n\n" + npc.Personality + "\n\n")
		npcMd.WriteString("## Segreto\n\n" + npc.Secret + "\n\n")
		npcMd.WriteString("## Storia degli Incontri\n\n_Nessun incontro registrato._\n")
		if err := cm.saveNPCFile(slug, safeName+".md", npcMd.String()); err != nil {
			return err
		}
	}

	// ── world/starting_area.md ────────────────────────────────────────────
	var area strings.Builder
	area.WriteString(fmt.Sprintf("# %s\n\n", world.StartingArea.Name))
	area.WriteString(fmt.Sprintf("**Tipo:** %s\n\n", world.StartingArea.Type))
	area.WriteString(world.StartingArea.Description + "\n\n")
	if len(world.StartingArea.KeyLocations) > 0 {
		area.WriteString("## Luoghi Chiave\n\n")
		for _, loc := range world.StartingArea.KeyLocations {
			area.WriteString(fmt.Sprintf("- %s\n", loc))
		}
	}
	if err := cm.saveWorldFile(slug, "starting_area.md", area.String()); err != nil {
		return err
	}

	// ── world/factions.md ─────────────────────────────────────────────────
	if len(world.Factions) > 0 {
		var fac strings.Builder
		fac.WriteString("# Fazioni\n\n")
		for _, f := range world.Factions {
			fac.WriteString(fmt.Sprintf("## %s\n\n", f.Name))
			fac.WriteString(fmt.Sprintf("**Leader:** %s  \n", f.Leader))
			fac.WriteString(fmt.Sprintf("**Obiettivo:** %s  \n", f.Goal))
			fac.WriteString(fmt.Sprintf("**Atteggiamento iniziale:** %s  \n\n", f.Attitude))
			fac.WriteString("### Segreto\n\n" + f.Secret + "\n\n")
		}
		if err := cm.saveWorldFile(slug, "factions.md", fac.String()); err != nil {
			return err
		}
	}

	// ── Update campaign.json with world info and ascii art ────────────────
	c.WorldName    = world.WorldName
	c.CoverAscii   = world.AsciiArt
	c.CoverPrompt  = world.CoverPrompt
	c.LastSummary  = world.Premise
	c.UpdatedAt    = time.Now()

	return cm.save(c)
}

// saveWorldFile saves a file under data/campaigns/{slug}/world/.
func (m *Manager) saveWorldFile(slug, filename, content string) error {
	return m.saveSubFile(slug, "world", filename, content)
}

// saveNPCFile saves a file under data/campaigns/{slug}/npcs/.
func (m *Manager) saveNPCFile(slug, filename, content string) error {
	return m.saveSubFile(slug, "npcs", filename, content)
}

func (m *Manager) saveSubFile(slug, subdir, filename, content string) error {
	dir := m.campaignDir(slug) + "/" + subdir
	if err := mkdirAll(dir); err != nil {
		return err
	}
	return writeFile(dir+"/"+filename, []byte(content))
}

func slugifyName(name string) string {
	s := strings.ToLower(name)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, s)
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}
	return strings.Trim(s, "_")
}
