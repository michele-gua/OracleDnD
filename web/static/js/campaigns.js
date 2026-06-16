/* ═══════════════════════════════════════════════
   Oracle DnD — Campaigns Hub
═══════════════════════════════════════════════ */

// ─── State ──────────────────────────────────────
let allCampaigns = [];
let currentFilter = 'all';
let currentSort   = 'recent';
let deleteTarget  = null;
let modalStep     = 1;

// ─── DOM refs ────────────────────────────────────
const grid        = document.getElementById('campaigns-grid');
const emptyState  = document.getElementById('empty-state');
const loadState   = document.getElementById('loading-state');
const searchInput = document.getElementById('search-input');
const sortSelect  = document.getElementById('sort-select');
const toastCtr    = document.getElementById('toast-container');

// New campaign modal
const modalNew   = document.getElementById('modal-new');
const modalClose = document.getElementById('modal-close');
const modalBack  = document.getElementById('modal-back');
const modalNext  = document.getElementById('modal-next');
const confirmSum = document.getElementById('confirm-summary');

// Delete modal
const modalDel     = document.getElementById('modal-delete');
const deleteName   = document.getElementById('delete-name');
const deleteCancel = document.getElementById('delete-cancel');
const deleteConfirm= document.getElementById('delete-confirm');

// ─── API helpers ─────────────────────────────────
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

// ─── Data ────────────────────────────────────────
async function loadCampaigns() {
  loadState.hidden = false;
  grid.innerHTML   = '';
  emptyState.hidden = true;

  try {
    const data = await api('GET', '/api/campaign/list');
    allCampaigns = data.campaigns || [];
    renderGrid();
  } catch (e) {
    toast('Errore nel caricamento campagne: ' + e.message, 'error');
    allCampaigns = [];
    renderGrid();
  } finally {
    loadState.hidden = true;
  }
}

// ─── Rendering ───────────────────────────────────
function renderGrid() {
  const query    = searchInput.value.trim().toLowerCase();
  let campaigns  = allCampaigns.slice();

  // Filter by status
  if (currentFilter !== 'all') {
    campaigns = campaigns.filter(c => c.status === currentFilter);
  }

  // Filter by search
  if (query) {
    campaigns = campaigns.filter(c =>
      c.name.toLowerCase().includes(query) ||
      (c.world_name || '').toLowerCase().includes(query)
    );
  }

  // Sort
  campaigns.sort((a, b) => {
    switch (currentSort) {
      case 'name':     return a.name.localeCompare(b.name, 'it');
      case 'sessions': return (b.session_count || 0) - (a.session_count || 0);
      case 'duration': return (b.total_minutes || 0) - (a.total_minutes || 0);
      default: {
        const ta = a.last_session_at || a.updated_at || a.created_at;
        const tb = b.last_session_at || b.updated_at || b.created_at;
        return new Date(tb) - new Date(ta);
      }
    }
  });

  grid.innerHTML = '';

  if (campaigns.length === 0) {
    emptyState.hidden = false;
    return;
  }
  emptyState.hidden = true;

  campaigns.forEach(c => {
    const card = buildCard(c);
    grid.appendChild(card);
  });
}

