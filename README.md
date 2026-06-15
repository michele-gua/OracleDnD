# OracleDnD
A web-app made entirely with AI to play DnD (even with friends) with an AI powered Dungeon Master.

🇮🇹 
Ho usato Claude per scrivere un prompt per incollarlo in un'altra sessione di Claude per creare questa app, tutte le features che l'app avra' sono elencate nel prompt qui sotto:
# PROMPT: D&D AI Dungeon Master Web App in Go

## Contesto del progetto

Crea una web app completa in **Go** (backend) + HTML/CSS/JS vanilla (frontend) che funziona da **Dungeon Master AI** per D&D, alimentata da **Ollama** con modelli locali. L'AI non è solo il narratore — è il **creatore completo del mondo**: genera la campagna da zero, popola il mondo di NPC, dungeon, fazioni e trame prima ancora che il giocatore muova il primo passo.

---

## Stack tecnico

- **Backend:** Go (net/http standard library, no framework)
- **Frontend:** HTML/CSS/JS vanilla, single-page app con fetch API
- **AI:** Ollama API locale (`http://localhost:11434`) — streaming support
- **Storage:** File JSON locali (tecnico) + file Markdown/TXT (narrativo leggibile)
- **Modello default:** `deepseek-r1:8b` (reasoning, veloce, leggero)
- **Modelli alternativi selezionabili:** `gpt-oss:120b`, `qwen3-coder:30b`, `llava:latest` — dropdown nella UI per cambiarli al volo
- **TTS locale:** integrazione opzionale con Piper o Kokoro per voice acting
- **Porta:** `8080`

---

## Struttura directory

```
dnd-dm/
├── cmd/server/main.go
├── internal/
│   ├── ai/            # client Ollama, streaming, prompt builder
│   ├── ruleset/       # caricamento e gestione regolamenti
│   ├── character/     # creazione, validazione, storage personaggi
│   ├── session/       # stato sessione, storia, inventario, world state
│   ├── dice/          # roller dadi con seed e log
│   ├── world/         # mappa procedurale, bestiario, lore journal, timeline mondo, side quest
│   ├── npc/           # memoria NPC, reputazione, relazioni
│   ├── companion/     # generazione e gestione compagni AI (single-player)
│   ├── qr/            # generazione QR code (campagna, PG, locandina, companion)
│   ├── progression/   # XP tracking, level up, milestone system
│   ├── campaign/      # generazione campagna AI, world building iniziale
│   └── multiplayer/   # WebSocket hub, gestione party reale
├── web/
│   ├── templates/     # HTML Go templates
│   └── static/        # CSS, JS, immagini, suoni
├── rulesets/          # file JSON dei regolamenti
│   ├── dnd5e.json
│   ├── dnd5e2024.json
│   ├── dnd35e.json
│   └── pathfinder.json
└── data/
    └── campaigns/
        └── {campaign_name}/
            ├── campaign.json
            ├── campaign_overview.md
            ├── characters/
            ├── companions/
            ├── sessions/
            ├── world/
            ├── npcs/
            └── items/
```

---

## Gestione Campagne Multiple — Save Slot

L'app supporta un numero illimitato di campagne salvate contemporaneamente, ognuna completamente indipendente dalle altre: mondo, personaggi, NPC, lore, sessioni e mappa sono separati per campagna. Il giocatore può passare da una campagna all'altra liberamente senza perdere nulla.

### Schermata Hub Campagne (`/campaigns`)

È la schermata principale dell'app, mostrata sempre dopo il login/avvio. Funziona come una collezione di **save slot** visivi:

- Griglia di card, una per campagna salvata
- Ogni card mostra:
  - **Nome campagna** e **nome del mondo**
  - **Ruleset** usato (icona + nome)
  - **Personaggi attivi** (avatar/icone con nome e classe)
  - **Numero sessioni** completate e durata totale giocata
  - **Data e ora dell'ultima sessione**
  - **Ultima riga del summary** dell'ultima sessione (anteprima narrativa)
  - **Stato:** Attiva / Completata / Abbandonata (settabile manualmente)
  - **Immagine di copertina** generata dall'AI come prompt testuale per Stable Diffusion o, in alternativa, un'illustrazione ASCII stilizzata generata dall'AI stessa
- Bottoni su ogni card: **Riprendi** / **Gestisci** / **Esporta** / **Elimina**
- Bottone principale in alto: **+ Nuova Campagna**
- Barra di ricerca e filtri: per ruleset, stato, data, numero sessioni
- Ordinamento: per data ultimo accesso (default), per nome, per durata totale

### Indipendenza totale tra campagne

Ogni campagna è una cartella separata in `data/campaigns/{campaign_slug}/`. Il mondo, la trama, gli NPC, le fazioni, la mappa, il lore e la reputazione sono completamente indipendenti tra campagne — nulla di tutto questo passa automaticamente da una campagna all'altra. L'unica cosa condivisa è il **personaggio**: se il giocatore lo porta in una nuova campagna tramite export `.dndch` e import, mantiene la sua identità narrativa, il **livello globale** e il suo **inventario personale semi-globale** (vedi sezione "Personaggi cross-campagna" per i dettagli su cosa persiste e cosa viene resettato).

### Gestione campagna (`/campaign/{name}/manage`)

Pagina di gestione per ogni campagna con:
- Rinomina campagna / mondo
- Cambia immagine di copertina
- Aggiungi/rimuovi personaggi dal party
- Archivia campagna (nasconde dalla griglia principale ma non elimina)
- Segna come completata (con data di completamento e note finali)
- Statistiche complete: sessioni, tempo totale, mostri, dadi tirati, achievement
- Lista sessioni con possibilità di rileggere transcript e summary di ognuna
- Export campagna in `.dndca` o `.dndx`

### Personaggi cross-campagna

Un personaggio creato per una campagna può essere esportato come `.dndch` e importato in un'altra campagna. L'app gestisce la conversione automatica se il ruleset è diverso tra le campagne (con avviso e lista dei campi non compatibili). Il personaggio mantiene il suo backstory, la sua storia, i suoi tratti narrativi, il **livello (globale al personaggio, vedi sezione Livelli e Progressione)** e il suo **inventario personale semi-globale** (vedi sotto). **La reputazione e tutto ciò che è legato a NPC/fazioni della campagna di origine vengono resettati** all'import: non hanno senso in un mondo diverso. Il personaggio riparte al livello già raggiunto (entro i cap della nuova campagna, vedi sotto) e con l'inventario personale aggiornato all'ultimo stato della campagna precedente.

### Inventario semi-globale

L'inventario del personaggio è **semi-globale**: persiste tra campagne, ma con una distinzione netta tra due categorie di oggetti.

- **Oggetti personali (persistono):** armi, armature, amuleti, pozioni, oro e qualsiasi oggetto acquisito, comprato, craftato o trovato che appartiene al personaggio in quanto tale. Se durante una campagna l'arma del personaggio si rompe e viene sostituita, l'inventario riflette quel cambiamento — il personaggio porterà con sé la **nuova** arma (o nessuna, se non sostituita) nella prossima campagna, non quella originale perduta. L'inventario è quindi uno snapshot dello stato corrente, non un log storico da "ripristinare".
- **Oggetti campaign-bound (NON persistono):** oggetti legati a quest, chiavi, McGuffin narrativi, oggetti rubati a NPC/fazioni specifiche, ricompense di trama il cui significato esiste solo nel contesto di quella campagna. Questi oggetti vengono marcati come `campaign_bound: true` al momento della creazione/assegnazione (dal DM AI o manualmente) e **rimossi automaticamente dall'inventario all'export/import** tra campagne.
- L'oro (`PP/PO/PE/PA/PR`) è considerato oggetto personale e persiste, salvo l'opzione del giocatore di azzerarlo all'import (per campagne "da zero" volute esplicitamente).
- Alla creazione di un nuovo oggetto, il DM AI o il sistema di loot determina automaticamente se marcarlo `campaign_bound` (es. "Sigillo della Torre Nera" → sì; "Spada lunga +1" → no); il giocatore può correggere manualmente la marcatura dalla scheda inventario.
- All'import in una nuova campagna, viene mostrato un riepilogo: "Porterai con te: [lista oggetti personali]. Rimossi perché legati alla campagna precedente: [lista oggetti campaign-bound]."

