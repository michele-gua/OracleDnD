/* ═══════════════════════════════════════════════
   Oracle DnD — Campaign Generation
═══════════════════════════════════════════════ */

// ─── URL params ──────────────────────────────────
const params   = new URLSearchParams(location.search);
const slug     = location.pathname.split('/')[2]; // /campaign/{slug}/generate
const skipRedir = params.get('skip');

// ─── DOM ────────────────────────────────────────
const stateReady      = document.getElementById('state-ready');
const stateGenerating = document.getElementById('state-generating');
const stateDone       = document.getElementById('state-done');
const stateError      = document.getElementById('state-error');

const modelSelect   = document.getElementById('model-select');
const btnStart      = document.getElementById('btn-start');
const btnSkip       = document.getElementById('btn-skip');
const btnRetry      = document.getElementById('btn-retry');

const progressBar   = document.getElementById('progress-bar');
const progressMsg   = document.getElementById('progress-message');
const tokenStream   = document.getElementById('token-stream');

// Phase progress map
const PHASES = {
  world:  { label: '⚙ Elaborazione…', pct: 15, final: 65 },
  ascii:  { label: '⚙ Elaborazione…', pct: 70, final: 88 },
  saving: { label: '⚙ Salvataggio…',  pct: 90, final: 98 },
  done:   { label: 'Completato!',      pct: 100, final: 100 },
};

const NARRATIVE_MESSAGES = {
  world:  [
    'Il Dungeon Master plasma la realtà dal nulla…',
    'Le fazioni prendono forma nell\'oscurità…',
    'Gli NPC svegliano la loro coscienza…',
    'I segreti del mondo vengono sigillati nei tomi antichi…',
    'La trama principale emerge dalle nebbie del destino…',
  ],
  ascii:  [
    'Il bardo incide la copertina su una lastra di ossidiana…',
    'L\'illustrazione prende vita carattere per carattere…',
  ],
  saving: [
    'Il monaco scribe trascrive tutto su pergamena…',
    'I file vengono sigillati con cera arcana…',
  ],
};

let currentPhase = null;
let msgInterval  = null;
let msgIdx       = 0;

// ─── Init ────────────────────────────────────────
async function init() {
  // Load campaign info
  try {
    const data = await api('GET', `/api/campaign/${slug}`);
    document.getElementById('campaign-name-title').textContent = data.name || 'Forgiatura del Mondo';
    document.getElementById('campaign-world-sub').textContent  = data.world_name || '';
    btnSkip.href    = `/campaign/${slug}/manage`;
    document.getElementById('btn-manage').href = `/campaign/${slug}/manage`;
    document.getElementById('btn-play-now').addEventListener('click', () => {
      location.href = `/session/new?campaign=${slug}`;
    });
  } catch (e) {
    showError('Campagna non trovata: ' + e.message);
    return;
  }

  // Load models
  try {
    const data = await api('GET', '/api/models');
    const models = data.models || [];
    modelSelect.innerHTML = models.length
      ? models.map(m => `<option value="${esc(m.name)}">${esc(m.name)}</option>`).join('')
      : '<option value="deepseek-r1:8b">deepseek-r1:8b</option>';
  } catch {
    modelSelect.innerHTML = '<option value="deepseek-r1:8b">deepseek-r1:8b</option>';
  }

  btnStart.addEventListener('click', startGeneration);
  btnRetry.addEventListener('click', () => {
    setState('ready');
    startGeneration();
  });
  btnSkip.addEventListener('click', e => {
    e.preventDefault();
    location.href = btnSkip.href;
  });
}

// ─── Generation ──────────────────────────────────
async function startGeneration() {
  setState('generating');
  setProgress(5, 'Aprendo il portale tra i piani…');
  tokenStream.textContent = '';

  const model = modelSelect.value || 'deepseek-r1:8b';

  let finalData = null;

  try {
    const res = await fetch('/api/campaign/generate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ slug, model }),
    });

    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }));
      throw new Error(err.error || res.statusText);
    }

    const reader  = res.body.getReader();
    const decoder = new TextDecoder();
    let   buffer  = '';

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop(); // keep incomplete line

      for (const line of lines) {
        if (!line.startsWith('data: ')) continue;
        const raw = line.slice(6).trim();
        if (!raw) continue;

        let evt;
        try { evt = JSON.parse(raw); } catch { continue; }

        handleEvent(evt);

        if (evt.done) {
          finalData = evt;
          break;
        }
        if (evt.error) throw new Error(evt.error);
      }

      if (finalData) break;
    }

  } catch (e) {
    showError(e.message);
    return;
  }

  if (!finalData || finalData.error) {
    showError(finalData?.error || 'Generazione fallita.');
    return;
  }

  showDone(finalData);
}

