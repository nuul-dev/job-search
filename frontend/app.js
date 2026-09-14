const $ = (id) => document.getElementById(id);
function syncThemeToggle() {
  const dark = document.documentElement.dataset.theme === 'dark';
  $('theme-toggle').setAttribute('aria-pressed', String(dark));
  $('theme-icon').textContent = dark ? '☾' : '☀';
}
$('theme-toggle').onclick = () => {
  const theme = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark';
  document.documentElement.dataset.theme = theme;
  syncThemeToggle();
  try { localStorage.setItem('job-inbox-theme', theme); }
  catch { notice('Тема изменена, но браузер не сохранил выбор. Разрешите локальное хранилище для сохранения темы.'); }
};
syncThemeToggle();
const labels = {all: 'Вакансии', saved: 'Избранное', applied: 'Откликнулся', hidden: 'Отложено'};
let runs = [], jobs = [], selected = null, view = 'all', marks = {};
function notice(message) { $('notice').textContent = message; $('notice').hidden = !message; }
try { marks = JSON.parse(localStorage.getItem('job-inbox-v1') || '{}'); if (!marks || typeof marks !== 'object' || Array.isArray(marks)) marks = {}; }
catch { notice('Не удалось прочитать заметки в браузере. Новые изменения можно сохранить, если хранилище доступно.'); }
function save() { try { localStorage.setItem('job-inbox-v1', JSON.stringify(marks)); } catch { notice('Браузер не сохранил изменения. Разрешите локальное хранилище; пока не закрывайте страницу.'); } }
function mark(job) { return marks[job.id] || {}; }
function el(tag, className, text) { const node = document.createElement(tag); if (className) node.className = className; if (text !== undefined) node.textContent = text; return node; }
function hasSalary(job) { return job.salary && !/не указан|не указана|not specified/i.test(job.salary); }
function safeURL(value) { try { const url = new URL(value); return ['https:', 'http:'].includes(url.protocol) ? url.href : null; } catch { return null; } }
function filtered() {
  const query = $('search').value.trim().toLocaleLowerCase();
  return jobs.filter(j => (!query || `${j.title} ${j.company} ${j.description}`.toLocaleLowerCase().includes(query)) && (!$('run').value || j.runs.includes($('run').value)) && (!$('source').value || j.source === $('source').value) && (!$('remote').checked || j.remote) && (!$('salary').checked || hasSalary(j)) && (view === 'all' ? mark(j).status !== 'hidden' : view === 'saved' ? mark(j).saved : mark(j).status === view));
}
function render() {
  for (const key of Object.keys(labels)) $(key + '-count').textContent = jobs.filter(j => key === 'all' ? mark(j).status !== 'hidden' : key === 'saved' ? mark(j).saved : mark(j).status === key).length;
  const visible = filtered();
  if (!visible.some(j => j.id === selected)) selected = visible[0]?.id || null;
  $('result-count').textContent = `Найдено: ${visible.length} из ${jobs.length}`;
  $('list').replaceChildren();
  if (!visible.length) $('list').append(el('div', 'empty', jobs.length ? 'Подходящих вакансий нет. Попробуйте изменить фильтры или выбрать другой раздел.' : 'Подборок пока нет. Нажмите «Найти вакансии», чтобы собрать первую подборку.'));
  for (const job of visible) {
    const button = el('button', 'job' + (selected === job.id ? ' selected' : ''));
    button.setAttribute('aria-pressed', String(selected === job.id));
    const company = el('div', 'company', job.company || 'Компания не указана');
    if (mark(job).saved) company.append(el('span', 'saved-star', '★'));
    button.append(company, el('h2', '', job.title || 'Без названия'), el('p', 'salary', hasSalary(job) ? job.salary : 'Зарплата не указана'));
    const meta = el('div', 'meta'); meta.append(el('span', '', job.source || 'Источник не указан'), el('span', '', job.remote ? 'Удалённо' : 'Формат не уточнён'));
    if (mark(job).status === 'applied') meta.append(el('span', '', 'Откликнулся'));
    button.append(meta); button.onclick = () => { selected = job.id; render(); }; $('list').append(button);
  }
  renderDetail(visible.find(j => j.id === selected));
}
function renderDetail(job) {
  const panel = $('detail'); panel.replaceChildren();
  if (!job) { panel.append(el('div', 'empty', 'Здесь будут подробности выбранной вакансии')); return; }
  panel.append(el('p', 'detail-company', job.company), el('h2', '', job.title), el('p', 'detail-salary', hasSalary(job) ? job.salary : 'Зарплата не указана'));
  const meta = el('div', 'meta'); meta.append(el('span', '', job.source), el('span', '', job.remote ? 'Удалённая работа' : 'Формат работы не уточнён')); panel.append(meta);
  const actions = el('div', 'actions'); const url = safeURL(job.url);
  if (url) { const link = el('a', 'primary', 'Открыть вакансию ↗'); link.href = url; link.target = '_blank'; link.rel = 'noopener noreferrer'; actions.append(link); }
  const favorite = el('button', 'secondary favorite', mark(job).saved ? '★ В избранном' : '☆ В избранное'); favorite.setAttribute('aria-pressed', String(!!mark(job).saved)); favorite.onclick = () => { marks[job.id] = {...mark(job), saved: !mark(job).saved}; save(); render(); }; actions.append(favorite); panel.append(actions);
  const description = el('section', 'detail-section'); description.append(el('h3', '', 'О вакансии'));
  // Descriptions are untrusted source content: convert markup to inert plain text.
  const parsed = new DOMParser().parseFromString(String(job.description || '').replace(/<\/(p|div|li|h[1-6])>|<br\s*\/?\s*>/gi, '\n'), 'text/html');
  parsed.querySelectorAll('script,style,iframe').forEach(n => n.remove());
  description.append(el('div', 'description', parsed.body.textContent.trim() || 'В подборке нет описания. Откройте вакансию на сайте, чтобы посмотреть требования и условия.')); panel.append(description);
  const personal = el('section', 'detail-section'); personal.append(el('h3', '', 'Мои заметки'));
  const statusLabel = el('label', '', 'Статус'); statusLabel.htmlFor = 'status'; const status = el('select'); status.id = 'status';
  for (const [value, text] of [['new','Не разобрано'],['applied','Откликнулся'],['hidden','Отложено']]) { const option = el('option', '', text); option.value = value; status.append(option); }
  status.value = mark(job).status || 'new'; status.onchange = () => { marks[job.id] = {...mark(job), status: status.value}; save(); render(); };
  const noteLabel = el('label', '', 'Что хочу уточнить или обсудить'); noteLabel.htmlFor = 'note'; const note = el('textarea'); note.id = 'note'; note.placeholder = 'Например: узнать о команде и этапах интервью'; note.value = mark(job).note || '';
  note.oninput = () => { marks[job.id] = {...mark(job), note: note.value}; save(); };
  personal.append(statusLabel, status, noteLabel, note, el('p', 'hint', 'Сохраняется автоматически в этом браузере. Статус — ваша отметка; отправить отклик можно на сайте вакансии.')); panel.append(personal);
}
async function load() {
  $('refresh').disabled = true;
  try {
    const response = await fetch('/api/runs'); if (!response.ok) throw new Error();
    const data = await response.json(); runs = data.runs;
    const unique = new Map();
    for (const run of runs) for (const vacancy of run.vacancies) {
      const id = safeURL(vacancy.url) || `${vacancy.source}:${vacancy.company}:${vacancy.title}`;
      if (unique.has(id)) unique.get(id).runs.push(run.id);
      else unique.set(id, {...vacancy, id, runs: [run.id]});
    }
    jobs = [...unique.values()];
    for (const [id, options] of [['run', runs.map(r => [r.id, r.id.replace(/^(\d{4})-(\d{2})-(\d{2})-(\d{2})(\d{2})$/, '$3.$2.$1 $4:$5')])], ['source', [...new Set(jobs.map(j => j.source).filter(Boolean))].sort().map(s => [s,s])]]) {
      const select = $(id), previous = select.value; while (select.options.length > 1) select.remove(1);
      for (const [value, text] of options) { const option = el('option', '', text); option.value = value; select.append(option); }
      select.value = options.some(([v]) => v === previous) ? previous : '';
    }
    if (data.errors.length) notice(`Не удалось прочитать подборки: ${data.errors.join(', ')}. Проверьте файлы и обновите страницу.`);
    render();
  } catch { notice('Не удалось загрузить вакансии. Проверьте, что локальный сервер запущен, и нажмите «Обновить подборки».'); $('result-count').textContent = 'Ошибка загрузки'; }
  finally { $('refresh').disabled = false; }
}
document.querySelectorAll('[data-view]').forEach(button => button.onclick = () => { view = button.dataset.view; $('page-title').textContent = labels[view]; document.querySelectorAll('[data-view]').forEach(b => b.classList.toggle('active', b === button)); render(); });
for (const id of ['search', 'run', 'source', 'remote', 'salary']) $(id).addEventListener(id === 'search' ? 'input' : 'change', render);
$('reset').onclick = () => { for (const id of ['search','run','source']) $(id).value = ''; for (const id of ['remote','salary']) $(id).checked = false; render(); };
$('refresh').onclick = load;
let searchTimer, previousSearch = '', searchRequestPending = false, searchPollPending = false;
function showSearch(state) {
  const messages = {
    idle: 'Готов к поиску новых вакансий.',
    running: 'Ищем вакансии и готовим отчёт. Можно продолжать разбирать подборки.',
    succeeded: 'Поиск завершён. Подборки обновлены.',
    failed: 'Поиск завершился с ошибкой. ' + (state.error || 'Попробуйте запустить его снова.'),
  };
  $('search-status').textContent = messages[state.status] || 'Состояние поиска неизвестно.';
  $('search-status').dataset.state = state.status;
  $('start-search').disabled = searchRequestPending || state.status === 'running';
  $('start-search').textContent = state.status === 'running' ? 'Поиск идёт…' : 'Найти вакансии';
  const key = `${state.status}:${state.started_at || ''}`;
  if (['succeeded', 'failed'].includes(state.status) && previousSearch !== key) load();
  previousSearch = key;
}
async function pollSearch() {
  if (searchPollPending) return;
  searchPollPending = true;
  clearTimeout(searchTimer);
  try {
    const response = await fetch('/api/search');
    if (!response.ok) throw new Error();
    showSearch(await response.json());
  } catch {
    $('search-status').textContent = 'Нет связи с поиском. Проверяем подключение…';
    $('start-search').disabled = true;
  } finally {
    searchPollPending = false;
    searchTimer = setTimeout(pollSearch, 3000);
  }
}
$('start-search').onclick = async () => {
  searchRequestPending = true;
  $('start-search').disabled = true;
  $('search-status').textContent = 'Запускаем поиск…';
  try {
    const response = await fetch('/api/search', {method: 'POST', headers: {'Content-Type': 'application/json'}, body: '{}'});
    if (!response.ok && response.status !== 409) throw new Error();
  } catch {
    notice('Не удалось подтвердить запуск поиска. Проверяем его состояние; повторно нажать кнопку можно после восстановления связи.');
  } finally {
    searchRequestPending = false;
    pollSearch();
  }
};
load();
pollSearch();
