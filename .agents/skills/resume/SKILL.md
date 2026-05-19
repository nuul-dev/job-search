---
name: resume
description: >
  Expert resume writing skill for software engineers. Produces ATS-friendly,
  impact-driven resumes using the Google XYZ formula. Tailors content to specific
  vacancies, rewrites weak bullet points, identifies gaps, and writes cover letters.
  Trigger when user asks to write, rewrite, improve, or tailor a resume or CV,
  or when asked to write a cover letter. Also triggers on /resume.
---

# Resume Writing — Rules & Method

## Core Formula (Google XYZ)

Every bullet point follows: **Accomplished [X] as measured by [Y] by doing [Z]**

- **X** = result or impact, starts with strong action verb
- **Y** = measurement: %, users, $, ms, requests/sec, hours saved
- **Z** = method: specific tools, architecture decision, technique

**Bad:** "Worked on backend services and improved performance"
**Good:** "Reduced API latency by 71% by introducing Redis caching layer, serving 2M+ daily requests"

**Bad:** "Developed microservices architecture"
**Good:** "Designed 10-service Go microservice system (Clean Architecture + gRPC) cutting deployment cycle from 2 days to 40 min via automated CI/CD pipelines"

When the user has no exact metrics, estimate logically and mark as approximate (~). Never leave a bullet without some form of scale indicator.

## Bullet Point Rules

- Start with past-tense action verb (Designed / Built / Reduced / Led / Migrated / Automated / Optimized / Implemented / Scaled)
- 1–2 lines max per bullet
- 3–5 bullets per role (recent roles get more, older roles get 2–3)
- Show impact, NOT responsibilities — "responsible for X" is always wrong
- Each bullet should answer: "so what?" and "how big?"
- Include concrete tech stack in the bullet itself, not only in skills section

## Resume Structure

```
[Name] — [Target Role]
[email] | [phone] | [city] | [GitHub] | [LinkedIn/Telegram]

ОПЫТ РАБОТЫ

[Company] — [Role]                              [Date range]
[Stack: lang, framework, DB, infra]
— Bullet (XYZ)
— Bullet (XYZ)
— Bullet (XYZ)

НАВЫКИ
[Backend]: ...
[Databases]: ...
[AI/LLM]: ...
[Infrastructure]: ...
[Blockchain]: ... (if relevant)

ОБРАЗОВАНИЕ
[Degree, University, Years]
```

No objective/summary section unless specifically requested. No "about me" fluff.

## Skills Section Rules

- Group by category (Backend / Databases / AI / Infra / Blockchain)
- 8–12 items total per category
- List only what you can discuss in depth at interview
- Order within category: strongest/most relevant first
- No rating bars, no "beginner/intermediate" labels — just list

## Length

- **< 5 years experience:** 1 page, strict
- **5–10 years:** 1–2 pages
- **Never exceed 2 pages**

## Tailoring to Vacancy

When rewriting for a specific vacancy:
1. Extract top 5–7 required skills from the job description
2. For each skill — find matching experience in the resume, reframe bullet to highlight it
3. Mirror exact terminology from the vacancy (if they say "event-driven" use that phrase)
4. Move most-relevant bullets to top of each role
5. Drop or shorten bullets irrelevant to this role

## What to Cut

- Soft skill claims without proof ("team player", "fast learner")
- Duties that everyone in the role does by default
- Tech stack items listed twice (in bullet AND skills section is fine, but not 3x)
- Education details for anyone with 3+ years experience (just degree + school + year)
- "References available upon request"
- Objective statements / career summaries (unless specifically requested)

## Cover Letter Structure (3–4 paragraphs)

1. **Hook** — why this specific company/product, not generic opener
2. **Match** — 2–3 concrete intersections between your experience and their requirements, with brief proof
3. **Value add** — one specific thing you'd bring that's hard to get elsewhere
4. **Close** — short, confident, no "I hope to hear from you"

Cover letters in Russian by default. Switch to English only if the vacancy is in English.
Max 250 words. No filler. No "I am writing to express my interest in..."

## Language (for Russian-market resumes)

- Bullet points in Russian
- Tech terms stay in English (Go, gRPC, PostgreSQL, CI/CD, etc.)
- Numbers in Arabic numerals (not written out)
- Action verbs: Спроектировал / Разработал / Сократил / Внедрил / Оптимизировал / Автоматизировал / Масштабировал / Мигрировал

## Red Flags to Fix Immediately

- Any bullet starting with "Участвовал в" / "Помогал с" / "Был ответственен за" → rewrite to show direct ownership
- Bullet with no number → add scale estimate
- Role description without tech stack → add stack
- Skills section with 20+ items → trim to most relevant
- Same bullet appears in two different roles → delete one

## Skill Resources (read on demand)

Don't read these eagerly. Read only when you actually start the corresponding task:

- `examples/bullets.md` — concrete before/after bullet rewrites in EN and RU. Read this when rewriting bullets and you need pattern inspiration.
- `templates/two-column.html` — production-tested two-column A4 resume template (WeasyPrint-compatible). Read this when actually building an HTML/PDF resume. Has placeholder markers `REPLACE-...` to fill in.
- `scripts/build-pdf.py` — converts the HTML to PDF via WeasyPrint. Usage: `python3 .agents/skills/resume/scripts/build-pdf.py <input.html> <output.pdf>`.

## Workflow when building a polished PDF resume

1. Gather user data (existing resume, vacancy if tailoring, metrics they remember).
2. Apply XYZ formula and red-flag pass to all bullets — refer to `examples/bullets.md` for pattern inspiration if stuck.
3. Read `templates/two-column.html`, copy it to the working file (e.g., `resumes/<name>-<target>.html`).
4. Fill in all `REPLACE-...` placeholders. Add/duplicate sections as needed.
5. Run `python3 .agents/skills/resume/scripts/build-pdf.py <html> <pdf>` to render.
6. Open the PDF, check: nothing clipped on the right edge, length ≤ 2 pages, AI/key section appears above the fold.

## Workflow when only critiquing (no PDF needed)

1. Read the user's resume.
2. For each bullet — XYZ check: identify missing X / Y / Z component.
3. Output a table or per-bullet list: original → rewritten suggestion + the reason.
4. End with a short list of the 3 most impactful changes ranked by ROI.