function handleEvent(evt) {
  // Phase transitions
  if (evt.phase && evt.phase !== currentPhase && evt.phase !== 'done' && evt.phase !== 'error') {
    setPhase(evt.phase);
  }

  // Narrative message
  if (evt.message) {
    setProgress(null, evt.message);
  }

  // Streamed token
  if (evt.token) {
    tokenStream.textContent += evt.token;
    // Auto-scroll to bottom
    tokenStream.parentElement.scrollTop = tokenStream.parentElement.scrollHeight;
    // Bump progress bar slightly while streaming
    const pct = parseInt(progressBar.style.width) || 0;
    const cfg = PHASES[currentPhase] || {};
    if (pct < (cfg.final || 60)) {
      setProgress(Math.min(pct + 0.3, cfg.final || 60), null);
    }
  }
}

function setPhase(phase) {
  currentPhase = phase;
  const cfg = PHASES[phase] || {};

  // Update phase step UI
  document.querySelectorAll('.phase-step').forEach(el => {
    const p = el.dataset.phase;
    const stateEl = el.querySelector('.phase-state');
    if (p === phase) {
      el.classList.add('active');
      el.classList.remove('done');
      stateEl.textContent = '⚙ In corso…';
      stateEl.className   = 'phase-state running';
    } else if (PHASE_ORDER.indexOf(p) < PHASE_ORDER.indexOf(phase)) {
      el.classList.remove('active');
      el.classList.add('done');
      stateEl.textContent = '✔ Fatto';
      stateEl.className   = 'phase-state done';
    }
  });

  setProgress(cfg.pct || 0, null);
  startNarrativeMessages(phase);
}

const PHASE_ORDER = ['world', 'ascii', 'saving', 'done'];

function startNarrativeMessages(phase) {
  if (msgInterval) clearInterval(msgInterval);
  const messages = NARRATIVE_MESSAGES[phase] || [];
  if (!messages.length) return;
  msgIdx = 0;
  setProgress(null, messages[0]);
  if (messages.length > 1) {
    msgInterval = setInterval(() => {
      msgIdx = (msgIdx + 1) % messages.length;
      setProgress(null, messages[msgIdx]);
    }, 3500);
  }
}

function setProgress(pct, message) {
  if (pct !== null) progressBar.style.width = Math.min(100, pct) + '%';
  if (message)      progressMsg.textContent = message;
}

// ─── States ──────────────────────────────────────
function setState(state) {
  stateReady.hidden      = state !== 'ready';
  stateGenerating.hidden = state !== 'generating';
  stateDone.hidden       = state !== 'done';
  stateError.hidden      = state !== 'error';
}

function showDone(evt) {
  if (msgInterval) clearInterval(msgInterval);
  setProgress(100, 'Completato!');

  // Mark all phases done
  document.querySelectorAll('.phase-step').forEach(el => {
    el.classList.add('done');
    el.classList.remove('active');
    const s = el.querySelector('.phase-state');
    s.textContent = '✔ Fatto';
    s.className   = 'phase-state done';
  });

  setState('done');

  const campaign = evt.campaign || {};
  document.getElementById('done-world-name').textContent =
    campaign.world_name ? `🌍 ${campaign.world_name}` : '';

  // ASCII art
  if (evt.ascii) {
    const wrap = document.getElementById('ascii-cover-wrap');
    document.getElementById('ascii-cover').textContent = evt.ascii;
    wrap.hidden = false;
  }

  // Opening scene
  if (evt.opening) {
    const wrap = document.getElementById('opening-scene-wrap');
    document.getElementById('opening-scene').textContent = evt.opening;
    wrap.hidden = false;
  }

  document.getElementById('btn-manage').href = `/campaign/${slug}/manage`;
}

function showError(msg) {
  if (msgInterval) clearInterval(msgInterval);
  setState('error');
  document.getElementById('error-message').textContent = msg;
}

// ─── API ────────────────────────────────────────
async function api(method, path, body) {
  const opts = { method, headers: { 'Content-Type': 'application/json' } };
  if (body) opts.body = JSON.stringify(body);
  const res = await fetch(path, opts);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || res.statusText);
  }
  return res.json();
}

function esc(s) {
  return String(s || '')
    .replace(/&/g, '&amp;').replace(/</g, '&lt;')
    .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

// ─── Boot ────────────────────────────────────────
setState('ready');
init();