---

## Flusso Nuovo Giocatore — Onboarding Completo

Questo è il flusso che un nuovo utente percorre per creare una nuova campagna. Ogni step è una schermata dedicata.

### Step 1 — Benvenuto e Ruleset

- Schermata di benvenuto con breve spiegazione di cosa è D&D e come funziona l'app
- Selezione ruleset tra:
  - **D&D 5e (2014)** — *consigliato per principianti*
  - **D&D 5e 2024 (One D&D)**
  - **D&D 3.5e**
  - **Pathfinder 1e / 2e**
- Badge "Consigliato per chi inizia" su D&D 5e
- Tooltip su ogni opzione che spiega brevemente le differenze

### Step 2 — Creazione Personaggio

Obbligatoria prima di creare la campagna. Form guidato step-by-step:

1. **Nome e genere**
2. **Razza** — descrizione narrativa e bonus meccanici spiegati in semplice
3. **Classe** — spiegazione del ruolo in party (tank, healer, damage, support)
4. **Statistiche** — 3 metodi a scelta:
   - *Roll 4d6 drop lowest* (classico, consigliato)
   - *Standard array* (bilanciato)
   - *Point buy* (ottimizzazione)
   - Tooltip che spiega ogni statistica in termini concreti
5. **Abilità e proficienze** — con spiegazione di quando si usa ciascuna
6. **Equipaggiamento iniziale** — dalla classe o random, con descrizione oggetti
7. **Allineamento** — con spiegazione narrativa
8. **Tratti di personalità, ideali, legami, difetti** — suggeriti dall'AI
9. **Backstory** — tre modalità:
   - Scrivi tu liberamente
   - Inserisci 3 parole chiave → l'AI genera il background completo
   - "Sorprendimi" → l'AI genera tutto casualmente ma coerente con razza/classe
10. **Icona del personaggio** — immagine/avatar che rappresenta il PG sulla mappa:
    - **Upload personalizzato:** il giocatore importa un'immagine propria (PNG/JPG/SVG, ritagliata automaticamente in icona quadrata/circolare)
    - **Generata dall'AI:** prompt testuale per Stable Diffusion basato su razza/classe/aspetto, oppure illustrazione ASCII/SVG stilizzata generata dall'AI
    - **Preset:** scelta da una libreria di icone base per razza/classe
    - L'icona è salvata in `characters/{nome}/icon.*` ed è **portabile** insieme al personaggio negli export `.dndch`

Alla fine: scheda personaggio completa in stile D&D, bottone "Salva e continua".

### Step 3 — Creazione Campagna AI

Dopo aver salvato il personaggio, il giocatore sceglie:

**Opzione A — Genera campagna automatica:**

Form con domande narrative:
- *"Che tipo di avventura vuoi?"* → Fantasy classico / Dark & Gritty / Politico e intrighi / Horror / Commedia / Mare e pirati / Deserto e antiche rovine / Lascia decidere all'AI
- *"Quanto vuoi che duri la campagna?"* → One-shot / Breve (3-5 sessioni) / Media (10-15) / Epica (20+)
- *"Tono della storia?"* → Eroico / Moralmente grigio / Oscuro / Leggero
- *"Hai preferenze sul mondo?"* → campo testo libero opzionale
- *"Livello di difficoltà combattimenti?"* → Facile / Normale / Difficile / Mortale
- *"Range di livello della campagna?"* → Novizi (1-5) / Avventurieri (1-10) / Eroi (5-15) / Epica (10-20) / Leggendaria (1-20) / Custom (min/max manuali) — definisce il `level_cap_min`/`level_cap_max` (vedi sezione Livelli e Progressione)
- *"Sistema di progressione?"* → XP classici (accumulo punti esperienza per scontri/eventi) / Milestone (livello su a eventi narrativi chiave decisi dal DM)
- *"Vuoi giocare con dei compagni di squadra?"* → Sì, l'AI genera 3 compagni / No, gioco da solo (solo per campagne single-player)

**Generazione mondo — coerenza prima di tutto:**

L'AI **non inventa la trama "man mano"**. Prima che il giocatore muova il primo passo, l'AI deve generare e fissare su disco una **trama principale completa end-to-end**: setup, sviluppo, climax e risoluzione (in atti, anche se abbozzati), inclusi gli eventi che accadranno **indipendentemente dalle azioni del giocatore**. Questo è essenziale per la coerenza narrativa: generare eventi solo al momento porta a contraddizioni, NPC che dimenticano motivazioni, trame che si sfaldano.

Concretamente, l'AI genera in anticipo:
- L'intera arco narrativo principale (atti, eventi chiave, twist, finale/i possibili)
- Una **timeline del mondo indipendente** (`world_timeline.md`): cosa fanno villain, fazioni e NPC chiave session per session/col passare del tempo, **anche se il giocatore non interviene** — il mondo è vivo e prosegue
- 3-5 side quest collegate a personaggi, fazioni o location della trama principale, ognuna con un **punto di ricongiungimento** alla main plot definito in anticipo

Durante il gioco l'AI non "inventa da zero" — consulta e adatta questi materiali pre-generati. Eventuali nuovi dettagli emersi in sessione (improvvisazioni necessarie) vengono sempre annotati e **riconciliati** con la timeline esistente per evitare incongruenze future.

L'AI genera un pacchetto mondo completo mostrato con progress bar e messaggi narrativi (*"Sto forgiando il mondo..."*, *"Popolo le taverne di volti familiari..."*, *"Nascondo i segreti nelle ombre..."*):

1. Nome e concept del mondo/regione
2. La città/luogo di partenza con descrizione
3. La trama principale (hook narrativo in 3 atti **completi**, inclusi eventi chiave, climax e possibili risoluzioni — non solo abbozzati)
4. **Timeline del mondo** — cosa accade indipendentemente dal giocatore (mosse del villain, evoluzione delle fazioni, eventi globali) in corrispondenza di milestone temporali/sessioni
5. 3-5 side quest, ognuna ancorata a un NPC/fazione/location della trama principale e con un **punto di ricongiungimento** esplicito alla main plot
6. 8-12 NPC pre-generati con nome, ruolo, personalità, segreto
7. La mappa iniziale (area di partenza + aree adiacenti conosciute)
8. Il villain/antagonista principale con motivazioni non banali e un proprio piano che evolve nella timeline
9. 3 dungeon/location chiave abbozzate
10. 2-3 fazioni con obiettivi in conflitto
11. Hook iniziale personalizzato sul backstory del personaggio
12. Se richiesto, **3 compagni di squadra AI**: nome, classe, personalità, ruolo nel party, relazione con il protagonista e arco narrativo personale che si intreccia con la main plot

Tutti questi dati vengono salvati immediatamente nei file Markdown della campagna e nella griglia delle campagne compare la nuova card.

**Opzione B — Campagna personalizzata:**
- Editor libero dove il giocatore descrive il mondo
- L'AI integra, espande e chiede chiarimenti se necessario
- Stesso output finale dell'Opzione A

**Opzione C — Riprendi campagna esistente:**
- Reindirizza all'Hub Campagne con la lista delle campagne salvate

---

## Interfaccia di Gioco

Layout a due colonne:
- **Sinistra:** chat con il DM (stile pergamena, streaming)
- **Destra:** pannello con tab — Personaggio / Inventario / Mappa / NPC / Lore

