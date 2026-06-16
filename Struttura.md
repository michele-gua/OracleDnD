# 🎲 Oracle DnD — Struttura del Progetto

## Avvio rapido

```bash
cp config.example.json config.json
go run ./cmd/server
```

Server su **http://localhost:8080**. Richiede Ollama attivo con almeno un modello.

## Config

```json
{
  "port": "8080",
  "ollama_url": "http://localhost:11434",
  "default_model": "deepseek-r1:8b",
  "rulesets_dir": "./rulesets"
}
```

## API

### Campagne
| Metodo | Endpoint | Descrizione |
|--------|----------|-------------|
| `GET`    | `/api/campaign/list`              | Lista campagne |
| `POST`   | `/api/campaign/create`            | Crea campagna |
| `POST`   | `/api/campaign/generate`          | **Genera mondo AI (SSE)** |
| `GET`    | `/api/campaign/{slug}`            | Dettaglio |
| `PATCH`  | `/api/campaign/{slug}`            | Aggiorna metadati |
| `DELETE` | `/api/campaign/{slug}`            | Elimina |
| `POST`   | `/api/campaign/{slug}/archive`    | Archivia |
| `POST`   | `/api/campaign/{slug}/complete`   | Completa |
| `GET`    | `/api/campaign/{slug}/files`      | Lista file markdown |
| `GET`    | `/api/campaign/{slug}/file?path=` | Leggi file markdown |

### Rulesets
| Metodo | Endpoint | Descrizione |
|--------|----------|-------------|
| `GET`  | `/api/rulesets`        | Lista regolamenti disponibili |
| `POST` | `/api/ruleset/select`  | Cambia regolamento sessione |

### AI & Chat
| Metodo | Endpoint | Descrizione |
|--------|----------|-------------|
| `POST` | `/api/chat`          | Chat DM (SSE streaming) |
| `GET`  | `/api/models`        | Lista modelli Ollama |
| `POST` | `/api/model/select`  | Cambia modello sessione |

### Dadi
| Metodo | Endpoint | Descrizione |
|--------|----------|-------------|
| `POST` | `/api/roll`         | Tira dadi (es. `2d6+3`) |
| `GET`  | `/api/roll/suggest` | Suggerimenti contestuali |

## Struttura file

```
dnd-dm/
├── cmd/server/main.go
├── internal/
│   ├── ai/
│   │   ├── client.go        # Client Ollama + streaming SSE
│   │   └── tags.go          # Parser tag [ROLL:] [XP:] ecc.
│   ├── campaign/
│   │   ├── campaign.go      # CRUD campagne + persistenza JSON
│   │   ├── generator.go     # Generazione mondo AI + salvataggio markdown
│   │   └── fs.go            # Helpers OS
│   ├── config/config.go
│   ├── dice/roller.go
│   ├── handlers/
│   │   ├── handlers.go          # Chat, modelli, dadi
│   │   ├── campaign_handlers.go # CRUD campagne
│   │   ├── generate_handlers.go # Generazione AI + rulesets
│   │   └── file_handlers.go     # Lettura file markdown
│   ├── ruleset/ruleset.go   # Loader JSON regolamenti
│   └── session/session.go
├── web/
│   ├── templates/
│   │   ├── campaigns.html   # Hub campagne
│   │   ├── manage.html      # Gestione campagna
│   │   └── generate.html    # Generazione mondo AI
│   └── static/
│       ├── css/
│       │   ├── main.css      # Design system dark fantasy
│       │   ├── campaigns.css
│       │   ├── manage.css
│       │   └── generate.css
│       └── js/
│           ├── campaigns.js
│           ├── manage.js
│           └── generate.js
├── rulesets/
│   ├── dnd5e.json
│   ├── dnd5e2024.json
│   ├── dnd35e.json
│   └── pathfinder.json
├── data/campaigns/          # Generato a runtime (gitignored)
├── config.example.json
├── go.mod
└── Struttura.md
```

## Roadmap

- [x] **Parte 1** — Backend routing + Ollama client + SSE
- [x] **Parte 2** — Hub Campagne + gestione campagne multiple
- [x] **Parte 3** — Generazione campagna AI + storage Markdown + rulesets
- [ ] Parte 4 — Creazione personaggio
- [ ] Parte 5 — Interfaccia sessione di gioco
- [ ] Parte 6 — Mappa procedurale
- [ ] Parte 7 — Multiplayer WebSocket
- [ ] Parte 8 — Export/Import
- [ ] Parte 9 — Voice acting Piper
