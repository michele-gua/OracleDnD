/* ═══════════════════════════════════════════════
   Oracle DnD — Manage Campaign
═══════════════════════════════════════════════ */

// ─── State ──────────────────────────────────────
const slug = location.pathname.split('/')[2]; // /campaign/{slug}/manage
let campaign = null;

// ─── DOM ────────────────────────────────────────
const pageTitle    = document.getElementById('page-title');
const pageSubtitle = document.getElementById('page-subtitle');
const toastCtr     = document.getElementById('toast-container');

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

// ─── Load ────────────────────────────────────────
async function loadCampaign() {
  try {
    campaign = await api('GET', `/api/campaign/${slug}`);
    renderAll();
  } catch (e) {
    toast('Campagna non trovata: ' + e.message, 'error');
  }
}

function renderAll() {
  // Header
  pageTitle.textContent    = campaign.name;
  pageSubtitle.textContent = campaign.world_name || campaign.ruleset || '';

  // Sidebar stats
  document.getElementById('stat-sessions').textContent = campaign.session_count || 0;
  document.getElementById('stat-hours').textContent    =
    campaign.total_minutes ? Math.floor(campaign.total_minutes / 60) + 'h' : '0h';
  document.getElementById('stat-chars').textContent    = (campaign.characters || []).length;

  const created = campaign.created_at
    ? new Intl.DateTimeFormat('it-IT', { day: '2-digit', month: 'short', year: 'numeric' })
        .format(new Date(campaign.created_at))
    : '—';
  document.getElementById('sidebar-meta').textContent = `Creata il ${created}`;

  // Play button
  document.getElementById('btn-play').onclick = () => {
    location.href = `/session/new?campaign=${slug}`;
  };

  renderOverview();
  renderCharacters();
  renderWorldFiles();
}

// ─── World files tab (markdown file list) ─────────
async function renderWorldFiles() {
  // Show generate button in world tab if no files yet
  const worldTab = document.getElementById('tab-world');
  if (!worldTab) return;

  const worldPlaceholder = worldTab.querySelector('.world-placeholder p');

  // Check if campaign has been generated (has ascii art or last_summary)
  if (campaign.cover_ascii || campaign.last_summary) {
    // Campaign was generated — show links to markdown files
    const worldWrap = worldTab.querySelector('.world-placeholder');
    worldWrap.innerHTML = `
      <div class="world-files">
        <a class="world-file-link" href="/campaign/${slug}/generate" target="_self">
          🔄 Rigenera il mondo con AI
        </a>
        <div class="world-file-list">
          <h3 class="world-files-title">File generati</h3>
          ${[
            { file: 'campaign_overview.md', label: '📖 Panoramica campagna' },
            { file: 'world/starting_area.md', label: '🏰 Area di partenza' },
            { file: 'world/factions.md',      label: '⚔️ Fazioni' },
            { file: 'world/world_timeline.md','label': '🗓 Cronologia' },
            { file: 'world/side_quests.md',   label: '📜 Quest Hooks' },
          ].map(f => `
            <div class="world-file-entry">
              <span>${f.label}</span>
              <span class="file-badge">data/campaigns/${slug}/${f.file}</span>
            </div>
          `).join('')}
        </div>
        ${campaign.cover_ascii ? `
        <div class="ascii-preview-wrap">
          <h3 class="world-files-title">Copertina ASCII</h3>
          <pre class="ascii-preview">${esc(campaign.cover_ascii)}</pre>
        </div>` : ''}
        ${campaign.last_summary ? `
        <div class="premise-wrap">
          <h3 class="world-files-title">Premessa</h3>
          <p class="premise-text">${esc(campaign.last_summary)}</p>
        </div>` : ''}
      </div>
    `;
  } else {
    // Not generated yet — show generate CTA
    const worldWrap = worldTab.querySelector('.world-placeholder');
    worldWrap.innerHTML = `
      <div class="placeholder-icon">🌍</div>
      <p class="placeholder-text">Il mondo non è ancora stato generato.</p>
      <a href="/campaign/${slug}/generate" class="btn-new" style="margin-top:1rem;display:inline-block">
        🎲 Genera il Mondo con AI
      </a>
    `;
  }
}

// ─── Overview tab ────────────────────────────────
function renderOverview() {
  const hero = document.getElementById('overview-hero');
  const body = document.getElementById('overview-body');

  hero.innerHTML = `
    <div>
      <div class="campaign-title-hero">${esc(campaign.name)}</div>
      ${campaign.world_name ? `<div class="campaign-world-hero">🌍 ${esc(campaign.world_name)}</div>` : ''}
    </div>
  `;

  const ruleLabel = {
    dnd5e:      'D&D 5e (2014)',
    dnd5e2024:  'D&D 5e (2024)',
    dnd35e:     'D&D 3.5e',
    pathfinder: 'Pathfinder 2e',
  }[campaign.ruleset] || campaign.ruleset || '—';

  const status = {
    active:    'Attiva',
    completed: 'Completata',
    archived:  'Archiviata',
    abandoned: 'Abbandonata',
  }[campaign.status] || campaign.status;

  body.innerHTML = `
    <div class="overview-meta-row">
      <div class="meta-item"><span class="meta-key">Regolamento</span><span class="meta-val">${ruleLabel}</span></div>
      <div class="meta-item"><span class="meta-key">Stato</span><span class="meta-val">${status}</span></div>
      ${campaign.last_session_at ? `
      <div class="meta-item">
        <span class="meta-key">Ultima sessione</span>
        <span class="meta-val">${formatDate(campaign.last_session_at)}</span>
      </div>` : ''}
    </div>
    ${campaign.notes ? `<div>${campaign.notes.split('\n').map(l => `<p>${esc(l)}</p>`).join('')}</div>` : `
    <p style="color:var(--bone-dim);font-style:italic">
      Nessuna nota. Modifica la campagna per aggiungere una descrizione.
    </p>`}
    ${campaign.last_summary ? `
    <h3>Ultimo riassunto</h3>
    <p>${esc(campaign.last_summary)}</p>` : ''}
  `;

  // Populate edit form
  document.getElementById('edit-name').value  = campaign.name;
  document.getElementById('edit-world').value = campaign.world_name || '';
  document.getElementById('edit-notes').value = campaign.notes || '';
}

