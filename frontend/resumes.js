const resumeState = {files: [], current: null, dirty: false, busy: false, preview: false, preferred: ''};
try { resumeState.preferred = localStorage.getItem('job-inbox-preferred-resume') || ''; } catch {}
function showVacancyWorkspace() {
  $('resume-workspace').hidden = true; $('vacancy-workspace').hidden = false;
  $('resumes-nav').classList.remove('active'); document.title = 'Поиск работы · Вакансии';
}
function resumeFeedback(message) { $('resume-feedback').textContent = message; }
function resumeBusy(busy) {
  resumeState.busy = busy;
  for (const id of ['resume-new','resume-upload','resume-refresh']) $(id).disabled = busy;
}
function canChangeResume() { return !resumeState.busy && (!resumeState.dirty || window.confirm('Есть несохранённые изменения. Открыть другое резюме без их сохранения?')); }
async function loadResumes() {
  try {
    resumeState.files = (await draftRequest('/api/resumes')).resumes;
    availableResumes = null;
    renderResumeList();
  } catch (error) { resumeFeedback(error.message); }
}
function renderResumeList() {
  const list = $('resume-list'); list.replaceChildren();
  if (!resumeState.files.length) list.append(el('p','empty','Пока нет резюме. Загрузите файл или начните с чистого листа.'));
  for (const name of resumeState.files) {
    const row = el('button','resume-item' + (resumeState.current?.name === name ? ' selected' : ''));
    row.disabled = resumeState.busy; row.setAttribute('aria-pressed', String(resumeState.current?.name === name));
    row.append(el('span','resume-file-kind',name.toLowerCase().endsWith('.pdf') ? 'PDF' : 'Текст'),el('strong','',name.replace(/\.(md|pdf)$/i,'')));
    row.append(el('span','hint',name === resumeState.preferred ? 'Для откликов по умолчанию' : name));
    row.onclick = () => openResume(name); list.append(row);
  }
}
async function openResume(name, confirmed = false) {
  if (resumeState.busy || (!confirmed && !canChangeResume())) return;
  resumeBusy(true); $('resume-document').inert=true; renderResumeList(); resumeFeedback('Открываем резюме…');
  try {
    resumeState.current = await draftRequest(`/api/resumes/${encodeURIComponent(name)}`);
    resumeState.dirty = false; resumeState.preview = !resumeState.current.editable;
    resumeFeedback(resumeState.current.editable ? 'Текст можно редактировать и сохранять.' : 'PDF сохранён как оригинал. Для правок создайте текстовую копию.');
  } catch (error) {
    resumeFeedback(error.message + ' ');
    const download = el('a', '', 'Скачать оригинал'); download.href = `/api/resumes/${encodeURIComponent(name)}/download`; $('resume-feedback').append(download);
  }
  finally { resumeBusy(false); $('resume-document').inert=false; renderResumeList(); renderResumeDocument(); }
}
function createResume(copy = false) {
  if (resumeState.busy || (!copy && !canChangeResume())) return;
  const current = resumeState.current;
  resumeState.current = {name:'',new:true,editable:true,text:copy ? current.text : '',version:'', suggested:copy ? current.name.replace(/\.[^.]+$/,'')+'-копия' : ''};
  resumeState.dirty = !!resumeState.current.text; resumeState.preview = false;
  resumeFeedback(copy ? 'Текстовая копия. Проверьте переносы и порядок текста после извлечения из PDF.' : 'Добавьте реальные факты о себе: опыт, проекты, навыки и контакты.');
  renderResumeList(); renderResumeDocument(); $('resume-name')?.focus();
}
function resumePreview(text) {
  const article = el('article','resume-preview');
  if (!text.trim()) { article.append(el('p','hint','Здесь появится текст вашего резюме.')); return article; }
  for (const line of text.split('\n')) {
    const heading = /^(#{1,3})\s+(.+)$/.exec(line);
    if (heading) article.append(el(`h${Math.min(heading[1].length+1,4)}`,'',heading[2]));
    else if (line.trim()) article.append(el('p','',line));
  }
  return article;
}
function renderResumeDocument() {
  const panel = $('resume-document'), current = resumeState.current;
  if (!current) return;
  panel.replaceChildren();
  const top = el('div','resume-document-top');
  const heading = el('div'); heading.append(el('p','hint',current.new ? 'Новая версия' : current.editable ? 'Редактируемое резюме' : 'Оригинал PDF'),el('h2','',current.new ? 'С чистого листа' : current.name.replace(/\.(md|pdf)$/i,'')));
  top.append(heading); panel.append(top);
  if (current.new) {
    const label = el('label','option-field','Название версии'); label.htmlFor = 'resume-name';
    const input = el('input'); input.id = 'resume-name'; input.value = current.suggested || ''; input.placeholder = 'Например: Backend Go — Middle'; input.maxLength = 100;
    input.oninput = () => { current.suggested = input.value; resumeState.dirty = true; };
    label.append(input); panel.append(label);
  }
  const tools = el('div','resume-tools');
  if (current.editable) {
    const edit = el('button',resumeState.preview ? 'secondary' : 'primary','Редактировать');
    const preview = el('button',resumeState.preview ? 'primary' : 'secondary','Предпросмотр');
    edit.setAttribute('aria-pressed',String(!resumeState.preview)); preview.setAttribute('aria-pressed',String(resumeState.preview));
    edit.onclick = () => { resumeState.preview=false; renderResumeDocument(); }; preview.onclick = () => { resumeState.preview=true; renderResumeDocument(); };
    tools.append(edit,preview);
  } else {
    const copy = el('button','primary','Создать редактируемую копию'); copy.onclick=()=>createResume(true); tools.append(copy);
  }
  if (!current.new) {
    if (current.editable) { const copy = el('button','secondary','Создать копию'); copy.onclick=()=>createResume(true); tools.append(copy); }
    const download = el('a','secondary','Скачать'); download.href=`/api/resumes/${encodeURIComponent(current.name)}/download`; tools.append(download);
    const prefer = el('button','secondary',resumeState.preferred === current.name ? 'Выбрано для откликов' : 'Использовать для откликов');
    prefer.setAttribute('aria-pressed',String(resumeState.preferred===current.name));
    prefer.onclick=()=>{try {localStorage.setItem('job-inbox-preferred-resume',current.name);resumeState.preferred=current.name;resumeFeedback('Это резюме будет выбрано по умолчанию для новых откликов.');renderResumeList();renderResumeDocument();}catch{resumeFeedback('Браузер не сохранил выбор. Проверьте доступ к локальному хранилищу.');}};
    tools.append(prefer);
  }
  panel.append(tools);
  if (resumeState.preview || !current.editable) panel.append(resumePreview(current.text));
  else {
    const label=el('label','visually-hidden','Текст резюме');label.htmlFor='resume-content';
    const input=el('textarea','resume-content');input.id='resume-content';input.value=current.text;input.placeholder='Имя и целевая роль\n\nОпыт работы\nКомпания, период, задачи и результаты\n\nПроекты\n\nНавыки\n\nОбразование и контакты';input.spellcheck=true;
    input.oninput=()=>{current.text=input.value;resumeState.dirty=true;const status=$('resume-save-status');if(status)status.textContent='Есть несохранённые изменения';};panel.append(label,input);
  }
  const footer=el('div','resume-footer');const status=el('span','hint',resumeState.dirty?'Есть несохранённые изменения':current.new?'Заполните резюме и сохраните':'Все изменения сохранены');status.id='resume-save-status';footer.append(status);
  if(current.editable){const save=el('button','primary','Сохранить резюме');save.onclick=()=>saveResume();footer.append(save);}
  panel.append(footer);
}
async function saveResume() {
  if(resumeState.busy)return;
  const current=resumeState.current;
  if(!current.text.trim()){resumeFeedback('Добавьте текст резюме перед сохранением.');return;}
  let name=current.name;
  if(current.new){name=(current.suggested||'').trim();if(!name){resumeFeedback('Укажите название версии.');$('resume-name')?.focus();return;}if(!name.toLowerCase().endsWith('.md'))name+='.md';}
  resumeBusy(true);$('resume-document').inert=true;resumeFeedback('Сохраняем резюме…');
  try{
    const result=await draftRequest(current.new?'/api/resumes':`/api/resumes/${encodeURIComponent(name)}`,draftJSON(current.new?'POST':'PUT',current.new?{name,text:current.text}:{text:current.text,version:current.version}));
    resumeState.current=result;resumeState.dirty=false;await loadResumes();renderResumeDocument();resumeFeedback('Резюме сохранено. Оно доступно для создания сопроводительных писем.');
  }catch(error){resumeFeedback(error.status===409?'Версия с таким именем уже есть или файл изменился. Ваш текст остался в редакторе: сохраните его отдельной копией.':error.message);}
  finally{resumeBusy(false);$('resume-document').inert=false;renderResumeList();}
}
$('resumes-nav').onclick=()=>{$('vacancy-workspace').hidden=true;$('resume-workspace').hidden=false;document.querySelectorAll('[data-view]').forEach(button=>button.classList.remove('active'));$('resumes-nav').classList.add('active');document.title='Поиск работы · Резюме';loadResumes();};
$('resume-new').onclick=()=>createResume();$('resume-refresh').onclick=loadResumes;
$('resume-upload').onclick=()=>{if(canChangeResume())$('resume-file').click();};
$('resume-file').onchange=async()=>{
  const file=$('resume-file').files[0];if(!file)return;
  if(!/\.(pdf|md)$/i.test(file.name)||file.size>8*1024*1024){resumeFeedback('Выберите PDF до 8 МБ или Markdown до 100 КБ.');$('resume-file').value='';return;}
  resumeBusy(true);$('resume-document').inert=true;renderResumeList();resumeFeedback('Загружаем файл…');
  try{const form=new FormData();form.append('file',file);const result=await draftRequest('/api/resumes/upload',{method:'POST',body:form});await loadResumes();resumeBusy(false);await openResume(result.name,true);}
  catch(error){resumeFeedback(error.message);}
  finally{resumeBusy(false);$('resume-document').inert=false;renderResumeList();$('resume-file').value='';}
};
window.addEventListener('beforeunload',event=>{if(resumeState.dirty){event.preventDefault();event.returnValue='';}});
