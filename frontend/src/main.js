import './fonts.css';
import './style.css';
import { OpenFileDialog, GetFileInfo, AnalyzeFile, GetDBStats, SaveReport,
         GetDatabaseSources, DownloadDatabase, CancelDownload, DeleteDatabase,
         LookupRSID, SaveSession, ListSessions, LoadSession, DeleteSession,
         CheckDatabaseUpdates, CompareFiles } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

// ── STATE ────────────────────────────────────────────────────────────────────
let selectedFilePath = null;
let analysisResult   = null;
let hideRiskFindings = false;
let dbStats          = null;
// Identity key for the current file (filename|sizeMB). Used to scope per-file
// notes so annotations don't bleed between analyses of different people's DNA.
let fileIdent        = null;
// Ancestry for the currently-loaded analysis. Does NOT recalculate risk — we
// don't have per-population effect sizes — but surfaces a caveat banner and
// lets future versions filter studies by study population.
let analysisAncestry = 'any';
// Biological sex for the *currently-loaded* analysis. Not persisted — the
// user is prompted each time because one install may be used to review
// multiple people's files.
let analysisSex      = 'any'; // 'male' | 'female' | 'any'
// The user's selection at upload time — we remember it so the toolbar
// toggle can flip between "filter by my chosen sex" and "show everything".
let chosenSex        = 'any';
// findingGroups[cat] = [{ gid, trait, count, worstStatus, findings: [...] }]
// populated when a category tab is first rendered; each group's detail cards
// are lazy-rendered only on first expand to keep the DOM small.
let findingGroups    = {};

// ── BOOT ─────────────────────────────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', async () => {
  // Restore the user's last theme choice (dark is default).
  try {
    const saved = localStorage.getItem('gr-theme');
    if (saved === 'light') document.documentElement.classList.add('theme-light');
  } catch (_) {}
  renderUploadScreen();
  try {
    dbStats = await GetDBStats();
    updateDBStatDisplay();
  } catch (e) {
    console.error('GetDBStats failed:', e);
  }

  // Listen for database download progress / completion events
  EventsOn('db:progress', (data) => updateDBRow(data));
  EventsOn('db:done',     (data) => onDBDone(data));

  // Single delegated click handler. Inline onclick="" attributes violate
  // our strict CSP (script-src 'self'), so every interactive element uses
  // data-action + data-* params instead.
  document.addEventListener('click', (e) => {
    const el = e.target.closest('[data-action]');
    if (!el) return;
    const action = el.dataset.action;
    switch (action) {
      case 'reload':           location.reload(); break;
      case 'switchTab':        switchTab(el.dataset.tab); break;
      case 'togglePanel':      togglePanel(el.dataset.panel, el); break;
      case 'toggleGroup':      toggleGroup(el.dataset.cat, el.dataset.gid, el); break;
      case 'filter':           setSectionFilter(el.dataset.cat, el.dataset.mode, el); break;
      case 'downloadDB':       downloadDB(el.dataset.source); break;
      case 'cancelDBDownload': cancelDBDownload(el.dataset.source); break;
      case 'deleteDB':         deleteDB(el.dataset.source); break;
      case 'showRisk':
        e.preventDefault();
        hideRiskFindings = false;
        applyRiskVisibility();
        break;
      case 'copyDoctor': {
        const pre = document.getElementById('doctor-pre');
        if (!pre) return;
        navigator.clipboard.writeText(pre.textContent).then(() => {
          el.textContent = '✓ Copied';
          setTimeout(() => { el.textContent = 'Copy to clipboard'; }, 2000);
        });
        break;
      }
      case 'addNote':       promptNote(el.dataset.key, el); break;
      case 'deleteNote':    deleteNote(el.dataset.key); break;
      case 'runLookup':     runRsidLookup(); break;
      case 'openSession':   openSession(el.dataset.id); break;
      case 'deleteSession': confirmDeleteSession(el.dataset.id); break;
      case 'saveCurrentSession': saveCurrentSession(); break;
      case 'checkUpdates':  checkDatabaseUpdates(); break;
      case 'startCompare':  startCompareFlow(); break;
      case 'pickCompareB':  pickCompareFileB(); break;
    }
  });
});

// ── HELPERS ──────────────────────────────────────────────────────────────────
function h(str) {
  if (str == null) return '';
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#x27;');
}

function updateDBStatDisplay() {
  document.querySelectorAll('.stat-snps-val').forEach(el => {
    if (dbStats) el.textContent = dbStats.totalSNPs + '+';
  });
}

// ── UPLOAD SCREEN ─────────────────────────────────────────────────────────────
function renderUploadScreen() {
  document.getElementById('app').innerHTML = uploadScreenHTML();
  document.getElementById('pick-btn').addEventListener('click', pickFile);
  const dz = document.getElementById('drop-zone');
  dz.addEventListener('dragover',  e => { e.preventDefault(); dz.classList.add('drag-over'); });
  dz.addEventListener('dragleave', () => dz.classList.remove('drag-over'));
  dz.addEventListener('drop',      handleDrop);
  // Home-screen tab switching
  document.getElementById('home-tab-analyze').addEventListener('click',    () => switchHomeTab('analyze'));
  document.getElementById('home-tab-sessions').addEventListener('click',   () => switchHomeTab('sessions'));
  document.getElementById('home-tab-compare').addEventListener('click',    () => switchHomeTab('compare'));
  document.getElementById('home-tab-databases').addEventListener('click',  () => switchHomeTab('databases'));
  document.getElementById('home-theme-toggle-btn').addEventListener('click', toggleTheme);
  updateDBStatDisplay();
}

function switchHomeTab(tab) {
  document.querySelectorAll('.home-tab').forEach(t => t.classList.toggle('active', t.dataset.tab === tab));
  document.getElementById('home-pane-analyze').style.display   = tab === 'analyze'   ? '' : 'none';
  document.getElementById('home-pane-sessions').style.display  = tab === 'sessions'  ? '' : 'none';
  document.getElementById('home-pane-compare').style.display   = tab === 'compare'   ? '' : 'none';
  document.getElementById('home-pane-databases').style.display = tab === 'databases' ? '' : 'none';
  if (tab === 'databases') loadDatabasesTab();
  if (tab === 'sessions')  loadSessionsTab();
  if (tab === 'compare')   renderComparePane();
}
window.switchHomeTab = switchHomeTab;

async function pickFile() {
  try {
    const path = await OpenFileDialog();
    if (path) await loadFilePath(path);
  } catch (e) {
    console.error(e);
  }
}

async function handleDrop(e) {
  e.preventDefault();
  e.currentTarget.classList.remove('drag-over');
  const files = e.dataTransfer.files;
  if (files && files.length > 0) {
    const path = files[0].path || files[0].name;
    if (path) await loadFilePath(path);
  }
}

async function loadFilePath(path) {
  selectedFilePath = path;
  try {
    const info = await GetFileInfo(path);
    fileIdent = `${info.name}|${info.sizeMB}`;
    document.getElementById('fr-name').textContent = info.name;
    document.getElementById('fr-meta').textContent = `${info.sizeMB} MB · ready to analyze`;
    const fr = document.getElementById('file-ready');
    fr.classList.add('visible');
    document.getElementById('fr-analyze').addEventListener('click', askSexThenAnalyze);
  } catch (e) {
    console.error('GetFileInfo failed:', e);
  }
}

// ── SEX PROMPT ────────────────────────────────────────────────────────────────
// Each file may represent a different person, so we ask every time rather
// than persisting a default. The choice only affects *filtering* of
// recommendations/findings, never the underlying analysis — the user can
// always switch to "show all" in the report toolbar.
function askSexThenAnalyze() {
  const existing = document.getElementById('sex-modal');
  if (existing) existing.remove();
  const modal = document.createElement('div');
  modal.id = 'sex-modal';
  modal.className = 'welcome-overlay';
  modal.innerHTML = `
    <div class="welcome-modal">
      <div class="welcome-eyebrow">// before we start</div>
      <h2 class="welcome-title">About this file</h2>
      <p class="welcome-desc">
        A few quick questions so findings and recommendations are relevant. You can change
        these later — they only affect filtering, never the underlying analysis.
      </p>
      <div class="pre-privacy">
        🔒 These answers stay on your device alongside your DNA data. They're only used
        locally to filter and match results — never uploaded or shared.
      </div>
      <div class="pre-group">
        <div class="pre-group-label">Gender</div>
        <div class="pre-opts">
          <button class="pre-opt" data-sex="female">♀ Female</button>
          <button class="pre-opt" data-sex="male">♂ Male</button>
          <button class="pre-opt" data-sex="any">⚥ Prefer not to say</button>
        </div>
      </div>
      <div class="pre-group">
        <div class="pre-group-label">Genetic ancestry (best rough match)</div>
        <div class="pre-opts">
          <button class="pre-opt anc-opt" data-anc="eur">European</button>
          <button class="pre-opt anc-opt" data-anc="afr">African</button>
          <button class="pre-opt anc-opt" data-anc="eas">East Asian</button>
          <button class="pre-opt anc-opt" data-anc="sas">South Asian</button>
          <button class="pre-opt anc-opt" data-anc="amr">Admixed American</button>
          <button class="pre-opt anc-opt" data-anc="any">Prefer not to say</button>
        </div>
        <div class="pre-note">
          Most GWAS effect sizes come from European cohorts. For non-European ancestry,
          risk findings should be treated as directional only.
        </div>
      </div>
      <div class="pre-actions">
        <button class="btn-accent" id="pre-continue" disabled>Continue →</button>
      </div>
    </div>`;
  document.body.appendChild(modal);
  let pickedSex = null, pickedAnc = null;
  const continueBtn = modal.querySelector('#pre-continue');
  const refresh = () => { continueBtn.disabled = !(pickedSex && pickedAnc); };
  modal.querySelectorAll('.pre-opt[data-sex]').forEach(btn => {
    btn.addEventListener('click', () => {
      modal.querySelectorAll('.pre-opt[data-sex]').forEach(b => b.classList.remove('selected'));
      btn.classList.add('selected');
      pickedSex = btn.dataset.sex;
      refresh();
    });
  });
  modal.querySelectorAll('.pre-opt[data-anc]').forEach(btn => {
    btn.addEventListener('click', () => {
      modal.querySelectorAll('.pre-opt[data-anc]').forEach(b => b.classList.remove('selected'));
      btn.classList.add('selected');
      pickedAnc = btn.dataset.anc;
      refresh();
    });
  });
  continueBtn.addEventListener('click', () => {
    analysisSex      = pickedSex;
    chosenSex        = pickedSex;
    analysisAncestry = pickedAnc;
    modal.remove();
    startAnalysis();
  });
}

