// Draft editor state survives list redraws and switching between vacancies.
const letters = new Map();
let availableResumes = null;
async function draftRequest(path, options) {
  const response = await fetch(path, options);
  let data;
  try { data = await response.json(); } catch { throw new Error('Сервер не вернул ответ. Проверьте подключение и повторите.'); }
  if (!response.ok) { const error = new Error(data.error || 'Не удалось выполнить запрос. Повторите попытку.'); error.status = response.status; throw error; }
  return data;
}
function draftJSON(method, value) { return {method, headers: {'Content-Type': 'application/json'}, body: JSON.stringify(value)}; }
function updateLetter(job) {
  if (selected !== job.id) return;
  const panel = $('detail').querySelector('.letter-editor');
  if (panel) renderLetter(panel, job);
}
async function openLetter(job) {
  if (!letters.has(job.id)) {
    letters.set(job.id, {resume: '', description: job.description || '', text: '', message: '', busy: false, dirty: false, loading: true});
  }
  renderDetail(job);
  $('detail').querySelector('.letter-editor')?.scrollIntoView({block: 'nearest'});
  const state = letters.get(job.id);
  if (availableResumes === null) {
    try { availableResumes = (await draftRequest('/api/resumes')).resumes; }
    catch (error) { state.message = error.message; }
  }
  state.loading = false;
  if (!state.resume && availableResumes?.length) {
    let preferred = ''; try { preferred = localStorage.getItem('job-inbox-preferred-resume') || ''; } catch {}
    state.resume = availableResumes.includes(preferred) ? preferred : availableResumes[0];
  }
  updateLetter(job);
}
function renderLetter(panel, job) {
  const state = letters.get(job.id);
  panel.replaceChildren(el('h3', '', 'Сопроводительное письмо'));
  panel.append(el('p', 'hint', 'Codex подготовит короткий отклик по выбранному резюме. Проверьте текст перед отправкой.'));
  const resumeLabel = el('label', '', 'Резюме для этого отклика'); resumeLabel.htmlFor = 'letter-resume';
  const resume = el('select'); resume.id = 'letter-resume'; resume.disabled = state.busy || state.loading;
  for (const filename of availableResumes || []) { const option = el('option', '', filename); option.value = filename; resume.append(option); }
  resume.value = state.resume; resume.onchange = () => { state.resume = resume.value; };
  panel.append(resumeLabel, resume);
  if (state.loading) panel.append(el('p', 'hint', 'Загружаем список резюме…'));
  else if (!availableResumes?.length) {
    panel.append(el('p', 'hint', 'Добавьте PDF или Markdown-резюме в папку resumes, затем обновите список.'));
    const retry = el('button', 'secondary', 'Обновить резюме');
    retry.onclick = () => { availableResumes = null; state.loading = true; openLetter(job); }; panel.append(retry);
  }
  const descriptionLabel = el('label', '', 'Описание вакансии'); descriptionLabel.htmlFor = 'letter-description';
  const description = el('textarea'); description.id = 'letter-description'; description.value = state.description;
  description.placeholder = 'Если описания нет в подборке, вставьте требования и задачи с сайта вакансии.';
  description.maxLength = 20000; description.disabled = state.busy;
  description.oninput = () => { state.description = description.value; };
  panel.append(descriptionLabel, description);
  if (!state.description.trim()) panel.append(el('p', 'hint', 'Без описания письмо будет опираться только на название вакансии и резюме.'));
  const generate = el('button', 'primary', state.busy ? 'Готовим письмо…' : state.pendingId ? 'Проверить результат' : state.text ? 'Создать новый вариант' : 'Создать черновик');
  generate.disabled = state.busy || state.loading || !state.resume || state.dirty;
  generate.onclick = () => generateLetter(job); panel.append(generate);
  if (state.busy) { const progress = el('progress', 'letter-progress'); progress.setAttribute('aria-label', 'Генерация письма'); panel.append(progress); }
  if (state.text || state.id) {
    const label = el('label', '', 'Текст письма'); label.htmlFor = 'letter-text';
    const text = el('textarea', 'letter-text'); text.id = 'letter-text'; text.value = state.text; text.disabled = state.busy || state.saving;
    const actions = el('div', 'actions');
    const copy = el('button', 'secondary', 'Скопировать');
    const saveButton = el('button', 'primary', state.saving ? 'Сохраняем…' : 'Сохранить изменения');
    saveButton.disabled = !state.dirty || !state.text.trim() || state.saving || state.busy;
    const feedback = el('p', 'hint', state.dirty ? 'Есть несохранённые изменения.' : 'Черновик сохранён. Отправка выполняется вами на сайте вакансии.');
    text.oninput = () => { state.text = text.value; state.dirty = true; saveButton.disabled = !state.text.trim(); generate.disabled = true; feedback.textContent = 'Есть несохранённые изменения. Сохраните их перед созданием нового варианта.'; };
    copy.onclick = async () => {
      try { await navigator.clipboard.writeText(state.text); feedback.textContent = 'Текст скопирован.'; }
      catch { text.focus(); text.select(); feedback.textContent = 'Текст выделен. Нажмите Ctrl+C или ⌘C.'; }
    };
    saveButton.onclick = async () => {
      state.saving = true; state.message = ''; updateLetter(job);
      try {
        const result = await draftRequest(`/api/drafts/${encodeURIComponent(state.id)}`, draftJSON('PATCH', {text: state.text}));
        state.path = result.path; state.dirty = false; state.message = 'Изменения сохранены.';
      } catch (error) { state.message = error.message; }
      finally { state.saving = false; updateLetter(job); }
    };
    actions.append(copy, saveButton); panel.append(label, text, actions, feedback);
  }
  if (state.path) panel.append(el('p', 'hint', `Файл черновика: ${state.path}`));
  const message = el('p', 'hint', state.message); message.setAttribute('role', 'status'); panel.append(message);
}
async function generateLetter(job) {
  const state = letters.get(job.id);
  if (state.busy || state.dirty || !state.resume) return;
  state.busy = true; state.message = 'Пишем письмо. Это может занять пару минут.'; updateLetter(job);
  try {
    const vacancy = Object.fromEntries(['title','company','url','source','salary','remote'].map(key => [key, job[key]]));
    vacancy.description = state.description;
    let result = state.pendingId
      ? await draftRequest(`/api/drafts/${encodeURIComponent(state.pendingId)}`)
      : await draftRequest('/api/drafts', draftJSON('POST', {vacancy, resume: state.resume}));
    state.pendingId = result.id;
    let failures = 0;
    while (result.status === 'running') {
      await new Promise(resolve => setTimeout(resolve, 1500));
      try { result = await draftRequest(`/api/drafts/${encodeURIComponent(result.id)}`); failures = 0; }
      catch (error) { if (++failures >= 3) throw error; }
    }
    state.pendingId = null;
    if (result.status !== 'succeeded') {
      if (result.text) { state.id = result.id; state.text = result.text; state.path = result.path; state.dirty = true; }
      throw new Error(result.error || 'Не удалось создать письмо. Попробуйте ещё раз.');
    }
    state.id = result.id; state.text = result.text; state.path = result.path; state.dirty = false;
    state.message = 'Черновик готов. Можно отредактировать и скопировать.';
  } catch (error) {
    if (error.status === 404 && state.pendingId) { state.pendingId = null; state.message = 'Задание больше недоступно, возможно сервер перезапущен. Проверьте applications перед повторной генерацией.'; }
    else state.message = error.message;
  }
  finally { state.busy = false; updateLetter(job); }
}

window.addEventListener('beforeunload', event => {
  if ([...letters.values()].some(state => state.dirty)) { event.preventDefault(); event.returnValue = ''; }
});
