# application

Skill for writing job application responses (отклики) and recruiter chat replies that sound human, not AI.

## What it does

Produces plain-text chat messages tuned to the recruiter's tone. If they wrote "Привет)" you get back "Привет!". If they wrote "Здравствуйте, рассмотрите нашу вакансию" you get formal restraint. No em dashes, no arrows, no markdown formatting, no AI tells.

## Why it exists

LLMs default to writing structured "essay" responses: em dashes for emphasis, arrows between steps, bold headers for each numbered point. Recruiters see this and immediately know it's AI. This skill enforces the opposite style: plain prose, tone-matched, ready to paste into Telegram or hh.ru chat.

## How to invoke

```
/application                 # generic — paste vacancy + recruiter message
напиши отклик на вакансию    # natural trigger
ответь рекрутёру             # natural trigger
```

Then provide:
1. The vacancy text (or link)
2. The recruiter's message if there was one
3. Anything specific you want highlighted

## Key rules enforced

- No `—` (em dash), `→` (arrow), `**bold**`, headers, semicolons in casual chat
- Tone register matched to recruiter (casual / friendly / business / direct)
- Response length ≤ 2x recruiter's message length
- One concrete proof point (number, project, result)
- Direct answers to every asked question
- Ends naturally, no "буду рад обратной связи"

## See also

- [`SKILL.md`](./SKILL.md) — full ruleset
- [`examples/responses.md`](./examples/responses.md) — paired tone-matching examples
- [`../resume/SKILL.md`](../resume/SKILL.md) — for resume/cover letter writing