// ── PROGRESS SCREEN ───────────────────────────────────────────────────────────
async function startAnalysis() {
  if (!selectedFilePath) return;
  document.getElementById('app').innerHTML = progressScreenHTML();

  const fill  = document.getElementById('prog-fill');
  const label = document.getElementById('prog-label');
  const text  = document.getElementById('prog-text');
  const count = document.getElementById('prog-count');

  EventsOn('analysis:progress', (data) => {
    fill.style.width  = data.pct + '%';
    label.textContent = data.label;
    text.textContent  = data.text;
    count.textContent = data.pct + '%';
  });

  try {
    analysisResult = await AnalyzeFile(selectedFilePath);
    setTimeout(() => renderWelcome(), 500);
  } catch (e) {
    document.getElementById('app').innerHTML = `
      <div style="display:flex;align-items:center;justify-content:center;height:100vh;flex-direction:column;gap:16px">
        <div style="color:var(--red);font-size:18px">Analysis failed</div>
        <div style="color:var(--ink2);font-size:13px;max-width:400px;text-align:center">${h(String(e))}</div>
        <button class="btn-primary" data-action="reload">Try Again</button>
      </div>`;
  }
}

// ── WELCOME MODAL ─────────────────────────────────────────────────────────────
function renderWelcome() {
  document.getElementById('app').innerHTML = welcomeModalHTML();
  document.getElementById('welcome-full').addEventListener('click', () => {
    hideRiskFindings = false;
    renderReport('overview');
  });
  document.getElementById('welcome-action').addEventListener('click', () => {
    hideRiskFindings = false;
    renderReport('action-plan');
  });
  document.getElementById('welcome-hide-risk').addEventListener('click', () => {
    hideRiskFindings = true;
    renderReport('overview');
  });
}

// ── REPORT ────────────────────────────────────────────────────────────────────
const CAT_META = {
  nutrition:       { title: 'Nutrition & Eating',        icon: '🥗', order: 1 },
  supplement:      { title: 'Supplement Protocol',       icon: '💊', order: 2 },
  fitness:         { title: 'Exercise & Training',       icon: '🏋️', order: 3 },
  sleep:           { title: 'Sleep & Recovery',          icon: '🌙', order: 4 },
  cardiovascular:  { title: 'Cardiovascular Health',     icon: '🫀', order: 5 },
  disease_risk:    { title: 'Disease Risk',              icon: '🔬', order: 6 },
  hormones:        { title: 'Hormones & Metabolism',     icon: '⚗️', order: 7 },
  mental_health:   { title: 'Mental Health & Cognition', icon: '🧠', order: 8 },
  longevity:       { title: 'Longevity & Aging',         icon: '⏳', order: 9 },
  immune:          { title: 'Immune Function',           icon: '🛡️', order: 10 },
  neurological:    { title: 'Neurological',              icon: '⚡', order: 11 },
  gut:             { title: 'Gut Health',                icon: '🦠', order: 12 },
  bone:            { title: 'Bone & Joint',              icon: '🦴', order: 13 },
  cancer_risk:     { title: 'Cancer Risk',               icon: '🎗️', order: 14 },
};

function renderReport(startTab = 'overview') {
  if (!analysisResult) return;
  const { parsed, matched, categories, summary, actionPlan } = analysisResult;

  const cats = Object.keys(categories)
    .filter(k => categories[k] && categories[k].length > 0)
    .sort((a, b) => (CAT_META[a]?.order || 99) - (CAT_META[b]?.order || 99));

  document.getElementById('app').innerHTML = reportShellHTML(parsed, matched, cats, summary);

  renderOverviewTab(summary, matched, parsed);
  renderActionPlanTab(actionPlan);
  cats.forEach((cat, idx) => renderCategoryTab(cat, categories[cat], idx));
  if (summary.drugs && summary.drugs.length > 0) renderDrugsTab(summary.drugs);
  renderLookupTab();
  renderDoctorTab();

  document.getElementById('risk-toggle-btn').addEventListener('click', toggleRisk);
  document.getElementById('sex-filter-btn').addEventListener('click', toggleSexFilter);
  updateSexFilterBtn();
  document.getElementById('theme-toggle-btn').addEventListener('click', toggleTheme);
  document.getElementById('print-btn').addEventListener('click', printReport);
  document.getElementById('save-btn').addEventListener('click', saveReport);
  document.getElementById('new-analysis-btn').addEventListener('click', () => {
    selectedFilePath = null;
    analysisResult   = null;
    renderUploadScreen();
  });

  // Allow vertical mouse-wheel to scroll the tab bar horizontally when it
  // overflows the viewport. Shift+wheel already scrolls X on most systems,
  // but a plain wheel is more discoverable.
  const tabBar = document.querySelector('.tab-bar');
  if (tabBar) {
    tabBar.addEventListener('wheel', (e) => {
      if (e.deltaY === 0) return;
      if (tabBar.scrollWidth <= tabBar.clientWidth) return;
      e.preventDefault();
      tabBar.scrollLeft += e.deltaY;
    }, { passive: false });
  }

  switchTab(startTab);
  applyRiskVisibility();
}

function switchTab(name) {
  document.querySelectorAll('.tab-item').forEach(t => t.classList.toggle('active', t.dataset.tab === name));
  document.querySelectorAll('.rep-section').forEach(s => s.classList.toggle('active', s.dataset.tab === name));
}
window.switchTab = switchTab;

function toggleRisk() {
  hideRiskFindings = !hideRiskFindings;
  document.getElementById('risk-toggle-btn').classList.toggle('active', hideRiskFindings);
  applyRiskVisibility();
}

function applyRiskVisibility() {
  const riskSection = document.getElementById('tab-disease_risk');
  const notice      = document.getElementById('risk-hidden-notice');
  if (riskSection) riskSection.style.display = hideRiskFindings ? 'none' : '';
  if (notice)      notice.classList.toggle('visible', hideRiskFindings);
  const tabItem = document.querySelector('.tab-item[data-tab="disease_risk"]');
  if (tabItem) tabItem.style.display = hideRiskFindings ? 'none' : '';
}
window.applyRiskVisibility = applyRiskVisibility;

function toggleSexFilter() {
  // If user picked "Prefer not to say", the button is disabled (no filter to toggle).
  if (chosenSex === 'any') return;
  analysisSex = (analysisSex === 'any') ? chosenSex : 'any';
  updateSexFilterBtn();
  // Re-render the whole report so filters re-apply everywhere.
  renderReport(document.querySelector('.tab-item.active')?.dataset.tab || 'overview');
}

function updateSexFilterBtn() {
  const btn = document.getElementById('sex-filter-btn');
  if (!btn) return;
  if (chosenSex === 'any') {
    btn.textContent = '⚥ All';
    btn.disabled = true;
    btn.style.opacity = '0.5';
    btn.style.cursor = 'default';
    return;
  }
  const icon = chosenSex === 'female' ? '♀' : '♂';
  const label = chosenSex === 'female' ? 'Female' : 'Male';
  btn.textContent = analysisSex === 'any' ? `⚥ Showing all` : `${icon} ${label}`;
  btn.classList.toggle('active', analysisSex !== 'any');
}

function toggleTheme() {
  const root = document.documentElement;
  const nowLight = root.classList.toggle('theme-light');
  try { localStorage.setItem('gr-theme', nowLight ? 'light' : 'dark'); } catch (_) {}
}

// Print / export-as-PDF. The dedicated print stylesheet hides the report
// shell, tabs, and per-variant cards, and shows only the action plan plus
// each category summary — a clean, shareable one-document handout.
function printReport() {
  // Force all action-plan panels open so none are clipped in the printout.
  document.querySelectorAll('.ap-body').forEach(b => b.classList.add('open'));
  document.querySelectorAll('.ap-header').forEach(h => h.classList.add('open'));
  window.print();
}

async function saveReport() {
  if (!analysisResult) return;
  try {
    const path = await SaveReport(analysisResult.doctorText);
    if (path) {
      const btn = document.getElementById('save-btn');
      const orig = btn.textContent;
      btn.textContent = '✓ Saved';
      setTimeout(() => { btn.textContent = orig; }, 2000);
    }
  } catch (e) {
    console.error('Save failed:', e);
  }
}