function buildCard(c) {
  const card = document.createElement('article');
  card.className = 'campaign-card';
  card.setAttribute('role', 'listitem');
  card.setAttribute('data-status', c.status || 'active');
  card.setAttribute('tabindex', '0');
  card.setAttribute('aria-label', `Campagna: ${c.name}`);

  const badgeClass = {
    active:    'badge-active',
    completed: 'badge-completed',
    archived:  'badge-archived',
    abandoned: 'badge-abandoned',
  }[c.status] || 'badge-active';

  const badgeLabel = {
    active:    'Attiva',
    completed: 'Completata',
    archived:  'Archiviata',
    abandoned: 'Abbandonata',
  }[c.status] || 'Attiva';

  // Stats
  const sessions = c.session_count || 0;
  const hours    = c.total_minutes ? Math.floor(c.total_minutes / 60) : 0;
  const chars    = (c.characters || []);
  const lastDate = c.last_session_at
    ? formatDate(c.last_session_at)
    : formatDate(c.created_at);

  // Ruleset label
  const rulesetLabel = {
    dnd5e:      'D&D 5e (2014)',
    dnd5e2024:  'D&D 5e (2024)',
    dnd35e:     'D&D 3.5e',
    pathfinder: 'Pathfinder 2e',
  }[c.ruleset] || c.ruleset || 'D&D 5e';

  card.innerHTML = `
    <div class="card-header">
      <h2 class="card-title">${esc(c.name)}</h2>
      <span class="card-status-badge ${badgeClass}">${badgeLabel}</span>
    </div>

    ${c.world_name ? `<div class="card-world">${esc(c.world_name)}</div>` : ''}

    <div class="card-stats">
      <div class="card-stat">
        <span class="stat-value">${sessions}</span>
        <span class="stat-label">Sessioni</span>
      </div>
      ${hours > 0 ? `
      <div class="card-stat">
        <span class="stat-value">${hours}h</span>
        <span class="stat-label">Giocate</span>
      </div>` : ''}
      <div class="card-stat">
        <span class="stat-value">${chars.length}</span>
        <span class="stat-label">Personaggi</span>
      </div>
    </div>

    ${chars.length > 0 ? `
    <div class="card-characters">
      ${chars.slice(0, 3).map(ch => `
        <span class="char-chip">
          ${esc(ch.name)}
          <span class="char-chip-level">Lv${ch.level || 1}</span>
        </span>`).join('')}
      ${chars.length > 3 ? `<span class="char-chip">+${chars.length - 3}</span>` : ''}
    </div>` : ''}

    ${c.last_summary ? `<p class="card-summary">${esc(c.last_summary)}</p>` : ''}

    <div class="card-footer">
      <div>
        <div class="card-ruleset">${rulesetLabel}</div>
        <div class="card-date">Ultima attività: ${lastDate}</div>
      </div>
      <div class="card-actions" role="group" aria-label="Azioni campagna">
        <button class="card-action-btn" data-action="play"    data-slug="${esc(c.slug)}" title="Gioca">▶ Gioca</button>
        <button class="card-action-btn" data-action="manage"  data-slug="${esc(c.slug)}" title="Gestisci">⚙</button>
        ${c.status === 'active' ? `
        <button class="card-action-btn" data-action="archive" data-slug="${esc(c.slug)}" title="Archivia">📁</button>` : ''}
        <button class="card-action-btn danger" data-action="delete" data-slug="${esc(c.slug)}" data-name="${esc(c.name)}" title="Elimina">🗑</button>
      </div>
    </div>
  `;

  // Card click → play (unless clicking a button)
  card.addEventListener('click', e => {
    if (e.target.closest('[data-action]')) return;
    goPlay(c.slug);
  });
  card.addEventListener('keydown', e => {
    if (e.key === 'Enter' && !e.target.closest('[data-action]')) goPlay(c.slug);
  });

  // Action buttons
  card.querySelectorAll('[data-action]').forEach(btn => {
    btn.addEventListener('click', e => {
      e.stopPropagation();
      const { action, slug, name } = btn.dataset;
      handleCardAction(action, slug, name);
    });
  });

  return card;
}

function handleCardAction(action, slug, name) {
  switch (action) {
    case 'play':    goPlay(slug);          break;
    case 'manage':  goManage(slug);        break;
    case 'archive': doArchive(slug);       break;
    case 'delete':  openDeleteModal(slug, name); break;
  }
}

function goPlay(slug)   { window.location.href = `/session/new?campaign=${slug}`; }
function goManage(slug) { window.location.href = `/campaign/${slug}/manage`; }

async function doArchive(slug) {
  try {
    await api('POST', `/api/campaign/${slug}/archive`);
    toast('Campagna archiviata.', 'success');
    await loadCampaigns();
  } catch (e) {
    toast('Errore: ' + e.message, 'error');
  }
}

// ─── Delete modal ────────────────────────────────
function openDeleteModal(slug, name) {
  deleteTarget = slug;
  deleteName.textContent = name;
  modalDel.hidden = false;
}

deleteCancel.addEventListener('click', () => {
  modalDel.hidden = true;
  deleteTarget = null;
});

deleteConfirm.addEventListener('click', async () => {
  if (!deleteTarget) return;
  try {
    await api('DELETE', `/api/campaign/${deleteTarget}`);
    toast('Campagna eliminata.', 'success');
    modalDel.hidden = true;
    deleteTarget = null;
    await loadCampaigns();
  } catch (e) {
    toast('Errore eliminazione: ' + e.message, 'error');
  }
});

// ─── New campaign modal ───────────────────────────
function openNewModal() {
  modalStep = 1;
  syncModalStep();
  modalNew.hidden = false;
  document.getElementById('campaign-name').value = '';
  document.getElementById('world-name').value    = '';
  document.querySelector('input[name="ruleset"][value="dnd5e"]').checked = true;
}