// ─── Characters tab ──────────────────────────────
function renderCharacters() {
  const list = document.getElementById('char-list');
  const chars = campaign.characters || [];

  document.getElementById('btn-new-char').href = `/character/new?campaign=${slug}`;

  if (chars.length === 0) {
    list.innerHTML = '<p class="placeholder-text">Nessun personaggio in questa campagna.<br>Creane uno per iniziare l\'avventura.</p>';
    return;
  }

  list.innerHTML = chars.map(ch => `
    <div class="char-card" role="button" tabindex="0" onclick="location.href='/character/${esc(ch.id)}'">
      <div class="char-avatar">⚔️</div>
      <div class="char-info">
        <div class="char-name">${esc(ch.name)}</div>
        <div class="char-class">${esc(ch.race || '')} ${esc(ch.class || '')}</div>
      </div>
      <div class="char-level">Lv ${ch.level || 1}</div>
    </div>
  `).join('');
}

// ─── Tab navigation ──────────────────────────────
document.querySelectorAll('.nav-tab').forEach(btn => {
  btn.addEventListener('click', () => {
    const tab = btn.dataset.tab;
    document.querySelectorAll('.nav-tab').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    document.querySelectorAll('.tab-panel').forEach(p => p.hidden = true);
    document.getElementById(`tab-${tab}`).hidden = false;
  });
});

// ─── Edit overview ────────────────────────────────
document.getElementById('btn-edit-overview').addEventListener('click', () => {
  document.getElementById('overview-view').hidden = true;
  document.getElementById('overview-edit').hidden = false;
});

document.getElementById('btn-cancel-edit').addEventListener('click', () => {
  document.getElementById('overview-view').hidden = false;
  document.getElementById('overview-edit').hidden = true;
});

document.getElementById('btn-save-edit').addEventListener('click', async () => {
  const name  = document.getElementById('edit-name').value.trim();
  const world = document.getElementById('edit-world').value.trim();
  const notes = document.getElementById('edit-notes').value.trim();

  if (!name) { toast('Il nome non può essere vuoto.', 'error'); return; }

  try {
    campaign = await api('PATCH', `/api/campaign/${slug}`, { name, world_name: world, notes });
    document.getElementById('overview-view').hidden = false;
    document.getElementById('overview-edit').hidden = true;
    renderAll();
    toast('Campagna aggiornata.', 'success');
  } catch (e) {
    toast('Errore salvataggio: ' + e.message, 'error');
  }
});

// ─── New session ──────────────────────────────────
document.getElementById('btn-new-session').addEventListener('click', () => {
  location.href = `/session/new?campaign=${slug}`;
});

// ─── Settings: Complete / Archive / Delete ────────
document.getElementById('btn-complete').addEventListener('click', async () => {
  if (!confirm('Segnare questa campagna come completata?')) return;
  try {
    await api('POST', `/api/campaign/${slug}/complete`, {});
    toast('Campagna completata!', 'success');
    campaign.status = 'completed';
    renderAll();
  } catch (e) { toast(e.message, 'error'); }
});

document.getElementById('btn-archive').addEventListener('click', async () => {
  if (!confirm('Archiviare questa campagna?')) return;
  try {
    await api('POST', `/api/campaign/${slug}/archive`);
    toast('Campagna archiviata.', 'success');
    setTimeout(() => location.href = '/campaigns', 1000);
  } catch (e) { toast(e.message, 'error'); }
});

// Delete modal
const modalDel    = document.getElementById('modal-delete');
document.getElementById('btn-delete').addEventListener('click', () => {
  document.getElementById('delete-name').textContent = campaign?.name || slug;
  modalDel.hidden = false;
});
document.getElementById('delete-cancel').addEventListener('click', () => { modalDel.hidden = true; });
document.getElementById('delete-confirm').addEventListener('click', async () => {
  try {
    await api('DELETE', `/api/campaign/${slug}`);
    toast('Campagna eliminata.', 'success');
    setTimeout(() => location.href = '/campaigns', 800);
  } catch (e) { toast(e.message, 'error'); }
});

document.addEventListener('keydown', e => {
  if (e.key === 'Escape') modalDel.hidden = true;
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

function esc(s) {
  return String(s || '')
    .replace(/&/g, '&amp;').replace(/</g, '&lt;')
    .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function formatDate(iso) {
  try { return new Intl.DateTimeFormat('it-IT', { day: '2-digit', month: 'short', year: 'numeric' }).format(new Date(iso)); }
  catch { return iso; }
}

// ─── Init ────────────────────────────────────────
if (!slug) {
  toast('Slug campagna mancante nell\'URL.', 'error');
} else {
  loadCampaign();
}