// ── TAB RENDERERS ─────────────────────────────────────────────────────────────
function renderOverviewTab(summary, matched, parsed) {
  const el = document.getElementById('tab-overview');
  if (!el) return;
  const dbCount = dbStats ? dbStats.totalSNPs + '+' : '357+';

  el.innerHTML = `
    ${ancestryCaveatHTML()}
    <div class="ap-cta-banner" data-action="switchTab" data-tab="action-plan">
      <span class="ap-cta-icon">⚡</span>
      <div class="ap-cta-text">
        <div class="ap-cta-title">Want your personalized Action Plan?</div>
        <div class="ap-cta-sub">Diet, supplement, exercise &amp; sleep recommendations tailored to your variants — click to jump there.</div>
      </div>
      <span class="ap-cta-arrow">→</span>
    </div>
    <div class="risk-hidden-notice" id="risk-hidden-notice">
      <span>🔒</span>
      <span>Disease risk findings are hidden. <a href="#" data-action="showRisk">Show them</a> or use the Risk toggle in the toolbar.</span>
    </div>
    <div class="sec-hdr">
      <div class="sec-num">00</div><div class="sec-title">Analysis Overview</div><div class="sec-icon">📊</div>
    </div>
    <div class="overview-stats">
      <div class="ov-stat"><span class="ov-stat-val">${h(parsed.provider)}</span><span class="ov-stat-label">Provider</span></div>
      <div class="ov-stat"><span class="ov-stat-val">${(parsed.totalSNPs||0).toLocaleString()}</span><span class="ov-stat-label">SNPs in File</span></div>
      <div class="ov-stat"><span class="ov-stat-val">${matched}</span><span class="ov-stat-label">Variants Matched</span></div>
      <div class="ov-stat"><span class="ov-stat-val">${dbCount}</span><span class="ov-stat-label">SNPs in Database</span></div>
      <div class="ov-stat"><span class="ov-stat-val">${h(analysisResult.generatedAt)}</span><span class="ov-stat-label">Generated</span></div>
    </div>
    ${ovGroup('⬤ High Risk — Requires Attention', 'c-red',   (summary.high      || []).filter(findingPassesSex), 10)}
    ${ovGroup('⬤ Moderate — Worth Monitoring',    'c-amber', (summary.moderate  || []).filter(findingPassesSex), 10)}
    ${ovGroup('⬤ Protective — Working In Your Favor','c-green',(summary.protective|| []).filter(findingPassesSex), 0)}
    ${!summary.high?.length && !summary.moderate?.length && !summary.protective?.length ? `
    <div class="empty-state">
      <div class="empty-state-icon">🔬</div>
      <p>No matched variants found. Ensure the file is a valid raw DNA export (23andMe, AncestryDNA, etc.).</p>
    </div>` : ''}
  `;
}

function ovGroup(label, cls, findings, limit) {
  if (!findings || findings.length === 0) return '';
  const show = limit > 0 ? findings.slice(0, limit) : findings;
  return `
    <div class="ov-group">
      <div class="ov-group-label ${cls}">${label} (${findings.length})</div>
      ${show.map(f => `
      <div class="ov-row" data-action="switchTab" data-tab="${f.cat}">
        <div class="ov-dot ${cls}"></div>
        <div class="ov-trait">${h(f.trait)}</div>
        <div class="ov-gene">${h(f.gene)}</div>
        <div class="ov-cat">${CAT_META[f.cat]?.title || f.cat} →</div>
      </div>`).join('')}
      ${limit > 0 && findings.length > limit ? `<div class="ov-more">+${findings.length - limit} more — click category tab to see all</div>` : ''}
    </div>`;
}

function renderActionPlanTab(ap) {
  const el = document.getElementById('tab-action-plan');
  if (!el || !ap) return;

  const filterItems = (items) => (items || []).filter(it => sexMatches(it.text));
  const panels = [
    { id: 'ap-diet',     icon: '🥗', title: 'Diet & Nutrition',   items: filterItems(ap.diet) },
    { id: 'ap-supps',    icon: '💊', title: 'Supplement Protocol', items: filterItems(ap.supplements) },
    { id: 'ap-exercise', icon: '🏋️', title: 'Exercise & Training', items: filterItems(ap.exercise) },
    { id: 'ap-sleep',    icon: '🌙', title: 'Sleep Optimization',  items: filterItems(ap.sleep) },
    { id: 'ap-monitor',  icon: '🩺', title: 'Health Monitoring',   items: filterItems(ap.monitoring) },
  ];

  el.innerHTML = `
    <div class="sec-hdr">
      <div class="sec-num">⚡</div><div class="sec-title">Your Action Plan</div>
      <div class="sec-icon">📋</div><div class="sec-badge">Personalized to your variants</div>
    </div>
    <p style="font-size:13px;color:var(--ink2);line-height:1.7;margin-bottom:24px">
      These recommendations are derived directly from your actionable genetic findings.
      Each bullet is linked to the specific variant driving it.
    </p>
    ${panels.map((p, i) => `
    <div class="action-plan">
      <div class="ap-header ${i === 0 ? 'open' : ''}" data-action="togglePanel" data-panel="${p.id}">
        <span class="ap-icon">${p.icon}</span>
        <div>
          <div class="ap-title">${p.title}</div>
          <div class="ap-subtitle">${(p.items||[]).length} personalized recommendations</div>
        </div>
        <span class="ap-chevron">▼</span>
      </div>
      <div class="ap-body ${i === 0 ? 'open' : ''}" id="${p.id}">
        <ul class="ap-bullets">
          ${(p.items||[]).map(item => `
          <li>
            <span class="ap-bullet-dot">→</span>
            <span style="flex:1">${h(item.text)}</span>
            ${item.gene && item.gene !== 'General' ? `<span class="ap-gene-tag">${h(item.gene)}</span>` : ''}
          </li>`).join('')}
        </ul>
      </div>
    </div>`).join('')}
  `;
}

window.togglePanel = function(id, hdr) {
  const body = document.getElementById(id);
  if (!body) return;
  const open = body.classList.toggle('open');
  hdr.classList.toggle('open', open);
};

function renderCatSummary(meta, sorted, actionable) {
  const high       = sorted.filter(f => f.status === 'homozygous_risk');
  const moderate   = sorted.filter(f => f.status === 'heterozygous');
  const protective = sorted.filter(f => f.status === 'protective');

  // Dedupe action bullets by (rec text + gene) so repeat recommendations
  // across multiple variants collapse. Keep most-severe status' copy.
  const seen = new Map();
  for (const f of actionable) {
    const rec = (f.rec || '').trim();
    if (!rec) continue;
    const key = rec + '|' + (f.gene || '');
    if (!seen.has(key)) {
      seen.set(key, { text: rec, gene: f.gene, trait: f.trait, rsid: f.rsid, status: f.status });
    }
  }
  const bullets = [...seen.values()];

  if (high.length + moderate.length + protective.length === 0 && bullets.length === 0) {
    return '';
  }

  const statPill = (n, cls, label) => n > 0
    ? `<span class="cat-sum-pill ${cls}"><span class="cat-sum-pill-n">${n}</span>${label}</span>`
    : '';

  return `
    <div class="cat-summary">
      <div class="cat-summary-head">
        <span class="cat-summary-eyebrow">// section summary</span>
        <div class="cat-summary-title">What this section means for you</div>
      </div>
      <div class="cat-summary-stats">
        ${statPill(high.length,       'cat-sum-red',   'high risk')}
        ${statPill(moderate.length,   'cat-sum-amber', 'moderate')}
        ${statPill(protective.length, 'cat-sum-green', 'protective')}
      </div>
      ${bullets.length > 0 ? `
      <div class="cat-summary-actions">
        <div class="cat-summary-sub">Actionable items from variants in this section:</div>
        <ul class="cat-summary-list">
          ${bullets.map(b => `
          <li>
            <span class="cat-sum-dot cat-sum-dot-${b.status === 'homozygous_risk' ? 'red' : b.status === 'heterozygous' ? 'amber' : 'green'}"></span>
            <span class="cat-sum-text">${h(b.text)}</span>
            ${b.gene ? `<span class="cat-sum-tag">${h(b.gene)}</span>` : ''}
          </li>`).join('')}
        </ul>
      </div>` : ''}
    </div>`;
}

const STATUS_ORDER = { homozygous_risk: 0, heterozygous: 1, protective: 2, normal: 3 };

// Keyword patterns for sex-specific content. Keyword matching isn't perfect
// but it cleanly handles the vast majority of clinical genetic findings
// (breast/ovarian/cervical/prostate/testicular/PSA/mammogram/etc.) without
// needing a per-record sex annotation.
const FEMALE_ONLY = /\b(breast|mammogra|ovari|cervic|endometri|menstr|menopaus|pregnan|uter(?:us|ine)|vulv|vagin|pcos)\b/i;
const MALE_ONLY   = /\b(prostate|testic|psa\b|erectile|bph|benign prostatic)\b/i;

// True if this text is relevant given the currently-chosen sex.
// 'any' passes everything through.
function sexMatches(text) {
  if (!text || analysisSex === 'any') return true;
  const s = String(text);
  if (analysisSex === 'male'   && FEMALE_ONLY.test(s)) return false;
  if (analysisSex === 'female' && MALE_ONLY.test(s))   return false;
  return true;
}

