# 🎲 Oracle DnD — AI Dungeon Master

Web app Go + HTML/CSS/JS vanilla alimentata da **Ollama** (modelli locali).

## Requisiti

- Go 1.22+
- [Ollama](https://ollama.ai) in esecuzione su `http://localhost:11434`
- Almeno un modello scaricato, es: `ollama pull deepseek-r1:8b`

## Avvio rapido

```bash
cp config.example.json config.json
go run ./cmd/server
```

Il server parte su **http://localhost:8080**.

## Config

```json
{
  "port": "8080",
  "ollama_url": "http://localhost:11434",
  "default_model": "deepseek-r1:8b"
}
```

## API

### Campagne
| Metodo | Endpoint | Descrizione |
|--------|----------|-------------|
| `GET`    | `/api/campaign/list`            | Lista campagne |
| `POST`   | `/api/campaign/create`          | Crea campagna |
| `GET`    | `/api/campaign/{slug}`          | Dettaglio campagna |
| `PATCH`  | `/api/campaign/{slug}`          | Aggiorna metadati |
| `DELETE` | `/api/campaign/{slug}`          | Elimina |
| `POST`   | `/api/campaign/{slug}/archive`  | Archivia |
| `POST`   | `/api/campaign/{slug}/complete` | Segna completata |

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

## Struttura

```
dnd-dm/
├── cmd/server/main.go
├── internal/
│   ├── ai/           # Client Ollama + SSE + tag parser
│   ├── campaign/     # CRUD campagne + persistenza JSON
│   ├── config/       # Caricamento config.json
│   ├── dice/         # Roller thread-safe con log
│   ├── handlers/     # Handler HTTP (chat, campagne, dadi)
│   └── session/      # Stato sessione + auto-summary
├── web/
│   ├── templates/    # HTML pages
│   └── static/       # CSS dark fantasy + JS vanilla
├── data/campaigns/   # Generato a runtime (gitignored)
├── config.example.json
└── go.mod
```

## Roadmap

- [x] **Parte 1** — Backend routing + Ollama client + SSE
- [x] **Parte 2** — Hub Campagne + gestione campagne multiple
- [ ] Parte 3 — Generazione campagna AI + storage Markdown
- [ ] Parte 4 — Creazione personaggio
- [ ] Parte 5 — Interfaccia sessione di gioco
- [ ] Parte 6 — Mappa procedurale
- [ ] Parte 7 — Multiplayer WebSocket
- [ ] Parte 8 — Export/Import
- [ ] Parte 9 — Voice acting Piper