Il DM risponde in **streaming** token per token con effetto macchina da scrivere.

**Quick actions** sotto l'input: Attacca / Esamina / Parla / Riposa / Fuggi / Cerca

**Undo narrativo:** bottone "E se avessi scelto diversamente?" che torna indietro di N messaggi e salva il branch alternativo separatamente.

**Modalità Solo Rapido:** personaggio pregenerato, avventura one-shot 30 minuti, nessun setup.

**Cambio campagna rapido:** dropdown in header che mostra tutte le campagne salvate con l'ultima riga del loro summary — permette di passare da una campagna all'altra senza tornare all'Hub.

---

## Dice Box — Sistema Dadi Visuale

Pannello sempre accessibile come drawer laterale o fisso:

- **Dadi fisici** in CSS/SVG cliccabili: d4, d6, d8, d10, d12, d20, d100
- Click su dado → aggiunto alla mano corrente (es. 2d6 + d4)
- Campo modificatore +/- numerico
- Bottone **"Lancia!"** con animazione shake 3D + risultato in grande
- Log ultimi 10 tiri visibile sotto

**Tendina "Lanci Suggeriti"** sopra la dice box:
- Popolata automaticamente dal DM in base alla situazione corrente
- Click → pre-popola e lancia automaticamente
- Aggiornata tramite tag `[SUGGEST_ROLL: descrizione|formula]` nelle risposte AI

**Tab "Guida ai Tiri" — Sezione Newbie:**
- Spiega in linguaggio semplice e colloquiale:
  - Quando si usa ogni dado
  - Cosa sono i modificatori e come si calcolano
  - Differenza tra tiro per colpire, tiro danno, tiro salvezza, prova di caratteristica
  - Come funziona vantaggio/svantaggio
- Esempi pratici sempre visibili:
  - *"Vuoi colpire un nemico? → 1d20 + bonus attacco. Se superi la CA del nemico, hai colpito!"*
  - *"Vuoi scassinare una serratura? → 1d20 + Destrezza + eventuale competenza"*
  - *"Ti lanciano un incantesimo? → il DM ti dirà che tiro fare per resistere"*
- Bottone **"Chiedi al DM cosa tirare"** → spiegazione contestuale alla situazione attuale
- **Glossario rapido:** CA, PF, CD, tiro salvezza, prova di caratteristica — ognuno in 2 righe senza gergo

---

## Sistema di Combattimento

Quando il DM rileva combattimento, pannello laterale dedicato:
- Ordine iniziativa drag-and-drop
- HP nemici (stimati dal DM, aggiornabili manualmente)
- Round counter e log azioni per round
- **Condition tracker:** avvelenato, stordito, invisibile, concentrazione, ecc. con durata in round e reminder automatico
- **Spell slot tracker** visivo per caster (bollini consumabili)
- Il DM gestisce i turni nemici autonomamente

---

## Livelli e Progressione

### Il livello è globale al personaggio