modalClose.addEventListener('click', () => { modalNew.hidden = true; });
modalNew.addEventListener('click', e => { if (e.target === modalNew) modalNew.hidden = true; });
document.getElementById('btn-new-campaign').addEventListener('click', openNewModal);
document.getElementById('btn-new-empty')?.addEventListener('click', openNewModal);

modalBack.addEventListener('click', () => {
  if (modalStep > 1) { modalStep--; syncModalStep(); }
});

modalNext.addEventListener('click', async () => {
  if (modalStep < 3) {
    if (modalStep === 2) {
      const name = document.getElementById('campaign-name').value.trim();
      if (!name) {
        toast('Inserisci un nome per la campagna.', 'error');
        document.getElementById('campaign-name').focus();
        return;
      }
      buildConfirmSummary();
    }
    modalStep++;
    syncModalStep();
  } else {
    await submitNewCampaign();
  }
});

function syncModalStep() {
  [1, 2, 3].forEach(n => {
    document.getElementById(`step-${n}`).hidden = (n !== modalStep);
    const stepEl = document.querySelector(`.step[data-step="${n}"]`);
    stepEl.classList.toggle('active', n === modalStep);
    stepEl.classList.toggle('done',   n < modalStep);
  });

  modalBack.disabled  = (modalStep === 1);
  modalNext.textContent = modalStep === 3 ? '🎲 Crea Campagna' : 'Avanti →';
}

function buildConfirmSummary() {
  const name    = document.getElementById('campaign-name').value.trim();
  const world   = document.getElementById('world-name').value.trim() || '—';
  const ruleset = document.querySelector('input[name="ruleset"]:checked').value;
  const ruleLabel = {
    dnd5e:      'D&D 5e (2014)',
    dnd5e2024:  'D&D 5e (2024)',
    dnd35e:     'D&D 3.5e',
    pathfinder: 'Pathfinder 2e',
  }[ruleset] || ruleset;

  confirmSum.innerHTML = `
    <div class="confirm-row"><span class="confirm-key">Campagna</span><span class="confirm-val">${esc(name)}</span></div>
    <div class="confirm-row"><span class="confirm-key">Mondo</span><span class="confirm-val">${esc(world)}</span></div>
    <div class="confirm-row"><span class="confirm-key">Regolamento</span><span class="confirm-val">${ruleLabel}</span></div>
  `;
}

async function submitNewCampaign() {
  const name    = document.getElementById('campaign-name').value.trim();
  const world   = document.getElementById('world-name').value.trim();
  const ruleset = document.querySelector('input[name="ruleset"]:checked').value;

  modalNext.disabled = true;
  modalNext.textContent = '…';

  try {
    const c = await api('POST', '/api/campaign/create', {
      name, world_name: world, ruleset
    });
    toast(`Campagna "${c.name}" creata!`, 'success');
    modalNew.hidden = true;
    await loadCampaigns();
    // Redirect to generate page so AI builds the world
    setTimeout(() => { window.location.href = `/campaign/${c.slug}/generate`; }, 800);
  } catch (e) {
    toast('Errore creazione: ' + e.message, 'error');
  } finally {
    modalNext.disabled = false;
    modalNext.textContent = '🎲 Crea Campagna';
  }
}

// Keyboard: close modals with Escape
document.addEventListener('keydown', e => {
  if (e.key === 'Escape') {
    if (!modalNew.hidden)  modalNew.hidden  = true;
    if (!modalDel.hidden)  modalDel.hidden  = true;
  }
});

// ─── Toolbar events ───────────────────────────────
searchInput.addEventListener('input', () => renderGrid());

sortSelect.addEventListener('change', () => {
  currentSort = sortSelect.value;
  renderGrid();
});

document.querySelectorAll('.filter-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    currentFilter = btn.dataset.filter;
    renderGrid();
  });
});

// ─── Toast ───────────────────────────────────────
function toast(msg, type = '') {
  const el = document.createElement('div');
  el.className = `toast ${type}`;
  el.textContent = msg;
  toastCtr.appendChild(el);
  setTimeout(() => {
    el.classList.add('fade-out');
    el.addEventListener('animationend', () => el.remove());
  }, 3200);
}

// ─── Utils ───────────────────────────────────────
function esc(s) {
  if (!s) return '';
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function formatDate(iso) {
  if (!iso) return '—';
  try {
    return new Intl.DateTimeFormat('it-IT', { day: '2-digit', month: 'short', year: 'numeric' }).format(new Date(iso));
  } catch { return iso; }
}

// ─── Init ────────────────────────────────────────
loadCampaigns();
