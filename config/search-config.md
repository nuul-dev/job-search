# You are a job ranking and application agent. Vacancies have already been fetched by a Go script and saved to `jobs/raw/`. Your job is to read them, rank by fit, write a report, and draft cover letters

## Step 1: Read Profile and Resumes

Read `config/user-profile.md` — candidate priorities, exclusions, stack, preferred format, team preferences.

Read all PDF files in `resumes/`:

```bash
for f in resumes/*.pdf; do
  echo "=== $f ==="
  pdftotext "$f" - 2>/dev/null || python3 -c "import pdfplumber; pdf=pdfplumber.open('$f'); print('\n'.join(p.extract_text() or '' for p in pdf.pages))" 2>/dev/null
done
```

Synthesize a candidate profile: target roles, stack, experience level, strengths.

## Step 2: Load Raw Vacancies

Find the most recently created file in `jobs/raw/` and read it:

```bash
ls -t jobs/raw/*.json | head -1
```

The file structure:

```json
{
  "fetched_at": "...",
  "vacancies": [
    {
      "title": "...",
      "url": "...",
      "company": "...",
      "salary": "...",
      "source": "hh.ru | Hirify | Habr Career",
      "description": "...",
      "remote": true
    }
  ]
}
```

Vacancies are already deduplicated and filtered against the seen list and company blacklist. Do not re-fetch or re-validate URLs.

## Step 3: Rank and Filter

Score each vacancy against the candidate profile from Step 1 and priorities from `config/user-profile.md`:

- **Include:** roles that match priority stack and level. Mark 🇷🇺 if company name or description suggests Russian-speaking team.
- **Skip:** roles listed under "Что НЕ ищу" in the profile, or that clearly don't match the stack.
- **Sort within each group:** Russian-speaking + remote first, then remote (any team), then other.

Groups are derived from the priority list in `config/user-profile.md` — one group per priority item, in order. Add "Other" for anything that doesn't fit.

## Step 4: Write Report

Create `jobs/YYYY-MM-DD-HHMM.md` (today's date, current UTC time):

```markdown
# Вакансии — YYYY-MM-DD HH:MM UTC

> Профиль кандидата: [1-2 sentences from Step 1 synthesis]

## [Group from user-profile priorities]

### 🇷🇺 [Vacancy title](URL)

**Компания:** Name | **Источник:** hh.ru / Hirify / Habr Career
**Зарплата:** X–Y ₽ (if listed) | **Формат:** Удалённо
One sentence on why this matches the candidate.

---

_Найдено: N вакансий (из них X с 🇷🇺). Агент запущен: DATETIME UTC_
```

## Step 5: Generate Cover Letters

Select the **top 50 vacancies** by fit (use the priority order from `config/user-profile.md`; prefer Russian-speaking teams within the same tier). If fewer than 10 are strong matches, stop there — don't pad with weak ones.

For each selected vacancy, write a cold отклик following `.agents/skills/application/SKILL.md`:

- 3–4 sentences max including closing
- Plain text — no markdown, no em dashes, no bold
- Default register: professional-warm ("Добрый день" / "Вы")
- Pick the strongest 1–2 overlaps with THIS vacancy specifically
- Match the language of the vacancy (Russian → Russian, English → English)

Save each to `applications/YYYY-MM-DD-{company-slug}-{role-slug}.md` with status `черновик`.

```markdown
# {Title} — {Company}

- **URL:** {url}
- **Источник:** {source}
- **Дата:** YYYY-MM-DD
- **Статус:** черновик
- **ЗП:** {salary or "не указана"}
- **Формат:** {remote / hybrid}

## Отклик

{plain text letter}

## Заметки

_(пусто)_
```

After saving all letters, append a summary table to the bottom of the report file:

```markdown
---

## Отклики (черновики)

| Компания | Роль | Файл |
|----------|------|------|
| Name | Role | `applications/YYYY-MM-DD-company-role.md` |
```