Il livello (e l'XP, se sistema XP classici) appartiene al **personaggio**, non alla singola campagna: è un dato di `{nome}.json`/`{nome}.md` nella cartella `characters/`, indipendente da `campaign.json`. Se un personaggio sale di livello giocando una campagna, quel livello resta acquisito anche se il personaggio viene esportato (`.dndch`) e importato in una nuova campagna — coerente con la sezione "Personaggi cross-campagna". Una nuova campagna può quindi iniziare con un personaggio già di livello 3 (o superiore), se il giocatore lo importa da una campagna precedente.

### Due sistemi alternativi, scelti all'inizio della campagna

- **XP classici:** il DM AI assegna punti esperienza dopo ogni scontro, prova superata o evento narrativo significativo, secondo le tabelle del ruleset selezionato. Il totale XP è visibile nella scheda personaggio con barra di progresso verso il livello successivo.
- **Milestone:** nessun conteggio XP. Il level up avviene quando il DM determina che il party ha raggiunto un punto narrativo chiave (fine di un atto, sconfitta di un boss, completamento di un arco di trama). Il DM annuncia il level up in modo narrativo.

Il sistema scelto è fissato per tutta la campagna e salvato in `campaign.json`.

### Cap di livello per campagna

Ogni campagna definisce un **livello minimo e massimo** (`level_cap_min` / `level_cap_max` in `campaign.json`), parte delle domande di creazione campagna (Step 3):

- *"Range di livello della campagna?"* → preset rapidi (es. Novizi 1-5 / Avventurieri 1-10 / Eroi 5-15 / Epica 10-20 / Leggendaria 1-20) oppure range custom
- **Livello minimo:** se un personaggio importato è sotto il minimo, viene proposto di portarlo al minimo della campagna (level up "gratuito" applicato all'import, con le scelte guidate dall'AI come un normale level up)
- **Livello massimo:** funziona da cap di difficoltà — superato il massimo, il personaggio non guadagna più livelli (XP eccedenti non assegnati, o ignorati in modalità milestone); utile per campagne "impossibili"/contenute dove il livello resta basso anche con personaggi importati di livello alto
- Se un personaggio importato supera il massimo della campagna, viene proposto di abbassarlo al cap (con conferma esplicita, e nota che si tratta di una scelta volontaria del giocatore per quella campagna)
- Il range scelto influenza direttamente il **difficulty scaling** (vedi sotto): combinato con la difficoltà Facile/Normale/Difficile/Mortale, definisce il tipo di campagna (es. range basso + Mortale = "campagna per novizi ma spietata"; range alto + Facile = "power fantasy epica")

### Tracking XP automatico (modalità XP classici)

- Dopo ogni scontro o evento rilevante, il DM AI calcola e assegna XP usando il tag `[XP: quantità|motivo]`
- Gli XP vengono mostrati come notifica discreta ("+50 XP — Goblin sconfitto") e aggiunti al totale del personaggio
- Quando la soglia di livello viene superata, scatta automaticamente il level up (entro il cap massimo della campagna)

### Level up guidato dall'AI

- Alla soglia di livello (XP o milestone), il DM interrompe la narrazione con una schermata/modal di level up
- L'AI spiega in linguaggio semplice, specifico per la classe del personaggio, cosa cambia: nuovi PF massimi, nuove proficiency, incantesimi disponibili, tratti di classe/sottoclasse, eventuale scelta di sottoclasse o ASI (Ability Score Improvement)
- Per scelte multiple (sottoclasse, incantesimi da imparare, ASI vs Feat), l'AI presenta le opzioni con breve descrizione dell'impatto in gioco prima di lasciare scegliere al giocatore
- La scheda personaggio viene aggiornata automaticamente e il file Markdown del personaggio registra il level up con data/sessione/campagna

### Difficulty scaling basato sul livello

- Il livello (e la dimensione) del party è un input diretto per il DM AI nella generazione di nemici, incontri e ricompense
- I nemici generati per scontri vengono scalati in CR/difficoltà coerentemente col livello medio del party, il range `level_cap_min`/`level_cap_max` della campagna, e la difficoltà scelta in campagna (Facile/Normale/Difficile/Mortale)
- Quando il party sale di livello, gli incontri futuri vengono generati tenendo conto del nuovo potere del party — evitando sia scontri triviali sia TPK accidentali per scarto di livello
- Il livello del party è incluso nel system prompt (`{PARTY_DESCRIPTION}`) e considerato dal DM prima di ogni generazione di nemici o incontro

---

## World Building Dinamico

### Mappa Procedurale
- Generata parzialmente alla creazione campagna (area iniziale + dintorni)
- Espansa on-demand quando il party esplora aree nuove
- **Rendering grafico, non ASCII:** la mappa è un'immagine (SVG generato proceduralmente lato server, o illustrazione raster via Stable Diffusion/modello locale) — stanze/aree come tile o forme disegnate, con stile coerente al tema dark fantasy
- Ogni stanza/area generata ha una propria piccola illustrazione o tile associato (generata dall'AI in base alla descrizione narrativa), composta nella mappa complessiva
- Fog-of-war progressivo con animazione CSS (overlay semi-trasparente sulle aree non scoperte)
- Icone per: nemici sconfitti, tesori, uscite, punti interesse, NPC — sovrapposte come layer grafico sulla mappa
- **Icona personaggio (token):** l'icona/avatar di ogni PG scelta in creazione personaggio è renderizzata come token sulla mappa, posizionata nella stanza/area corrente; si sposta con animazione quando il party si muove tra stanze/aree (anche per i compagni AI e gli altri giocatori in multiplayer, ognuno col proprio token)
- Click su stanza visitata → riassunto di cosa è successo lì
- Aggiornata incrementalmente (nuove stanze/tile aggiunti alla mappa esistente), mai rigenerata da zero
- La mappa è salvata come SVG (struttura) + asset immagine delle singole stanze/tile in `world/map_assets/`

### Bestiario Dinamico
- Ogni mostro incontrato aggiunto al **bestiario di campagna** — ciò che il party nel suo insieme ha scoperto in quella storia
- Scheda: statistiche, tattiche osservate, debolezze scoperte, lore
- Consultabile dal menu laterale durante il gioco
- **Bestiario per personaggio:** sottoinsieme del bestiario di campagna, filtrato in base a cosa quel singolo PG era presente a scoprire — utile in multiplayer quando il party si divide e personaggi diversi incontrano mostri diversi
- Nessun bestiario globale cross-campagna: un personaggio non può "conoscere" mostri mai incontrati nella sua storia

### Lore Journal
- Popolato automaticamente dal DM tramite tag `[LORE: categoria|voce|descrizione]`
- Strutturato in categorie consultabili
- Esportabile come PDF narrativo

---

## Trama Persistente, Mondo Vivo e Compagni AI

### Principio cardine: la trama è scritta prima, non improvvisata

La coerenza narrativa è la priorità assoluta. L'AI **non deve inventare la storia "sul momento"** — farlo produce contraddizioni, NPC con motivazioni incoerenti, trame che si sfaldano dopo poche sessioni. Per questo, **tutta la trama principale viene generata e fissata su disco durante la creazione campagna**, prima che il giocatore muova il primo passo (vedi Step 3 — Creazione Campagna AI). Durante il gioco il DM **consulta e adatta** questo materiale, non lo inventa da zero.

### Il mondo prosegue anche senza il giocatore

La trama principale e il mondo **non aspettano il giocatore**. Anche se il party si allontana per esplorare, fare side quest, o ignora completamente l'hook principale per intere sessioni, gli eventi della `world_timeline.md` **continuano ad accadere**:

- Il villain prosegue il suo piano (es. recluta alleati, completa un rituale, conquista territorio)
- Le fazioni avanzano i loro obiettivi, anche in conflitto tra loro
- NPC chiave agiscono autonomamente in base alla loro agenda

Queste evoluzioni sono pre-scritte come **trigger temporali/di sessione** nella timeline (es. "entro la sessione 5, se il party non ha agito, il villain conquista la torre est"), così il DM AI sa sempre cosa succede "fuori scena" senza doverlo inventare al momento. Quando il party torna sulla trama principale, trova un mondo **conseguente alle proprie scelte e alla propria inazione** — questo crea peso reale alle decisioni senza richiedere improvvisazione.

### Side quest sempre ricollegate alla main plot

Le side quest (pre-generate o create al volo quando il party si allontana) **non sono mai completamente slegate**: ognuna ha un punto di ricongiungimento esplicito con la trama principale — un indizio, un NPC condiviso, un oggetto, una rivelazione che rimanda al villain o alla trama centrale. Il giocatore può divagare quanto vuole, ma ogni divagazione lo riporta, prima o poi, verso il filo principale. Questi collegamenti sono annotati in `side_quests.md` e il DM li richiama naturalmente quando narrativamente opportuno.

### Compagni di squadra AI (solo single-player)

Se scelto in onboarding, l'AI genera **3 compagni di squadra** che accompagnano il protagonista per tutta la campagna:

- Ognuno ha scheda completa (classe, ruolo nel party, statistiche), personalità distinta, relazione iniziale col protagonista e un proprio arco narrativo che si intreccia con la main plot
- Durante l'esplorazione e la narrazione, i compagni intervengono con dialoghi, opinioni e reazioni proprie (il DM AI li interpreta)
- In combattimento, le loro azioni sono gestite dal DM AI in modo tatticamente sensato ma il giocatore può dare indicazioni di massima ("copritemi", "concentratevi sul caster")
- Gli archi dei compagni evolvono nel tempo (rivelazioni, conflitti, crescita) e possono intersecare la trama principale o le side quest
- **Alternativa:** il giocatore può scegliere di giocare **completamente da solo**, senza compagni — in tal caso il DM scala gli incontri di conseguenza (vedi Difficulty scaling)

### Tono del DM: imparità, empatia e sfida

Il DM AI mantiene un equilibrio costante tra tre elementi:

- **Imparzialità:** il DM non favorisce né penalizza il giocatore arbitrariamente. Le conseguenze (positive o negative) derivano dalle scelte e dai tiri, non dall'umore della narrazione. Il villain e il mondo agiscono secondo logica propria, non per "punire" o "premiare" il giocatore fuori contesto.
- **Sfida:** la campagna deve restare stimolante — combattimenti che richiedono strategia, scelte con conseguenze reali, un mondo che non si piega passivamente al giocatore. Il divertimento nasce anche dalla tensione e dal rischio, non solo dalla narrazione piacevole.
- **Empatia nei momenti di bisogno:** se il giocatore è in difficoltà (frustrazione evidente, situazione di gioco che diventa opprimente, richiesta esplicita di aiuto), il DM può offrire un indizio più chiaro, un momento di respiro narrativo, o un'opzione di fuga plausibile — senza però annullare la sfida complessiva o "salvare" sistematicamente il giocatore dalle conseguenze delle sue scelte. L'empatia è un aggiustamento di tono nei momenti critici, non un cambio di filosofia della campagna.

---

## Memoria NPC e Reputazione

- Ogni NPC salvato con: nome, aspetto, personalità, storia, segreto
- Gli NPC pre-generati alla creazione campagna hanno già il loro file, attivato al primo incontro
- **Memoria interazioni:** azioni del party, promesse, torti, favori
- Stato relazione: Ostile / Neutrale / Amichevole / Alleato
- **Reputation tracker** per fazione: influenza le reazioni degli NPC
- **Allineamento tracker:** il DM traccia le scelte morali silenziosamente, notifica quando cambia

---

## Inventario ed Economia

- Inventario visivo drag-and-drop con slot equipaggiamento
- Registro monete PP/PO/PE/PA/PR con conversione automatica
- Oggetti magici con proprietà e attunement tracker
- **Bottino automatico:** quando il DM descrive un tesoro, propone di aggiungerlo con un click
- **Crafting system:** raccolta ingredienti e creazione oggetti secondo il ruleset
- Valutazione automatica valore oggetti in PO
- **Flag `campaign_bound`:** ogni oggetto ha un flag che determina se è personale (persiste tra campagne) o legato alla campagna corrente (rimosso all'export/import) — assegnato automaticamente dal DM AI al momento del loot/creazione, modificabile manualmente dal giocatore (vedi "Inventario semi-globale")

---

## Storage Narrativo — File Leggibili

Ogni campagna genera automaticamente una cartella di file Markdown leggibili anche senza l'app aperta.

### Struttura

```
data/campaigns/{campaign_name}/
├── campaign_overview.md        # mondo, trama, tono, fazioni — creato alla generazione
├── characters/
│   ├── {nome}.md               # scheda narrativa personaggio
│   ├── {nome}.json             # dati tecnici (stats, livello, inventario, icona)
│   └── {nome}_diary.md         # diario personale in prima persona, generato dall'AI
├── sessions/
│   └── session_01_YYYY-MM-DD/
│       ├── full_transcript.md  # trascritto integrale in append real-time
│       ├── summary.md          # riassunto "Previously on..." generato dall'AI
│       └── events.md           # bullet list eventi chiave
├── world/
│   ├── lore.md                 # lore per categoria, aggiornato in real-time
│   ├── bestiary.md             # mostri incontrati con note
│   ├── map.svg                 # struttura mappa corrente (stanze, connessioni, fog-of-war)
│   ├── map_assets/              # illustrazioni/tile delle singole stanze/aree
│   ├── map_log.md              # log testuale aree esplorate (per lore/coerenza)
│   ├── factions.md             # fazioni e reputazione party
│   ├── world_timeline.md       # eventi del mondo indipendenti dal giocatore
│   └── side_quests.md          # side quest generate e loro ricongiungimento alla main plot
├── companions/
│   └── {nome_companion}.md     # scheda + arco narrativo, solo se party AI attivo
├── npcs/
│   └── {nome_npc}.md          # un file per NPC, aggiornato ad ogni apparizione
└── items/
    └── inventory_log.md        # log oggetti trovati/persi/usati
```

### Comportamento

- `full_transcript.md` scritto in **append real-time** ad ogni messaggio — nessuna perdita in caso di crash
- `campaign_overview.md` creato **prima che inizi il gioco**, alla generazione campagna
- Gli NPC pre-generati hanno già il loro `.md` con campi "Interazioni" vuoti
- `summary.md` generato dall'AI quando il giocatore clicca "Termina sessione"
- `lore.md` aggiornato ad ogni tag `[LORE:]` nella risposta AI
- Tutti i file UTF-8, Markdown standard, compatibili con Obsidian, Notion, GitHub

### Formato full_transcript.md

```markdown
# Sessione 01 — 2025-06-15

## Il Dungeon Master
*Davanti a voi si erge una torre di pietra nera...*

## Thorin (Guerriero Nano)
Entro dalla porta principale con la spada sguainata.

## Il Dungeon Master
[TIRO: Percezione — 1d20+3 → **17**]
*I vostri occhi si adattano al buio: tre goblin vi fissano...*
```

### Formato summary.md

```markdown
# Riassunto Sessione 01

Il party ha scoperto la Torre Nera di Valdris. Thorin ha guidato l'ingresso,
trovando tre goblin di guardia. Elara ha subito la prima ferita della campagna.
Nelle viscere della torre: una lettera misteriosa firmata "L."

**Cliffhanger:** Chi è L.? E perché conosceva il nome di Thorin?
```

### Campi progressione nella scheda personaggio

Ogni `{nome}.json`/`{nome}.md` del personaggio include:
- `level`: livello attuale
- `xp_current` / `xp_to_next`: presenti solo se sistema = XP classici
- `progression_system`: `xp` | `milestone` (eredita dalla campagna)
- `level_history`: log dei level up con sessione/data e scelte fatte (sottoclasse, ASI, incantesimi)

### Diario del Personaggio

Oltre al lore journal (oggettivo, sul mondo) e ai transcript (dialoghi grezzi), ogni personaggio ha un **diario in prima persona** (`{nome}_diary.md`): pagine scritte dal punto di vista del PG, con pensieri, paure, dubbi morali e riflessioni sulle scelte fatte. Arricchisce la backstory nel tempo, rendendo il personaggio una "persona" che cresce con la campagna.

- **Generazione automatica:** a fine sessione (insieme al `summary.md`), l'AI genera una breve voce di diario (1-2 paragrafi) dal punto di vista del personaggio, basata sugli eventi della sessione e sul tono/personalità del PG
- **Generazione su richiesta:** il giocatore può chiedere in qualsiasi momento "Scrivi una pagina di diario" per una riflessione a metà sessione
- Si collega all'**allineamento tracker**: scelte moralmente significative possono generare voci di diario più introspettive, evidenziando dubbi o conflitti interiori
- Consultabile dal pannello Personaggio (tab dedicata "Diario"), in ordine cronologico
- Esportabile come PDF o incluso nell'export del personaggio (`.dndch`)

Formato:

```markdown
# Diario di Thorin — Sessione 01

Non mi aspettavo che la Torre Nera fosse così... silenziosa. I goblin sono
caduti facilmente, ma quella lettera firmata "L." mi ha lasciato un nodo
allo stomaco. Conoscevano il mio nome. Chi altro, in questa terra, sa chi
sono davvero? Forse è ora che lo scopra, prima che lo scopra qualcun altro.
```

### Formato NPC .md

```markdown
# Bram il Locandiere

**Prima apparizione:** Sessione 01, Taverna del Cervo Ubriaco
**Aspetto:** Uomo robusto, cicatrice sulla guancia, risata facile
**Personalità:** Generoso ma curioso, sa sempre più di quanto dice
**Segreto:** Ex spia della Gilda, ora in pensione

## Interazioni
- **Sessione 01:** Ha offerto una stanza, menzionato "strani rumori dalla torre"
- **Sessione 03:** Ha rivelato di conoscere il simbolo sulla lettera

**Stato relazione:** Amichevole → Alleato
```

---

## Voice Acting (opzionale)

- Integrazione con **Piper TTS** locale per leggere le risposte del DM
- Voce diversa per ogni NPC (pitch e velocità variabili assegnati alla creazione)
- Toggle on/off, controllo volume
- Se Piper non installato: bottone con istruzioni di installazione

---

## Multiplayer Party Reale

- Fino a 6 giocatori connessi alla stessa sessione via **WebSocket**
- Ogni giocatore controlla solo il proprio personaggio
- Stanza identificata da codice 6 caratteri condivisibile
- **Modalità spettatore:** accesso read-only
- **DM override:** utente umano prende il controllo narrativo, AI in modalità assistente

### Sistema a Turni

Il multiplayer funziona rigorosamente a turni sia in combattimento che in esplorazione. Non è possibile inviare messaggi fuori dal proprio turno.

**Fuori dal combattimento — turni narrativi:**
- Ordine turni deciso all'inizio della sessione (o dal DM umano se presente)
- Quando è il tuo turno: input sbloccato, indicatore visivo lampeggiante "È il tuo turno"
- Il giocatore descrive l'azione del suo personaggio in testo libero
- Dopo l'invio, il DM AI risponde e passa automaticamente al turno successivo
- Timer opzionale configurabile per turno (es. 60 secondi), dopo il quale il turno passa con azione "Aspetto e osservo"
- Tutti i giocatori vedono in real-time le azioni degli altri ma non possono intervenire fuori turno
- Un giocatore può "passare il turno" esplicitamente

**In combattimento — turni strutturati:**
- Tiro iniziativa per tutti i personaggi + nemici (gestito dal DM AI)
- Tracker iniziativa visivo sempre visibile a tutti
- Highlight del personaggio di cui è il turno
- Quando tocca a un giocatore: input sbloccato, pannello azioni disponibili (Attacca / Incantesimo / Azione bonus / Movimento / Disimpegna / ecc.)
- Dopo l'azione, il DM AI risolve e passa al prossimo in iniziativa
- Turni nemici gestiti autonomamente dall'AI con narrazione
- Fine round: recap visivo delle azioni nel log

**Sincronizzazione WebSocket — eventi broadcast:**

| Evento | Descrizione |
|---|---|
| `TURN_START {player_id}` | Sblocca input al giocatore, mostra indicatore agli altri |
| `TURN_END {player_id}` | Blocca input, mostra "in attesa..." |
| `DM_RESPONSE {text_chunk}` | Streaming risposta DM visibile a tutti in real-time |
| `DICE_ROLL {player\|formula\|result}` | Tiro dado visibile a tutti |
| `MAP_UPDATE {delta}` | Aggiornamento mappa sincronizzato |
| `INITIATIVE_UPDATE {order}` | Aggiornamento tracker iniziativa |
| `PLAYER_JOINED / PLAYER_LEFT` | Notifiche connessione |
| `TIMER_TICK {seconds_left}` | Countdown turno se timer attivo |

**UI differenziata per stato turno:**
- **Turno attivo:** input evidenziato in oro, bordo luminoso, "È il tuo turno!"
- **In attesa:** input grigio disabilitato, mostra "Turno di {nome}..."
- **Spettatore:** input sempre disabilitato, etichetta "Stai guardando"

---

## Import Personaggio da Roll20

Roll20 non fornisce un'API pubblica di export diretta. L'import avviene tramite due metodi:

### Metodo 1 — Import JSON da Roll20 (consigliato)

Pagina `/character/import` con drag-and-drop area per il file JSON esportato da Roll20.

Mappatura campi:

| Campo Roll20 | Campo interno |
|---|---|
| `character_name` | `name` |
| `race` | `race` |
| `class` + `level` | `class`, `level` |
| `strength/dexterity/...` | `stats.STR/DES/...` |
| `hp`, `hp_max` | `hp.current`, `hp.max` |
| `ac` | `armor_class` |
| `speed` | `speed` |
| `proficiency` | `proficiency_bonus` |
| `saving_throws` | `saving_throws` |
| `skills` | `skills` |
| `attacks_and_spellcasting` | `attacks`, `spells` |
| `equipment` | `inventory` |
| `features_and_traits` | `features` |
| `backstory` | `backstory` |
| `alignment` | `alignment` |
| `personality_traits/ideals/bonds/flaws` | `personality` |

- Campi non mappabili: salvati in "Note aggiuntive"
- Anteprima della scheda convertita con campi mancanti evidenziati
- Spell slots e incantesimi importati nel tracker visivo se presenti

### Metodo 2 — Import manuale guidato

- Campo per incollare testo copiato dalla scheda Roll20
- Bottone "Compila con AI" → l'AI estrae i dati e popola il form automaticamente

### Validazione post-import

- L'AI verifica la coerenza della scheda rispetto al ruleset selezionato
- Avvisi per valori fuori range o combinazioni impossibili
- Scelta: mantieni valori as-is o ricalcola secondo il ruleset

### Export verso Roll20

- Endpoint `GET /api/character/{id}/export/roll20` genera JSON compatibile Roll20
- Bottone "Converti per Roll20" nella pagina import per conversione `.dndch` ↔ Roll20 JSON

---

## Estensioni Native — Export/Import

L'app usa tre estensioni proprietarie. Tutti i formati sono ZIP rinominati contenenti JSON + Markdown, leggibili anche manualmente.

### Estensioni

| Estensione | Contenuto |
|---|---|
| `.dndca` | Campagna: mondo, trama, fazioni, NPC, mappa, lore. **Non include** personaggi né sessioni. |
| `.dndch` | Personaggio singolo: scheda completa, inventario, backstory, progressione. Portabile tra campagne. |
| `.dndx` | Export completo: campagna + personaggi + NPC + sessioni + mappa + lore + bestiario. Backup totale. |

Le sessioni **non** hanno un formato di export dedicato — sono sempre incluse nel `.dndx`.

### Struttura interna

```
── .dndca ──────────────────────────────────
dndca_manifest.json
campaign/
  campaign.json
  campaign_overview.md
  world/
    lore.md
    factions.md
    map.svg
    map_assets/
    map_log.md
    bestiary.md
  npcs/
    {nome}.json
    {nome}.md

── .dndch ──────────────────────────────────
dndch_manifest.json
character/
  {nome}.json
  {nome}.md

── .dndx ───────────────────────────────────
dndx_manifest.json
campaign/
  campaign.json
  campaign_overview.md
  characters/
    {nome}.json
    {nome}.md
  sessions/
    session_01/
      transcript.md
      summary.md
      events.md
      session_state.json
  world/
    lore.md
    factions.md
    map.svg
    map_assets/
    map_log.md
    bestiary.md
  npcs/
    {nome}.json
    {nome}.md
  items/
    inventory_log.md
```

### Manifest (schema comune)

```json
{
  "format": "dndca | dndch | dndx",
  "version": "1.0",
  "app_version": "x.x.x",
  "created_at": "2025-06-15T14:30:00Z",
  "ruleset": "dnd5e",
  "campaign_name": "...",
  "character_name": "...",
  "checksum": "sha256:..."
}
```

### UI Export

Pagina `/campaign/{name}/export`:

```
[ Esporta Campagna   .dndca ]   → mondo, NPC, fazioni, mappa, lore
[ Esporta Personaggio  .dndch ]   → dropdown per selezionare quale PG
[ Esporta Tutto   .dndx ]   → backup completo incluse sessioni
```

Ogni bottone mostra la dimensione stimata prima del download.

**QR Code campagna:** accanto a "Esporta Campagna .dndca" è disponibile un bottone **"Genera QR"** che produce un QR code linkato al file `.dndca` (download diretto o link a `/api/campaign/{name}/export/dndca`). Permette di condividere o "regalare" il mondo/trama/lore di una campagna (senza personaggi né sessioni) scansionando il codice da un altro dispositivo, per poi importarlo come nuova campagna. Il QR è scaricabile come immagine ed eventualmente stampabile insieme alla cover della campagna.

### UI Import

Pagina `/import` con drag-and-drop che accetta tutti e tre i formati. Rilevamento automatico del tipo dal manifest.

**Anteprima prima di importare:**
- `.dndca` → nome campagna, ruleset, numero NPC, fazioni presenti
- `.dndch` → scheda riassuntiva PG con stats principali e classe/livello
- `.dndx` → nome campagna, numero sessioni, numero PG, data ultima sessione

**Opzioni import per `.dndca` e `.dndx`:**
- "Importa come nuova campagna" (default) → crea nuova card nell'Hub
- "Unisci a campagna esistente" → aggiunge a una già presente
- "Sostituisci campagna esistente" → con conferma esplicita

**Opzioni import per `.dndch`:**
- "Aggiungi a campagna esistente" → seleziona quale dal dropdown
- "Importa come personaggio standalone"

**Gestione conflitti:** se NPC o personaggio esiste già, mostra diff e chiede: mantieni locale / usa importato / unisci.

### Versioning

- Backward compatible: versioni precedenti sempre leggibili
- File di versioni future: import tentato ignorando campi sconosciuti, con avviso
- Changelog in `DNDX_FORMAT.md` nella repo

---

## Sessioni e Analytics

- **Session recap automatico** generato dall'AI a fine sessione
- **Dashboard statistiche per campagna:**
  - Sessioni giocate e durata totale
  - Mostri sconfitti con contatore
  - Distribuzione risultati dadi e numero di critici
  - Decisioni chiave estratte dall'AI
  - Grafo del percorso narrativo (mappa delle scelte)
  - Achievement sbloccati
- **Achievement system:** "Primo critico", "Sopravvissuto a TPK", "Diplomatico — risolvi 5 conflitti senza combattere", "Esploratore — visita 50 stanze", ecc.
- Export sessione in PDF narrativo o testo plain
- **Locandina di sessione:** a fine sessione, oltre a `summary.md`, l'AI genera una "locandina" illustrata in stile poster/episodio — titolo evocativo generato dall'AI, 2-3 momenti salienti, illustrazione ASCII/SVG generata in tema con la sessione, e un **QR code** come elemento grafico che, scansionato, apre il `summary.md` completo della sessione. Consultabile nella lista sessioni e esportabile come immagine/PDF.

---

## QR Code — Companion Card e App Companion

L'app genera QR code per estendere l'esperienza fuori dallo schermo principale, riusando una semplice libreria QR lato Go (URL con token, nessuna infrastruttura aggiuntiva).

### Biglietto da visita del personaggio

Ogni personaggio ha un QR code associato (visibile nella scheda personaggio, accanto all'icona/avatar) che, scansionato, apre una **pagina pubblica statica** con: icona/avatar, nome, razza/classe, livello, un estratto di backstory e l'ultimo achievement sbloccato. Utile per:
- Condividere il proprio PG prima di una sessione multiplayer (gli altri giocatori scansionano e vedono chi si uniscono)
- "Carta d'identità" stampabile del personaggio, come ricordo o per il tavolo fisico

Il QR è rigenerato se il personaggio cambia campagna (link aggiornato all'ultima scheda pubblica) e incluso nell'export `.dndch`.

### App companion da tavolo

Pensata per sessioni multiplayer locali attorno a un tavolo:
- Scansionando il QR di un **nemico/boss** durante il combattimento, si apre rapidamente la sua scheda (HP, condizioni, tattiche osservate dal bestiario) su un secondo schermo/telefono — utile per consultare senza disturbare lo schermo principale
- Scansionando il **codice di invito sessione** (lo stesso codice a 6 caratteri del multiplayer) da mobile, si entra istantaneamente nella sessione come giocatore o spettatore
- Le pagine companion sono read-only e leggere (HTML statico generato lato server), pensate per caricare velocemente su mobile in rete locale

---

## System Prompt struttura

```
Sei un Dungeon Master esperto di {RULESET_NAME}.
Regole essenziali del ruleset: {RULESET_RULES_SUMMARY}.
Il party è composto da: {PARTY_DESCRIPTION}.
Compagni AI presenti (se single-player con party AI): {COMPANIONS_DESCRIPTION}.
Livello medio del party: {PARTY_LEVEL}. Range di livello campagna: {LEVEL_CAP_MIN}-{LEVEL_CAP_MAX}. Sistema di progressione: {PROGRESSION_SYSTEM} (xp | milestone).
Mondo e campagna: {CAMPAIGN_OVERVIEW}.
Timeline del mondo (eventi indipendenti dal giocatore, già accaduti e prossimi): {WORLD_TIMELINE}.
Side quest attive e relativi punti di ricongiungimento alla main plot: {SIDE_QUESTS}.
Riassunto storia finora: {SESSION_SUMMARY}.
Ultimi messaggi: {RECENT_MESSAGES}.
NPC presenti e stato relazioni: {NPC_MEMORY}.
Reputazione party con fazioni: {REPUTATION_STATE}.

Regole comportamentali:
- Narra sempre in italiano (o nella lingua del giocatore)
- Rispetta rigorosamente le regole del ruleset selezionato
- Mantieni sempre coerenza con il mondo, la trama e la timeline pre-generati — NON inventare nuovi elementi di main plot al volo: consulta e adatta {CAMPAIGN_OVERVIEW} e {WORLD_TIMELINE}
- Il mondo prosegue anche se il giocatore non interviene: applica i trigger della timeline coerentemente al tempo/sessioni trascorse, anche "fuori scena"
- Le side quest, anche se create al momento, devono ricollegarsi alla main plot tramite il punto di ricongiungimento indicato in {SIDE_QUESTS}
- Se presenti compagni AI, interpretali con personalità e arco narrativo coerenti, facendoli intervenire in narrazione e combattimento
- Mantieni un tono imparziale e stimolante: le conseguenze derivano da scelte e tiri, non dall'umore narrativo; preserva la sfida
- Nei momenti di reale difficoltà del giocatore, offri empatia (indizio, respiro narrativo, via di fuga plausibile) senza annullare la sfida o salvare sistematicamente il giocatore
- Quando serve un tiro: [ROLL: descrizione|formula] (es. [ROLL: Tiro per colpire|1d20+5])
- Suggerisci lanci contestuali con [SUGGEST_ROLL: descrizione|formula] a fine risposta
- Nuova area esplorata: [MAP_UPDATE: tipo_stanza|descrizione_breve] — la descrizione viene usata per generare l'illustrazione/tile della stanza e per posizionarla nella mappa SVG
- Nuovo lore: [LORE: categoria|voce|descrizione]
- Sistema XP classici: dopo scontri/eventi rilevanti assegna esperienza con [XP: quantità|motivo]
- Sistema Milestone: a eventi narrativi chiave segnala level up con [LEVEL_UP: motivo]
- Scala difficoltà e CR dei nemici generati in base a {PARTY_LEVEL}, al range {LEVEL_CAP_MIN}-{LEVEL_CAP_MAX} e alla difficoltà scelta in campagna
- Mai rompere il personaggio (no "come AI...")
- Ogni 20 messaggi, se richiesto, genera riassunto compresso max 500 token
```

### Prompt speciale — Generazione Campagna

Usato solo durante la creazione mondo iniziale, prima di iniziare il gioco. **Questo è il passaggio in cui l'AI deve inventare l'intera storia in anticipo** — main plot completa, timeline del mondo, side quest collegate ed eventuali compagni — proprio per garantire coerenza durante tutta la campagna, senza dover improvvisare elementi strutturali in sessione:

```
Sei un game designer esperto di {RULESET_NAME}.
Il giocatore vuole: {CAMPAIGN_PREFERENCES}.
Il personaggio del giocatore è: {CHARACTER_DESCRIPTION}.
Modalità party: {PARTY_MODE} (con_compagni_ai | solo).

Genera un pacchetto mondo COMPLETO ed END-TO-END in JSON con questi campi.
La main_plot e la world_timeline devono coprire l'intera durata prevista della campagna,
inclusi gli esiti possibili anche se il giocatore non interviene mai sulla trama principale.

- world_name, world_description, tone
- starting_location (nome, descrizione, atmosfera)
- main_plot (trama principale completa: atti, eventi chiave, climax, una o più risoluzioni possibili)
- world_timeline (array di eventi con trigger temporale/di sessione, che accadono indipendentemente
  dalle azioni del giocatore: mosse del villain, evoluzione fazioni, eventi globali)
- side_quests (array di 3-5 side quest, ognuna con: nome, descrizione, npc/fazione/location collegata,
  reconnection_point — come e dove si ricollega alla main_plot)
- npcs (array di 10 NPC con: nome, ruolo, aspetto, personalità, segreto, relazione_iniziale)
- villain (nome, motivazione, piano completo con fasi, connessione_col_personaggio)
- key_locations (array di 3 dungeon/luoghi con: nome, tipo, descrizione, pericoli, tesori)
- factions (array di 3 fazioni con: nome, obiettivo, piano, relazione_col_party)
- starting_hook (testo narrativo personalizzato sul backstory del personaggio, 2-3 paragrafi)
- initial_map (descrizione testuale dell'area iniziale e adiacenti — nome, tipo, connessioni — usata per generare la mappa grafica/SVG iniziale)
- companions (array di 3 compagni se party_mode = con_compagni_ai, altrimenti array vuoto: ognuno con
  nome, classe, ruolo, personalità, relazione_iniziale_col_protagonista, arco_narrativo)

Rispondi SOLO con JSON valido, nessun testo aggiuntivo.
```

---

## Gestione Context Window

1. Mantieni sempre gli ultimi 20 messaggi completi
2. Ogni 20 messaggi: chiedi all'AI riassunto compresso (max 500 token) e sostituisci i messaggi vecchi
3. System prompt + riassunto + ultimi 20 messaggi non devono superare l'80% della context window del modello (configurabile per modello in `config.json`)
4. Il `campaign_overview.md` viene sempre incluso nel contesto in forma compressa
5. `world_timeline.md` e `side_quests.md` vengono sempre inclusi in forma compressa — sono il riferimento per la coerenza narrativa e gli eventi indipendenti dal giocatore
6. Le schede `companions/{nome}.md` (se presenti) sono incluse in forma compressa per mantenere coerenti personalità e archi narrativi dei compagni AI

---

## API Endpoints

```
GET  /                                    → Hub Campagne (schermata principale)
GET  /campaigns                           → Hub Campagne (alias)
GET  /campaign/{name}/manage              → gestione campagna
GET  /campaign/{name}/export              → pagina export
GET  /import                              → pagina import

GET  /character/new                       → creazione personaggio
POST /character/create                    → salva personaggio
GET  /character/{id}                      → scheda personaggio
GET  /character/import                    → import da Roll20 o .dndch

POST /api/campaign/generate               → genera campagna AI (long poll)
GET  /api/campaign/list                   → lista campagne salvate
POST /api/campaign/{name}/archive         → archivia campagna
POST /api/campaign/{name}/complete        → segna come completata

GET  /session/new                         → nuova sessione
GET  /session/{id}                        → riprendi sessione
POST /api/chat                            → messaggio DM (SSE streaming)
POST /api/roll                            → tira dadi
GET  /api/roll/suggest                    → lanci suggeriti contestuali
GET  /api/models                          → lista modelli Ollama
POST /api/model/select                    → cambia modello
GET  /api/sessions                        → lista sessioni campagna corrente
GET  /api/rulesets                        → lista ruleset
POST /api/ruleset/select                  → cambia ruleset

GET  /api/map/{session_id}                → mappa corrente (SVG + asset immagini stanze)
GET  /api/bestiary/{session_id}           → bestiario di campagna
GET  /api/bestiary/{session_id}/{character_id} → bestiario filtrato per personaggio
GET  /api/lore/{session_id}               → lore journal
GET  /api/diary/{character_id}            → diario del personaggio
POST /api/diary/{character_id}/generate   → genera nuova voce di diario su richiesta
GET  /api/npc/{session_id}                → lista NPC e memoria
GET  /api/stats/{session_id}              → statistiche campagna
GET  /api/progression/{character_id}      → XP/livello attuale, progresso al prossimo livello
POST /api/progression/{character_id}/levelup → applica level up (scelte AI-guidate)
POST /api/session/export                  → export PDF/testo

GET  /api/campaign/{name}/export/dndca    → export campagna
GET  /api/character/{id}/export/dndch     → export personaggio
GET  /api/campaign/{name}/export/dndx     → export completo
POST /api/import                          → import qualsiasi formato
POST /api/import/preview                  → anteprima pre-import
POST /api/import/convert/roll20           → converti .dndch ↔ Roll20 JSON
GET  /api/character/{id}/export/roll20    → export Roll20 JSON

GET  /api/tts                             → sintesi vocale Piper
WS   /ws/session/{id}                     → WebSocket multiplayer

GET  /api/character/{id}/card             → pagina pubblica "biglietto da visita" PG
GET  /api/character/{id}/qr               → QR code del biglietto da visita PG (immagine)
GET  /api/campaign/{name}/export/dndca/qr → QR code linkato all'export .dndca della campagna
GET  /api/session/{id}/poster             → locandina di sessione (immagine/HTML con QR verso summary)
GET  /api/companion/monster/{bestiary_id} → scheda companion nemico (HP/condizioni, read-only)
GET  /api/companion/monster/{bestiary_id}/qr → QR code verso la scheda companion del nemico
GET  /api/session/{id}/invite/qr          → QR code del codice invito multiplayer
```

---

## Design UI

**Skill di riferimento per lo sviluppo:** durante l'implementazione frontend, seguire le linee guida della skill **UI UX Pro Max** (repo GitHub `nextlevelbuilder/ui-ux-pro-max-skill`) per scelte di stile, palette colori, pairing dei font, layout, accessibilità (contrasti, focus states, ARIA, touch target) e pattern di animazione/interazione, adattandole al tema dark fantasy descritto sotto.

- **Tema:** dark fantasy — sfondo `#0d0a07`, testo pergamena `#e8d5a3`, oro `#c9a227`, rosso `#8b0000`, verde muschio `#2d4a1e`
- **Font:** `Cinzel` titoli, `IM Fell English` narrazione, `JetBrains Mono` stats/dadi
- **Hub Campagne:** griglia di card con texture pergamena, ogni card ha un'illustrazione ASCII generata dall'AI come "copertina" della campagna
- **Atmosfera:** texture pergamena CSS noise filter, bordi ornamentali SVG, separatori runici
- **Schermata generazione campagna:** progress bar con messaggi narrativi animati
- **Dice Box:** dadi CSS puro con sfaccettature, animazione shake 3D
- **Mappa:** mappa grafica SVG/illustrata componibile a tile per stanza, fog-of-war CSS come overlay, token PG/NPC/compagni come icone sovrapposte e animate
- **Responsive:** mobile con bottom nav bar per i pannelli
- **Animazioni:** typing streaming, shake dado, fog-of-war reveal, fade-in pannelli, movimento token sulla mappa — `prefers-reduced-motion` rispettato

---

## Hotkeys

| Tasto | Azione |
|---|---|
| `r` | Apri dice box |
| `i` | Inventario |
| `m` | Mappa |
| `j` | Lore journal |
| `b` | Bestiario |
| `n` | Lista NPC |
| `c` | Cambio campagna rapido |
| `Enter` | Invia messaggio |
| `Esc` | Chiudi pannelli |
| `?` | Overlay lista hotkeys |

---

## Note implementative

1. **SSE** per streaming chat, **WebSocket** solo per multiplayer
2. Singolo binario con `//go:embed` — zero dipendenze runtime oltre Ollama
3. `config.json`: porta, URL Ollama, modello default, lingua, context window per modello, Piper path
4. Tag AI (`[ROLL:]`, `[MAP_UPDATE:]`, `[LORE:]`, `[SUGGEST_ROLL:]`, `[XP:]`, `[LEVEL_UP:]`) parsati lato server, mai mostrati raw al client
5. La mappa procedurale aggiornata incrementalmente, mai rigenerata da zero
6. Il JSON della campagna generata viene immediatamente convertito in file Markdown e salvato su disco prima di mostrare qualsiasi UI di gioco
7. L'Hub Campagne carica le card leggendo `campaign_overview.md` e `session_state.json` di ogni cartella — nessun database centrale, tutto file-based
8. Il cambio campagna rapido dal dropdown in-game salva lo stato della sessione corrente prima di cambiare
9. Tag `[MAP_UPDATE:]` include anche la posizione corrente dei token PG/compagni quando il party si sposta, così il frontend può animare il movimento dell'icona sulla mappa
10. Per le scelte di stile/UI durante l'implementazione frontend, consultare la skill **UI UX Pro Max** (`nextlevelbuilder/ui-ux-pro-max-skill` su GitHub) e adattarne le linee guida al tema dark fantasy

**Ordine di sviluppo consigliato:**
1. Backend routing + Ollama client + SSE
2. Hub Campagne + gestione campagne multiple
3. Generazione campagna AI + storage file Markdown
4. Creazione personaggio
5. Templates HTML + CSS dark fantasy
6. JS interattivo + dice box
7. Mappa procedurale
8. Sistema turni multiplayer + WebSocket
9. Export/Import `.dndca` / `.dndch` / `.dndx`
10. Roll20 import/export
11. Voice acting Piper (opzionale)

---
Grazie per l'attenzione, inseriro' ogni singolo output di Claude all'interno di un commit e cerchero' di far fare ad ogni commit un punto della roadmap (Ordine di sviluppo consigliato).

-GM