// A finding clears the sex filter if none of its text fields are flagged as
// belonging to the opposite sex. We check trait, desc, and rec so we catch
// cases where the trait is neutral ("cancer risk") but the recommendation
// mentions "annual mammogram".
function findingPassesSex(f) {
  return sexMatches(f.trait) && sexMatches(f.desc) && sexMatches(f.rec) && sexMatches(f.effect);
}

// Normalize GWAS/ClinVar trait strings so cosmetic variants collapse together.
// GWAS in particular emits verbose labels like:
//   "Calcium levels"
//   "Calcium levels (UKB data field 30680)"
//   "Calcium (maximum, inv-norm transformed)"
//   "25-hydroxyvitamin D levels (skin colour stratified)"
// These all mean the same thing to an end user, so we strip:
//   1. Anything in parentheses / brackets (statistical transforms, dataset tags)
//   2. Trailing " levels" / " level"     (biomarker-measurement noise)
//   3. Common statistical qualifiers     (log, ln, rank-inverse, residualized…)
// then lowercase + collapse whitespace.
function normalizeTrait(t) {
  if (!t) return '(unspecified)';
  let s = String(t)
    .replace(/\s*\([^)]*\)/g, '')   // strip parenthetical suffixes
    .replace(/\s*\[[^\]]*\]/g, '')  // strip bracketed suffixes
    .replace(/\s+/g, ' ')
    .trim()
    .toLowerCase();
  s = s.replace(/\b(log[- ]?transformed|ln[- ]?transformed|inv[- ]?norm(?:al)? transformed|rank[- ]?inverse[- ]?normal|residualized|adjusted|stratified|age[- ]?adjusted|bmi[- ]?adjusted|sex[- ]?adjusted)\b/g, '');
  s = s.replace(/\s+(levels?|concentration|measurement|phenotype)$/g, '');
  s = s.replace(/\s+/g, ' ').trim();
  return s || '(unspecified)';
}

function groupFindings(findings) {
  const map = new Map();
  for (const f of findings) {
    const key = normalizeTrait(f.trait);
    if (!map.has(key)) map.set(key, []);
    map.get(key).push(f);
  }
  const groups = [];
  let gid = 0;
  for (const [key, items] of map) {
    items.sort((a, b) => (STATUS_ORDER[a.status] ?? 3) - (STATUS_ORDER[b.status] ?? 3));
    // Prefer the shortest raw trait string as the display label — it's usually
    // the least-noisy variant (e.g. "Calcium levels" beats
    // "Calcium (mean, inv-norm transformed)").
    const label = items.map(i => i.trait).filter(Boolean)
      .sort((a, b) => a.length - b.length)[0] || key;
    groups.push({
      gid: gid++,
      trait: label,
      count: items.length,
      worstStatus: items[0].status,
      allNormal: items.every(f => f.status === 'normal'),
      findings: items,
    });
  }
  groups.sort((a, b) => (STATUS_ORDER[a.worstStatus] ?? 3) - (STATUS_ORDER[b.worstStatus] ?? 3));
  return groups;
}

function groupRow(cat, g) {
  const badgeClass = g.worstStatus === 'homozygous_risk' ? 'red'
                   : g.worstStatus === 'heterozygous'    ? 'amber'
                   : g.worstStatus === 'protective'      ? 'green'
                   :                                       'muted';
  const badgeText = g.worstStatus === 'homozygous_risk' ? 'High Risk'
                  : g.worstStatus === 'heterozygous'    ? 'Moderate'
                  : g.worstStatus === 'protective'      ? 'Protective'
                  :                                       'Normal';
  const hidden = g.allNormal ? ' fg-all-normal" style="display:none' : '';
  return `
  <div class="finding-group${hidden}" data-cat="${h(cat)}" data-gid="${g.gid}">
    <div class="fg-header" data-action="toggleGroup" data-cat="${h(cat)}" data-gid="${g.gid}">
      <span class="fg-chevron">▶</span>
      <div class="fg-trait">${h(g.trait)}</div>
      <span class="fg-count">${g.count} variant${g.count > 1 ? 's' : ''}</span>
      <span class="fg-badge fg-badge-${badgeClass}">${badgeText}</span>
    </div>
    <div class="fg-body" id="fg-body-${h(cat)}-${g.gid}"></div>
  </div>`;
}

function toggleGroup(cat, gid, hdr) {
  const wrap = hdr.parentElement;
  const body = document.getElementById(`fg-body-${cat}-${gid}`);
  if (!wrap || !body) return;
  const open = wrap.classList.toggle('open');
  if (open && body.dataset.rendered !== '1') {
    const g = (findingGroups[cat] || [])[Number(gid)];
    if (g) {
      body.innerHTML = g.findings.map(f => findingCard(f, false)).join('');
      body.dataset.rendered = '1';
    }
  }
}

function renderCategoryTab(cat, findings, idx) {
  const el = document.getElementById('tab-' + cat);
  if (!el) return;
  const meta   = CAT_META[cat] || { title: cat, icon: '🔬' };
  const filtered = findings.filter(findingPassesSex);
  const sorted = [...filtered].sort((a, b) => (STATUS_ORDER[a.status] ?? 3) - (STATUS_ORDER[b.status] ?? 3));
  const actionable = sorted.filter(f => f.status !== 'normal');
  const normal     = sorted.filter(f => f.status === 'normal');

  const groups = groupFindings(sorted);
  findingGroups[cat] = groups;
  const normalGroups = groups.filter(g => g.allNormal).length;

  el.innerHTML = `
    <div class="sec-hdr">
      <div class="sec-num">0${idx + 1}</div>
      <div class="sec-title">${meta.title}</div>
      <div class="sec-icon">${meta.icon}</div>
      <div class="sec-badge">${groups.length} trait${groups.length === 1 ? '' : 's'} · ${filtered.length} variant${filtered.length === 1 ? '' : 's'}</div>
    </div>
    ${renderCatSummary(meta, sorted, actionable)}
    ${normalGroups > 0 ? `
    <div class="filter-bar">
      <span class="filter-bar-label">Show</span>
      <button class="filter-btn active" data-action="filter" data-cat="${cat}" data-mode="actionable">Actionable only</button>
      <button class="filter-btn" data-action="filter" data-cat="${cat}" data-mode="all">All findings</button>
      <span class="filter-count" id="filter-count-${cat}">${groups.length - normalGroups} actionable · ${normalGroups} normal hidden</span>
    </div>` : ''}
    <div id="findings-${cat}">
      ${groups.map(g => groupRow(cat, g)).join('')}
    </div>`;
}

function setSectionFilter(cat, mode, btn) {
  document.querySelectorAll(`#findings-${cat} .fg-all-normal`).forEach(el => {
    el.style.display = mode === 'all' ? '' : 'none';
  });
  btn.closest('.filter-bar').querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
  btn.classList.add('active');
  const countEl = document.getElementById('filter-count-' + cat);
  if (countEl) {
    const groups = findingGroups[cat] || [];
    const actionCount = groups.filter(g => !g.allNormal).length;
    const normCount   = groups.filter(g =>  g.allNormal).length;
    countEl.textContent = mode === 'all' ? 'showing all findings' : `${actionCount} actionable · ${normCount} normal hidden`;
  }
}
window.setSectionFilter = setSectionFilter;

function renderDrugsTab(drugs) {
  const el = document.getElementById('tab-drugs');
  if (!el) return;
  drugs = drugs.filter(findingPassesSex);
  el.innerHTML = `
    <div class="sec-hdr">
      <div class="sec-num">💉</div>
      <div class="sec-title">Drug &amp; Medication Response</div>
      <div class="sec-badge">${drugs.length} interactions</div>
    </div>
    <p style="font-size:13px;color:var(--ink2);line-height:1.75;margin-bottom:24px">
      Pharmacogenomic variants affect how your body processes medications.
      Share this section with your prescribing physicians and pharmacist before starting new medications.
    </p>
    <table class="drug-tbl">
      <thead><tr><th>Drug / Category</th><th>Gene</th><th>Your Effect</th><th>Action Required</th><th>Level</th></tr></thead>
      <tbody>
        ${drugs.map(d => `<tr>
          <td><strong style="color:var(--ink)">${h(d.subcat)||'Multiple'}</strong></td>
          <td><code style="font-family:'Fira Code',monospace;font-size:10px;color:var(--accent)">${h(d.gene)}</code>
              <span style="font-family:'Fira Code',monospace;font-size:9px;color:var(--muted)">${h(d.rsid)}</span></td>
          <td style="max-width:220px">${h(d.effect)}</td>
          <td style="max-width:240px;color:var(--amber)">${h(d.rec)}</td>
          <td><span class="${d.color==='red'?'d-sev-high':'d-sev-mod'}">${h(d.badge)}</span></td>
        </tr>`).join('')}
      </tbody>
    </table>`;
}

function renderDoctorTab() {
  const el = document.getElementById('tab-doctor');
  if (!el || !analysisResult) return;
  el.innerHTML = `
    <div class="sec-hdr">
      <div class="sec-num">🩺</div>
      <div class="sec-title">Doctor-Friendly Summary</div>
      <div class="sec-badge">Share with your provider</div>
    </div>
    <p style="font-size:13px;color:var(--ink2);line-height:1.7;margin-bottom:16px">
      Copy or save the text below and share with your healthcare provider or genetic counselor.
    </p>
    <button class="btn-secondary" style="margin-bottom:16px" data-action="copyDoctor">
      Copy to clipboard
    </button>
    <pre id="doctor-pre" class="doc-pre">${h(analysisResult.doctorText)}</pre>`;
}

