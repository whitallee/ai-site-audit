const form = document.getElementById('audit-form');
const urlInput = document.getElementById('url-input');
const submitBtn = document.getElementById('submit-btn');
const loading = document.getElementById('loading');
const errorBox = document.getElementById('error-box');
const results = document.getElementById('results');

form.addEventListener('submit', async (e) => {
  e.preventDefault();
  const url = urlInput.value.trim();
  if (!url) return;
  await runAudit(url);
});

async function runAudit(url) {
  setLoading(true);
  hideError();
  results.classList.add('hidden');

  try {
    const res = await fetch('/api/audit', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url }),
    });

    const text = await res.text();
    if (!res.ok) throw new Error(text || `Server error: ${res.status}`);

    const data = JSON.parse(text);
    renderResults(data.id, data.audit);
  } catch (err) {
    showError(err.message);
  } finally {
    setLoading(false);
  }
}

function renderResults(id, audit) {
  document.getElementById('overall-score').textContent = audit.score;
  document.getElementById('summary').textContent = audit.summary;
  document.getElementById('pdf-link').href = `/api/audit/${id}/pdf`;

  const sections = [
    { name: 'SEO',         data: audit.seo },
    { name: 'UX',          data: audit.ux },
    { name: 'Performance', data: audit.performance },
    { name: 'Conversion',  data: audit.conversion },
  ];

  const sectionsEl = document.getElementById('sections');
  sectionsEl.innerHTML = sections.map(({ name, data }) => `
    <div class="section-card">
      <div class="section-top">
        <span class="section-name">${name}</span>
        <span class="section-score ${scoreClass(data.score)}">${data.score}/100</span>
      </div>
      <div class="score-bar-bg">
        <div class="score-bar-fill" style="width:${data.score}%; background:${scoreColor(data.score)}"></div>
      </div>
      ${list('Issues', data.issues)}
      ${list('Recommendations', data.recommendations)}
    </div>
  `).join('');

  const qwEl = document.getElementById('quick-wins');
  qwEl.innerHTML = (audit.quick_wins || []).map(w => `<li>${w}</li>`).join('');

  results.classList.remove('hidden');
  results.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

function list(label, items) {
  if (!items || items.length === 0) return '';
  return `
    <div class="section-label">${label}</div>
    <ul>${items.map(i => `<li>${i}</li>`).join('')}</ul>
  `;
}

function scoreClass(score) {
  if (score >= 70) return 'good';
  if (score >= 45) return 'warn';
  return 'bad';
}

function scoreColor(score) {
  if (score >= 70) return '#16a34a';
  if (score >= 45) return '#d97706';
  return '#dc2626';
}

function setLoading(on) {
  loading.classList.toggle('hidden', !on);
  submitBtn.disabled = on;
  submitBtn.textContent = on ? 'Analyzing…' : 'Analyze';
}

function showError(msg) {
  errorBox.textContent = `Error: ${msg}`;
  errorBox.classList.remove('hidden');
}

function hideError() {
  errorBox.classList.add('hidden');
}
