# resume

Expert resume writing for software engineers. Impact-driven bullets, ATS-friendly structure, vacancy tailoring.

## What it does

Applies the **Google XYZ formula** to every bullet point: *Accomplished [X] as measured by [Y] by doing [Z]*. Turns vague responsibility lists into quantified achievements. Tailors resume content to specific vacancies by mirroring their terminology and surfacing the most relevant experience.

Also writes cover letters: short, specific, no fluff.

## How to invoke

```
/resume                         # review current resume and suggest improvements
/resume tailor                  # rewrite resume for a specific vacancy (paste vacancy text)
/resume cover-letter            # write a cover letter for a vacancy
/resume fix <bullet>            # rewrite a single weak bullet point
```

## Principles

| Rule | Why |
|------|-----|
| XYZ formula on every bullet | Proves impact, not just presence |
| 3–5 bullets per role | Enough depth, no padding |
| Metrics always | "Improved performance" is worthless without a number |
| 1 page under 5 years | Recruiters spend 6–10 seconds on first scan |
| Mirror vacancy language | ATS keyword matching + shows you read the JD |

## Example transformation

**Before:**
> Разрабатывал микросервисы на Go, работал с PostgreSQL и Redis

**After:**
> Спроектировал 10 Go-микросервисов (Clean Architecture, gRPC) с асинхронным обменом через RabbitMQ; сократил время деплоя с 2 дней до 40 мин через единые CI/CD-шаблоны

## See also

- [`SKILL.md`](./SKILL.md) — full LLM-facing instructions