function findingCard(f, hidden = false) {
  const pmidLink = f.pmid
    ? `<a href="https://pubmed.ncbi.nlm.nih.gov/${h(f.pmid)}" target="_blank" class="pubmed-link">PubMed ${h(f.pmid)}</a>`
    : '';
  const noteKey = noteKeyFor(f);
  const existingNote = noteKey ? (loadNote(noteKey) || '') : '';
  return `
  <div class="finding f-${h(f.color)}${hidden ? ' finding-normal' : ''}" style="${hidden ? 'display:none' : ''}">
    <div class="finding-header">
      <div class="finding-title">${h(f.trait)}</div>
      <div class="finding-meta">
        <code class="gene-tag">${h(f.gene)}</code>
        <span class="rsid-tag">${h(f.rsid)}</span>
        <span class="geno-tag">${h(f.a1)}/${h(f.a2)}</span>
      </div>
      <div class="finding-badge badge-${h(f.color)}">${h(f.badge)}</div>
    </div>
    <div class="finding-body">
      <p class="finding-effect">${h(f.effect)}</p>
      ${f.rec ? `<div class="finding-rec"><span class="rec-label">Recommendation:</span> ${h(f.rec)}</div>` : ''}
      ${existingNote ? `
      <div class="finding-note" id="note-box-${h(noteKey)}">
        <div class="finding-note-label">📝 Your note</div>
        <div class="finding-note-text">${h(existingNote)}</div>
        <div class="finding-note-actions">
          <button class="note-btn-small" data-action="addNote" data-key="${h(noteKey)}">Edit</button>
          <button class="note-btn-small" data-action="deleteNote" data-key="${h(noteKey)}">Delete</button>
        </div>
      </div>` : ''}
    </div>
    <div class="finding-footer">
      ${pmidLink}
      <span class="conf-tag">${h(f.confidence)} confidence</span>
      ${noteKey && !existingNote ? `<button class="note-btn" data-action="addNote" data-key="${h(noteKey)}">📝 Add note</button>` : ''}
    </div>
  </div>`;
}

// ── HTML TEMPLATES ────────────────────────────────────────────────────────────
function uploadScreenHTML() {
  return `
  <div id="upload-screen">
    <div class="topbar">
      <div class="logo">
        <div class="logo-helix">
          <svg viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M8 4 C8 4 24 10 24 16 C24 22 8 28 8 28" stroke="#4dff91" stroke-width="1.5" fill="none" opacity="0.9"/>
            <path d="M24 4 C24 4 8 10 8 16 C8 22 24 28 24 28" stroke="#4dff91" stroke-width="1.5" fill="none" opacity="0.5"/>
            <circle cx="8" cy="9.5" r="1.5" fill="#4dff91"/>
            <circle cx="24" cy="14" r="1.5" fill="#4dff91" opacity="0.6"/>
            <circle cx="8" cy="22.5" r="1.5" fill="#4dff91"/>
          </svg>
        </div>
        <div class="logo-wordmark">Genetica <em>Resolutio</em></div>
      </div>
      <div class="home-tabs">
        <div class="home-tab active" data-tab="analyze"   id="home-tab-analyze">🧬 Analyze</div>
        <div class="home-tab"        data-tab="sessions"  id="home-tab-sessions">📂 Recent</div>
        <div class="home-tab"        data-tab="compare"   id="home-tab-compare">🔀 Compare</div>
        <div class="home-tab"        data-tab="databases" id="home-tab-databases">🗄 Databases</div>
      </div>
      <div class="topbar-right">
        <span class="privacy-tag">100% Local · Zero Network</span>
        <button class="rb-btn" id="home-theme-toggle-btn" title="Toggle light / dark mode">🌓 Theme</button>
      </div>
    </div>

    <!-- ANALYZE PANE -->
    <div id="home-pane-analyze">
      <div class="hero">
        <div class="hero-left">
          <div class="hero-kicker">Desktop Edition · v1.0</div>
          <h1>Your Genome.<br><em>Decoded Privately.</em></h1>
          <p class="hero-desc">
            Upload your raw DNA file and receive a comprehensive, science-backed health analysis —
            cross-referenced against <strong class="stat-snps-val">357+</strong> curated variants spanning
            nutrition, metabolism, fitness, sleep, disease risk, and drug interactions.
            Everything runs locally. Nothing leaves your device.
          </p>
          <div class="stat-row">
            <div class="stat-item"><div class="stat-val stat-snps-val">357+</div><div class="stat-label">Curated Variants</div></div>
            <div class="stat-item"><div class="stat-val">15</div><div class="stat-label">Health Categories</div></div>
            <div class="stat-item"><div class="stat-val">100%</div><div class="stat-label">Offline &amp; Private</div></div>
          </div>
        </div>
        <div class="hero-right">
          <div class="upload-panel">
            <div class="up-label">Upload Raw DNA File</div>
            <div id="drop-zone" class="drop-zone">
              <div class="dz-icon">🧬</div>
              <div class="dz-title">Drop your DNA file here</div>
              <div class="dz-sub">or</div>
              <button class="btn-primary" id="pick-btn">Browse Files</button>
              <div class="dz-formats">
                <span class="pchip">AncestryDNA</span>
                <span class="pchip">23andMe</span>
                <span class="pchip">MyHeritage</span>
                <span class="pchip">FamilyTreeDNA</span>
                <span class="pchip">LivingDNA</span>
                <span class="pchip">VCF</span>
              </div>
            </div>
            <div id="file-ready" class="file-ready">
              <div class="fr-info">
                <div class="fr-icon">📄</div>
                <div>
                  <div class="fr-name" id="fr-name">filename.txt</div>
                  <div class="fr-meta" id="fr-meta">ready to analyze</div>
                </div>
              </div>
              <button class="btn-accent" id="fr-analyze">Analyze →</button>
            </div>
          </div>
          <div class="features">
            <div class="feat"><span class="feat-icon">🔒</span><div><div class="feat-title">Zero network access</div><div class="feat-desc">Your DNA never leaves your device. No cloud, no server, no telemetry.</div></div></div>
            <div class="feat"><span class="feat-icon">🔬</span><div><div class="feat-title">Evidence-based</div><div class="feat-desc">Every variant sourced from GWAS Catalog, ClinVar, PharmGKB with PubMed citations.</div></div></div>
            <div class="feat"><span class="feat-icon">⚡</span><div><div class="feat-title">Personalized Action Plan</div><div class="feat-desc">Diet, supplements, exercise, and sleep recommendations driven by your specific variants.</div></div></div>
          </div>
        </div>
      </div>
    </div>

    <!-- SESSIONS PANE -->
    <div id="home-pane-sessions" style="display:none">
      <div class="db-pane">
        <div class="db-pane-header">
          <div>
            <div class="db-pane-title">Recent Analyses</div>
            <div class="db-pane-sub">
              Saved reports from past analyses. Click to reopen without re-running matching —
              useful when reviewing multiple files or revisiting a report later.
            </div>
          </div>
        </div>
        <div id="sessions-list-wrap">
          <div class="db-loading">Loading sessions…</div>
        </div>
      </div>
    </div>

    <!-- COMPARE PANE -->
    <div id="home-pane-compare" style="display:none">
      <div class="db-pane">
        <div class="db-pane-header">
          <div>
            <div class="db-pane-title">Compare Two DNA Files</div>
            <div class="db-pane-sub">
              Useful for family members (parent/child inheritance, siblings) or for comparing
              the same person across providers (23andMe vs AncestryDNA). Both files are parsed
              locally — nothing is uploaded.
            </div>
          </div>
        </div>
        <div id="compare-pane-body"></div>
      </div>
    </div>

    <!-- DATABASES PANE -->
    <div id="home-pane-databases" style="display:none">
      <div class="db-pane">
        <div class="db-pane-header">
          <div>
            <div class="db-pane-title">Reference Database Manager</div>
            <div class="db-pane-sub">
              Download any combination of the four major public genomic databases to expand your analysis coverage.
              Your DNA is never sent anywhere — only reference data is downloaded.
            </div>
          </div>
        </div>
        <div id="db-table-wrap">
          <div class="db-loading">Loading database status…</div>
        </div>
      </div>
    </div>
  </div>`;
}

function progressScreenHTML() {
  return `
  <div id="progress-screen" class="progress-screen">
    <div class="progress-content">
      <div class="prog-helix">🧬</div>
      <div class="prog-title">Analyzing Your Genome</div>
      <div class="prog-subtitle">Cross-referencing your variants against the curated database…</div>
      <div class="prog-bar-wrap">
        <div class="prog-bar"><div class="prog-fill" id="prog-fill" style="width:0%"></div></div>
        <div class="prog-count" id="prog-count">0%</div>
      </div>
      <div class="prog-label" id="prog-label">Initializing</div>
      <div class="prog-text"  id="prog-text">Starting analysis…</div>
    </div>
  </div>`;
}

function welcomeModalHTML() {
  const { matched, summary } = analysisResult;
  return `
  <div class="welcome-overlay">
    <div class="welcome-modal">
      <div class="wm-icon">🧬</div>
      <h2>Analysis Complete</h2>
      <p>Your genome has been analyzed. <strong>${matched} variants</strong> matched across our curated database.</p>
      <div class="wm-stats">
        <div class="wm-stat wm-red"><div class="wm-stat-val">${(summary.high||[]).length}</div><div>High Risk</div></div>
        <div class="wm-stat wm-amber"><div class="wm-stat-val">${(summary.moderate||[]).length}</div><div>Moderate</div></div>
        <div class="wm-stat wm-green"><div class="wm-stat-val">${(summary.protective||[]).length}</div><div>Protective</div></div>
        <div class="wm-stat wm-blue"><div class="wm-stat-val">${(summary.drugs||[]).length}</div><div>Drug Interactions</div></div>
      </div>
      <p class="wm-disclaimer">
        <strong>Important:</strong> These are statistical tendencies, not diagnoses or certainties.
        Many people with "high risk" variants never develop the associated condition.
        Lifestyle, environment, and other genes all matter enormously.
      </p>
      <div class="wm-actions">
        <button class="btn-accent"     id="welcome-action">⚡ Show My Action Plan</button>
        <button class="btn-primary"    id="welcome-full">View Full Report</button>
        <button class="btn-secondary"  id="welcome-hide-risk">Hide Disease Risk Findings</button>
      </div>
    </div>
  </div>`;
}

function reportShellHTML(parsed, matched, cats, summary) {
  const tabItems = [
    `<div class="tab-item active" data-action="switchTab" data-tab="overview">Overview</div>`,
    `<div class="tab-item" data-action="switchTab" data-tab="action-plan">⚡ Action Plan</div>`,
    ...cats.map(c => `<div class="tab-item" data-action="switchTab" data-tab="${c}">${CAT_META[c]?.icon||'🔬'} ${CAT_META[c]?.title||c}</div>`),
    ...(summary.drugs?.length ? [`<div class="tab-item" data-action="switchTab" data-tab="drugs">💉 Drug Response</div>`] : []),
    `<div class="tab-item" data-action="switchTab" data-tab="lookup">🔎 SNP Lookup</div>`,
    `<div class="tab-item" data-action="switchTab" data-tab="doctor">🩺 Doctor Summary</div>`,
  ].join('');

  const tabPanels = [
    `<div class="rep-section active" id="tab-overview"     data-tab="overview"></div>`,
    `<div class="rep-section"        id="tab-action-plan"  data-tab="action-plan"></div>`,
    ...cats.map(c => `<div class="rep-section" id="tab-${c}" data-tab="${c}"></div>`),
    ...(summary.drugs?.length ? [`<div class="rep-section" id="tab-drugs" data-tab="drugs"></div>`] : []),
    `<div class="rep-section" id="tab-lookup" data-tab="lookup"></div>`,
    `<div class="rep-section" id="tab-doctor" data-tab="doctor"></div>`,
  ].join('');

  return `
  <div id="report-screen">
    <div class="rep-topbar">
      <div class="logo">
        <div class="logo-helix">
          <svg viewBox="0 0 32 32" fill="none">
            <path d="M8 4 C8 4 24 10 24 16 C24 22 8 28 8 28" stroke="#4dff91" stroke-width="1.5" fill="none" opacity="0.9"/>
            <path d="M24 4 C24 4 8 10 8 16 C8 22 24 28 24 28" stroke="#4dff91" stroke-width="1.5" fill="none" opacity="0.5"/>
            <circle cx="8" cy="9.5" r="1.5" fill="#4dff91"/>
            <circle cx="24" cy="14" r="1.5" fill="#4dff91" opacity="0.6"/>
            <circle cx="8" cy="22.5" r="1.5" fill="#4dff91"/>
          </svg>
        </div>
        <div class="logo-wordmark">Genetica <em>Resolutio</em></div>
      </div>
      <div class="rep-toolbar">
        <button class="rb-btn" id="risk-toggle-btn">⚠️ Risk</button>
        <button class="rb-btn" id="sex-filter-btn" title="Toggle sex-specific filtering for this file"></button>
        <button class="rb-btn" id="theme-toggle-btn" title="Toggle light / dark mode">🌓 Theme</button>
        <button class="rb-btn" id="print-btn" title="Print, or 'Save as PDF' in the print dialog">📄 PDF / Print</button>
        <button class="rb-btn" id="save-btn">💾 Save Report</button>
        <button class="rb-btn" data-action="saveCurrentSession" title="Save this analysis so it can be reopened later without re-running matching">📂 Save Session</button>
        <button class="rb-btn" id="new-analysis-btn">↩ New Analysis</button>
      </div>
    </div>
    <div class="rep-hero">
      <div class="rep-eyebrow">// genetica resolutio — genetic health report</div>
      <h1>Your Personal <em>Health Blueprint</em></h1>
    </div>
    <div class="tab-bar">${tabItems}</div>
    <div class="rep-body">${tabPanels}</div>
  </div>`;
}

// ── DATABASE MANAGEMENT ───────────────────────────────────────────────────────

async function loadDatabasesTab() {
  const wrap = document.getElementById('db-table-wrap');
  if (!wrap) return;
  wrap.innerHTML = '<div class="db-loading">Loading…</div>';
  try {
    const sources = await GetDatabaseSources();
    wrap.innerHTML = renderDBTable(sources);
  } catch (e) {
    wrap.innerHTML = `<div class="db-loading" style="color:var(--red)">Failed to load database info: ${h(String(e))}</div>`;
  }
}

function renderDBTable(sources) {
  const rows = sources.map(src => {
    const statusBadge = src.installed
      ? `<span class="db-badge db-badge-installed">✓ Installed</span>
         <div class="db-installed-meta">${(src.rowCount||0).toLocaleString()} records · ${src.downloadedAt||''}</div>`
      : `<span class="db-badge db-badge-not">Not installed</span>`;

    const actionBtn = src.downloading
      ? `<button class="btn-secondary db-action-btn" data-action="cancelDBDownload" data-source="${src.id}">✕ Cancel</button>`
      : src.installed
        ? `<button class="btn-secondary db-action-btn" data-action="deleteDB" data-source="${src.id}">🗑 Remove</button>
           <button class="btn-primary  db-action-btn" data-action="downloadDB" data-source="${src.id}">↻ Re-download</button>`
        : `<button class="btn-accent   db-action-btn" data-action="downloadDB" data-source="${src.id}">↓ Download</button>`;

    const warning = src.warning
      ? `<div class="db-warning">⚠ ${h(src.warning)}</div>`
      : '';

    return `
    <div class="db-row" id="db-row-${src.id}">
      <div class="db-col-name">
        <div class="db-name">${h(src.name)}</div>
        <div class="db-desc">${h(src.description)}</div>
        ${warning}
      </div>
      <div class="db-col-size">
        <div class="db-stat-val">${h(src.fileSize)}</div>
        <div class="db-stat-label">Download size</div>
      </div>
      <div class="db-col-variants">
        <div class="db-stat-val">${h(src.variants)}</div>
        <div class="db-stat-label">Approx. variants</div>
      </div>
      <div class="db-col-coverage">
        <div class="db-coverage">${h(src.coverage)}</div>
      </div>
      <div class="db-col-status">
        ${statusBadge}
      </div>
      <div class="db-col-action">
        ${actionBtn}
      </div>
      <div class="db-progress-row" id="db-progress-${src.id}" style="display:none">
        <div class="db-prog-bar-wrap">
          <div class="db-prog-bar">
            <div class="db-prog-fill" id="db-prog-fill-${src.id}" style="width:0%"></div>
          </div>
          <span class="db-prog-pct" id="db-prog-pct-${src.id}">0%</span>
        </div>
        <div class="db-prog-msg" id="db-prog-msg-${src.id}">Starting…</div>
      </div>
    </div>`;
  }).join('');

  return `
  <div class="db-notice">
    <span>🔒</span>
    <span>Only public scientific reference data is downloaded — your DNA never leaves your device.</span>
    <button class="btn-secondary" style="margin-left:auto" data-action="checkUpdates">↻ Check for updates</button>
  </div>
  <div class="db-table">
    <div class="db-header-row">
      <div class="db-col-name db-hdr">Source</div>
      <div class="db-col-size db-hdr">Size</div>
      <div class="db-col-variants db-hdr">Approx. Variants</div>
      <div class="db-col-coverage db-hdr">Data Coverage</div>
      <div class="db-col-status db-hdr">Status</div>
      <div class="db-col-action db-hdr">Action</div>
    </div>
    ${rows}
  </div>`;
}

async function downloadDB(sourceID) {
  // Show progress row immediately
  const prog = document.getElementById('db-progress-' + sourceID);
  if (prog) prog.style.display = '';
  setDBActionBtn(sourceID, `<button class="btn-secondary db-action-btn" data-action="cancelDBDownload" data-source="${sourceID}">✕ Cancel</button>`);
  try {
    await DownloadDatabase(sourceID);
  } catch (e) {
    showDBError(sourceID, String(e));
  }
}

async function cancelDBDownload(sourceID) {
  await CancelDownload(sourceID);
  const prog = document.getElementById('db-progress-' + sourceID);
  if (prog) prog.style.display = 'none';
  await loadDatabasesTab(); // refresh table
}

async function deleteDB(sourceID) {
  try {
    await DeleteDatabase(sourceID);
    await loadDatabasesTab();
  } catch (e) {
    console.error('Delete failed:', e);
  }
}

function updateDBRow(data) {
  const { sourceID, phase, message, pct } = data;
  const progRow  = document.getElementById('db-progress-' + sourceID);
  const fill     = document.getElementById('db-prog-fill-' + sourceID);
  const pctEl    = document.getElementById('db-prog-pct-'  + sourceID);
  const msgEl    = document.getElementById('db-prog-msg-'  + sourceID);
  if (!progRow) return;

  progRow.style.display = '';
  if (msgEl) msgEl.textContent = message || phase;
  if (fill && pct >= 0) {
    fill.style.width = pct + '%';
    if (pctEl) pctEl.textContent = Math.round(pct) + '%';
  }
}

async function onDBDone(data) {
  const { sourceID, count, error } = data;
  if (error) {
    showDBError(sourceID, error);
    return;
  }
  // Refresh table to show installed status
  await loadDatabasesTab();
  // Refresh DB stats on upload screen
  try {
    dbStats = await GetDBStats();
    updateDBStatDisplay();
  } catch (_) {}
}

function showDBError(sourceID, msg) {
  const msgEl = document.getElementById('db-prog-msg-' + sourceID);
  if (msgEl) {
    msgEl.style.color = 'var(--red)';
    msgEl.textContent = 'Error: ' + msg;
  }
}

function setDBActionBtn(sourceID, html) {
  const col = document.querySelector(`#db-row-${sourceID} .db-col-action`);
  if (col) col.innerHTML = html;
}

// ── ANCESTRY CAVEAT ──────────────────────────────────────────────────────────
const ANC_LABEL = {
  eur: 'European', afr: 'African', eas: 'East Asian',
  sas: 'South Asian', amr: 'Admixed American', any: 'Not specified',
};
function ancestryCaveatHTML() {
  if (!analysisAncestry || analysisAncestry === 'any' || analysisAncestry === 'eur') return '';
  return `
  <div class="anc-banner">
    <span class="anc-icon">🧭</span>
    <div class="anc-text">
      <strong>Ancestry note — ${h(ANC_LABEL[analysisAncestry] || analysisAncestry)}</strong>
      Most GWAS-derived effect sizes were calibrated on European cohorts.
      For ${h(ANC_LABEL[analysisAncestry] || 'this')} ancestry, risk magnitudes may differ.
      Treat these findings as <em>directional</em> rather than precise risk estimates.
    </div>
  </div>`;
}

// ── PER-FINDING NOTES ────────────────────────────────────────────────────────
// Notes are scoped per-file (fileIdent) + per-rsid + per-trait so annotations
// don't leak across different analyses of different people's DNA.
function noteKeyFor(f) {
  if (!fileIdent || !f || !f.rsid) return '';
  return `gr-note|${fileIdent}|${f.rsid}|${normalizeTrait(f.trait)}`;
}
function loadNote(key) {
  try { return localStorage.getItem(key); } catch (_) { return null; }
}
function saveNote(key, text) {
  try {
    if (text) localStorage.setItem(key, text);
    else localStorage.removeItem(key);
  } catch (_) {}
}
function promptNote(key, btn) {
  const current = loadNote(key) || '';
  // Build an inline editor right next to the button rather than window.prompt,
  // which Wails/WebKit don't always render nicely.
  const card = btn.closest('.finding');
  if (!card) return;
  // Remove any existing editor in this card
  const existing = card.querySelector('.note-editor');
  if (existing) { existing.remove(); return; }
  const editor = document.createElement('div');
  editor.className = 'note-editor';
  editor.innerHTML = `
    <textarea class="note-textarea" placeholder="e.g. discussed with Dr. Smith, retest 2027">${h(current)}</textarea>
    <div class="note-editor-actions">
      <button class="btn-secondary note-btn-small" data-role="save">Save</button>
      <button class="btn-secondary note-btn-small" data-role="cancel">Cancel</button>
    </div>`;
  card.querySelector('.finding-body').appendChild(editor);
  const ta = editor.querySelector('textarea');
  ta.focus();
  editor.querySelector('[data-role="save"]').addEventListener('click', () => {
    const v = ta.value.trim();
    saveNote(key, v);
    editor.remove();
    rerenderCurrentTabPreservingNotes();
  });
  editor.querySelector('[data-role="cancel"]').addEventListener('click', () => editor.remove());
}
function deleteNote(key) {
  saveNote(key, '');
  rerenderCurrentTabPreservingNotes();
}
function rerenderCurrentTabPreservingNotes() {
  // Re-render just the active tab. Simplest is a full renderReport.
  const active = document.querySelector('.tab-item.active')?.dataset.tab || 'overview';
  renderReport(active);
}

// ── RSID LOOKUP TAB ──────────────────────────────────────────────────────────
function renderLookupTab() {
  const el = document.getElementById('tab-lookup');
  if (!el) return;
  el.innerHTML = `
    <div class="sec-hdr">
      <div class="sec-num">🔎</div>
      <div class="sec-title">SNP Lookup</div>
      <div class="sec-badge">Search any rsID</div>
    </div>
    <p style="font-size:13px;color:var(--ink2);line-height:1.7;margin-bottom:20px">
      Paste any rsID (e.g. <code>rs1815739</code>) to see all matching annotations from
      your installed reference databases, plus your genotype if this SNP was in your file.
    </p>
    <div class="lookup-form">
      <input type="text" id="lookup-input" class="lookup-input"
             placeholder="rs1801133"
             autocomplete="off" spellcheck="false"/>
      <button class="btn-accent" data-action="runLookup">Search</button>
    </div>
    <div id="lookup-results" class="lookup-results"></div>`;
  const input = document.getElementById('lookup-input');
  input.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') runRsidLookup();
  });
}
async function runRsidLookup() {
  const input = document.getElementById('lookup-input');
  const out   = document.getElementById('lookup-results');
  if (!input || !out) return;
  const raw = input.value.trim().toLowerCase();
  if (!raw) return;
  const rsid = raw.startsWith('rs') ? raw : 'rs' + raw.replace(/^rs/i, '');
  out.innerHTML = '<div class="db-loading">Searching…</div>';
  let userGeno = null;
  try {
    if (analysisResult?.parsed?.snps) {
      const g = analysisResult.parsed.snps[rsid];
      if (g) userGeno = `${g[0]}/${g[1]}`;
    }
  } catch (_) {}
  try {
    const recs = await LookupRSID(rsid);
    if (!recs || recs.length === 0) {
      out.innerHTML = `
        <div class="empty-state">
          <div class="empty-state-icon">🔍</div>
          <p>No annotations found for <code>${h(rsid)}</code> in installed databases.</p>
          ${userGeno ? `<p>Your genotype for this SNP: <strong>${h(userGeno)}</strong></p>` : ''}
        </div>`;
      return;
    }
    out.innerHTML = `
      <div class="lookup-header">
        <div class="lookup-rsid">${h(rsid)}</div>
        ${userGeno ? `<div class="lookup-geno">Your genotype: <strong>${h(userGeno)}</strong></div>`
                   : `<div class="lookup-geno">Not in your uploaded file</div>`}
        <div class="lookup-count">${recs.length} annotation${recs.length === 1 ? '' : 's'} across installed databases</div>
      </div>
      <div class="lookup-list">
        ${recs.map(r => `
        <div class="lookup-card">
          <div class="lookup-card-head">
            <strong>${h(r.trait || '(no trait)')}</strong>
            ${r.gene ? `<code class="gene-tag">${h(r.gene)}</code>` : ''}
            ${r.riskAllele ? `<span class="geno-tag">Risk allele: ${h(r.riskAllele)}</span>` : ''}
          </div>
          <div class="lookup-card-body">
            ${r.desc ? `<div>${h(r.desc)}</div>` : ''}
            ${r.rec ? `<div style="margin-top:6px;color:var(--amber)"><strong>Rec:</strong> ${h(r.rec)}</div>` : ''}
          </div>
          <div class="lookup-card-foot">
            ${r.cat ? `<span>${h(CAT_META[r.cat]?.title || r.cat)}</span>` : ''}
            ${r.pmid ? `<a href="https://pubmed.ncbi.nlm.nih.gov/${h(r.pmid)}" target="_blank" class="pubmed-link">PubMed ${h(r.pmid)}</a>` : ''}
            ${r.conf ? `<span class="conf-tag">${h(r.conf)}</span>` : ''}
          </div>
        </div>`).join('')}
      </div>`;
  } catch (e) {
    out.innerHTML = `<div style="color:var(--red)">Lookup failed: ${h(String(e))}</div>`;
  }
}

// ── SESSIONS ─────────────────────────────────────────────────────────────────
async function loadSessionsTab() {
  const wrap = document.getElementById('sessions-list-wrap');
  if (!wrap) return;
  wrap.innerHTML = '<div class="db-loading">Loading…</div>';
  try {
    const list = await ListSessions();
    if (!list || list.length === 0) {
      wrap.innerHTML = `
        <div class="empty-state">
          <div class="empty-state-icon">📂</div>
          <p>No saved sessions yet. After analyzing a file, click "💾 Save Session" in the report toolbar to save it here.</p>
        </div>`;
      return;
    }
    wrap.innerHTML = `
      <div class="sess-list">
        ${list.map(s => `
        <div class="sess-row" data-action="openSession" data-id="${h(s.id)}">
          <div class="sess-icon">📄</div>
          <div class="sess-main">
            <div class="sess-filename">${h(s.filename || '(unnamed)')}</div>
            <div class="sess-meta">
              <span>${h(s.provider || '')}</span>
              <span>·</span>
              <span>${(s.matched||0).toLocaleString()} variants matched</span>
              <span>·</span>
              <span>saved ${h(s.savedAt || '')}</span>
            </div>
          </div>
          <div class="sess-stats">
            ${s.highCount ? `<span class="sess-stat sess-stat-red">${s.highCount} high</span>` : ''}
            ${s.modCount  ? `<span class="sess-stat sess-stat-amber">${s.modCount} mod</span>` : ''}
          </div>
          <button class="btn-secondary" data-action="deleteSession" data-id="${h(s.id)}" onclick="event.stopPropagation()">🗑</button>
        </div>`).join('')}
      </div>`;
    // Stop-propagation on the delete button so clicking delete doesn't also open.
    wrap.querySelectorAll('[data-action="deleteSession"]').forEach(b => {
      b.addEventListener('click', (e) => e.stopPropagation());
    });
  } catch (e) {
    wrap.innerHTML = `<div style="color:var(--red)">Failed to load sessions: ${h(String(e))}</div>`;
  }
}
async function saveCurrentSession() {
  if (!analysisResult) return;
  const btn = document.querySelector('[data-action="saveCurrentSession"]');
  const orig = btn ? btn.textContent : '';
  if (btn) {
    btn.disabled = true;
    btn.textContent = '⏳ Saving…';
    btn.style.opacity = '0.7';
  }
  showToast('Saving session…', 'info', 0);
  try {
    const name = fileIdent ? fileIdent.split('|')[0] : 'analysis';
    await SaveSession(analysisResult, name);
    hideToast();
    showToast('✓ Session saved', 'ok', 2500);
    if (btn) {
      btn.textContent = '✓ Saved';
      setTimeout(() => {
        btn.textContent = orig;
        btn.disabled = false;
        btn.style.opacity = '';
      }, 2500);
    }
  } catch (e) {
    hideToast();
    console.error('SaveSession failed:', e);
    showToast('Failed to save session', 'error', 4000);
    if (btn) { btn.textContent = orig; btn.disabled = false; btn.style.opacity = ''; }
  }
}
async function openSession(id) {
  // Render a loading screen immediately so the UI doesn't look frozen while
  // the (potentially large) session JSON is deserialised on the backend.
  document.getElementById('app').innerHTML = `
    <div id="progress-screen" class="progress-screen">
      <div class="progress-content">
        <div class="prog-helix">🧬</div>
        <div class="prog-title">Loading Saved Session</div>
        <div class="prog-subtitle">Deserialising report from disk…</div>
        <div class="prog-bar-wrap">
          <div class="prog-bar"><div class="prog-fill" style="width:100%;animation: pulseBar 1.2s ease-in-out infinite"></div></div>
        </div>
        <div class="prog-label">Reading</div>
        <div class="prog-text">This can take a few seconds for large reports.</div>
      </div>
    </div>`;
  try {
    const result = await LoadSession(id);
    analysisResult = result;
    fileIdent = (result.parsed?.provider || 'session') + '|0';
    analysisSex = 'any'; chosenSex = 'any'; analysisAncestry = 'any';
    renderReport('overview');
  } catch (e) {
    console.error('LoadSession failed:', e);
    document.getElementById('app').innerHTML = `
      <div style="display:flex;align-items:center;justify-content:center;height:100vh;flex-direction:column;gap:16px">
        <div style="color:var(--red);font-size:18px">Failed to load session</div>
        <div style="color:var(--ink2);font-size:13px;max-width:400px;text-align:center">${h(String(e))}</div>
        <button class="btn-primary" data-action="reload">Back to Home</button>
      </div>`;
  }
}

// ── TOAST ────────────────────────────────────────────────────────────────────
function showToast(msg, kind = 'info', autoHideMs = 0) {
  hideToast();
  const t = document.createElement('div');
  t.id = 'gr-toast';
  t.className = `gr-toast gr-toast-${kind}`;
  t.textContent = msg;
  document.body.appendChild(t);
  if (autoHideMs > 0) setTimeout(hideToast, autoHideMs);
}
function hideToast() {
  const t = document.getElementById('gr-toast');
  if (t) t.remove();
}
async function confirmDeleteSession(id) {
  try {
    await DeleteSession(id);
    await loadSessionsTab();
  } catch (e) {
    console.error('DeleteSession failed:', e);
  }
}

// ── DATABASE UPDATE CHECKER ──────────────────────────────────────────────────
async function checkDatabaseUpdates() {
  try {
    const info = await CheckDatabaseUpdates();
    if (!info) return;
    for (const [srcID, v] of Object.entries(info)) {
      const statusCol = document.querySelector(`#db-row-${srcID} .db-col-status`);
      if (!statusCol) continue;
      const old = statusCol.querySelector('.db-update-note');
      if (old) old.remove();
      const note = document.createElement('div');
      note.className = 'db-update-note';
      if (v.updateAvail) {
        note.innerHTML = `<span class="db-badge db-badge-update">↑ Update available</span>`;
      } else if (v.checked) {
        note.innerHTML = `<span class="db-uptodate">✓ Up to date</span>`;
      } else {
        note.innerHTML = `<span class="db-uptodate">? Couldn't check</span>`;
      }
      statusCol.appendChild(note);
    }
  } catch (e) {
    console.error('CheckDatabaseUpdates failed:', e);
  }
}

// ── COMPARE TWO FILES ────────────────────────────────────────────────────────
let compareA = null;
let compareB = null;
function renderComparePane() {
  const el = document.getElementById('compare-pane-body');
  if (!el) return;
  el.innerHTML = `
    <div class="compare-slots">
      <div class="compare-slot" id="compare-slot-a">
        <div class="compare-slot-label">File A</div>
        <div class="compare-slot-name">${compareA ? h(compareA.split('/').pop()) : '(no file)'}</div>
        <button class="btn-primary" id="compare-pick-a">Choose file A…</button>
      </div>
      <div class="compare-arrow">⇄</div>
      <div class="compare-slot" id="compare-slot-b">
        <div class="compare-slot-label">File B</div>
        <div class="compare-slot-name">${compareB ? h(compareB.split('/').pop()) : '(no file)'}</div>
        <button class="btn-primary" id="compare-pick-b">Choose file B…</button>
      </div>
    </div>
    <div class="compare-actions">
      <button class="btn-accent" id="compare-run" ${compareA && compareB ? '' : 'disabled'}>Run comparison →</button>
    </div>
    <div id="compare-results"></div>`;
  document.getElementById('compare-pick-a').addEventListener('click', () => pickCompareFile('a'));
  document.getElementById('compare-pick-b').addEventListener('click', () => pickCompareFile('b'));
  document.getElementById('compare-run').addEventListener('click', runCompare);
}
async function pickCompareFile(slot) {
  try {
    const path = await OpenFileDialog();
    if (!path) return;
    if (slot === 'a') compareA = path;
    else              compareB = path;
    renderComparePane();
  } catch (e) {
    console.error(e);
  }
}
async function runCompare() {
  const out = document.getElementById('compare-results');
  out.innerHTML = '<div class="db-loading">Parsing both files and comparing — this may take a minute…</div>';
  try {
    const res = await CompareFiles(compareA, compareB);
    renderCompareResult(res, out);
  } catch (e) {
    out.innerHTML = `<div style="color:var(--red)">Comparison failed: ${h(String(e))}</div>`;
  }
}
function renderCompareResult(r, out) {
  if (!r) { out.innerHTML = ''; return; }
  const rows = r.rows || [];
  out.innerHTML = `
    <div class="compare-stats">
      <div class="compare-stat"><div class="compare-stat-val">${(r.commonSNPs||0).toLocaleString()}</div><div>SNPs in both</div></div>
      <div class="compare-stat"><div class="compare-stat-val">${(r.identical||0).toLocaleString()}</div><div>Identical genotype</div></div>
      <div class="compare-stat compare-stat-amber"><div class="compare-stat-val">${(r.differ||0).toLocaleString()}</div><div>Different genotype</div></div>
      <div class="compare-stat"><div class="compare-stat-val">${(r.onlyA||0).toLocaleString()}</div><div>Only in A</div></div>
      <div class="compare-stat"><div class="compare-stat-val">${(r.onlyB||0).toLocaleString()}</div><div>Only in B</div></div>
    </div>
    <h3 style="margin:28px 0 12px;font-size:14px;color:var(--ink)">Annotated variants where A and B differ (${rows.length})</h3>
    ${rows.length === 0 ? `
      <div class="empty-state">
        <div class="empty-state-icon">✓</div>
        <p>No differences in annotated variants. Any overall genotype differences above are in non-curated SNPs.</p>
      </div>` : `
    <div class="compare-table">
      <div class="compare-row compare-head">
        <div>Trait</div><div>Gene</div><div>rsID</div><div>A</div><div>B</div><div>Status</div>
      </div>
      ${rows.slice(0, 500).map(row => `
      <div class="compare-row">
        <div>${h(row.trait || '')}</div>
        <div><code class="gene-tag">${h(row.gene || '')}</code></div>
        <div><span class="rsid-tag">${h(row.rsid || '')}</span></div>
        <div><span class="geno-tag">${h(row.a1Geno || '')}</span></div>
        <div><span class="geno-tag">${h(row.a2Geno || '')}</span></div>
        <div style="font-size:11px;color:var(--ink2)">${h((row.notes || []).join('; '))}</div>
      </div>`).join('')}
      ${rows.length > 500 ? `<div style="padding:10px;color:var(--muted);font-size:12px">Showing first 500 of ${rows.length.toLocaleString()}</div>` : ''}
    </div>`}`;
}
function startCompareFlow() { switchHomeTab('compare'); }
function pickCompareFileB() { pickCompareFile('b'); }
