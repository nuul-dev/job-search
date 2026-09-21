---
name: application
description: >
  Skill for writing personal job application responses (отклики) and replies
  to recruiters. Output is plain-text chat messages for Telegram, hh.ru, or
  email, that sound like a human wrote them — not AI. Mirrors the recruiter's
  tone exactly (casual / friendly / business / direct).
  Trigger when user asks to write an отклик, сопроводительное, сопровод,
  ответ рекрутёру, response to a vacancy, or a chat message to an HR.
---

# Job Application Response — Write Like a Human

## The one rule that matters most

Mirror the recruiter's tone. Match:
- Greeting style ("Привет!" vs "Добрый день" vs no greeting)
- Punctuation (smileys, parentheses, exclamation marks)
- Formality (Вы vs ты)
- Length (don't write 5 paragraphs to their 3-line message)
- Energy (chill / urgent / direct / formal)

If they wrote "Привет)" — your response opens "Привет!" or "Привет)".
If they wrote "Здравствуйте, рассмотрите нашу вакансию" — you write "Здравствуйте."

## Forbidden characters in chat output

These instantly mark text as AI-generated. Strip all of them before sending.

| Char | What to use instead |
|------|---------------------|
| `—` (em dash) | comma, period, or rewrite the sentence |
| `→` `->` `=>` (arrows) | words: "сначала, потом, дальше" or just commas |
| `**bold**` / `*italic*` | nothing — messengers show asterisks literally |
| `# ## ###` headers | nothing — paragraphs separate ideas |
| Bullet lists `- item` | only if recruiter used them; else flow prose |
| `;` semicolons in casual chat | period or comma |

## Anti-AI writing tells

Rewrite if you catch any of these:

- "В рамках..." / "В связи с..." / "Что касается..."
- "Я могу гордиться тем, что..." / "Хотел бы отметить..."
- Three-noun lists: "опыт, навыки, экспертизу"
- "Не просто X, а Y" formula
- "Не X, а Y" used to contrast (e.g., "не задача, а вызов")
- Perfect parallel structure: "Я делал X. Я делал Y. Я делал Z."
- Listing everything in groups of three
- "Хочется отметить" / "Стоит подчеркнуть"
- Generic adjectives: "сильный", "глубокий", "обширный опыт"

## What human writing looks like

- Mix short and long sentences. Some sentences are 3 words. Others run on a bit.
- Active "я", not passive.
- Concrete numbers and project names, not abstractions.
- Natural connectives: "ещё", "кстати", "тут", "вот", "по сути", "плюс".
- Small acknowledgement: "спасибо что написали", "интересная вакансия".
- One small slip is okay (chat message, not legal contract).
- It's normal to end on a question or a short statement, not a closing formula.

## Tone detection — what to mirror

| Their signal | Your response register |
|--------------|------------------------|
| "Привет!" + смайлики / скобочки )) | Casual: "Привет!", "ты", смайлики ок |
| "Привет, расскажите..." (ты, no smileys) | Friendly direct: "Привет", "ты" |
| "Добрый день. Подскажите..." | Professional warm: "Добрый день", "Вы" |
| "Здравствуйте, рассмотрите..." | Formal: "Здравствуйте", "Вы", restrained |
| No greeting, just questions | No greeting, just answers |
| Anti-corporate vacancy ("без бюрократии", "хватит ТЗ") | Skip pleasantries, lead with substance |
| Vacancy in English | Reply in English, same register match |

## Length matching

**Cold отклик (первое сообщение на вакансию, без предыдущей переписки): 3-4 предложения максимум, считая прощание.**

Никаких абзацев с разбивкой по проектам и стеку. Структура:
1. Greeting ("Привет!" / "Добрый день!").
2. 1-2 содержательных предложения с самым сильным match-аргументом и одним числом/проектом.
3. Одна короткая строка с готовностью/контактом ("Готов выйти на этой неделе, github.com/...").
4. Короткое человеческое прощание ("Спасибо за уделенное время, хорошего дня!").

**Telegram specifically:** в TG рекрутёр не знает кто ты по нику. Упомяни имя — либо в приветствии ("Привет, меня зовут Илья"), либо в строке с контактом ("Илья, github.com/nuul-dev").

Если хочется добавить ещё детали — не надо, лучше оставить рекрутёру повод задать вопрос.

Для follow-up ответов в `## Переписка` (рекрутёр уже задал конкретный вопрос) — матчим длину под вопрос:
- 1-line ping → 2-4 line answer.
- 3-question screening → 1 short paragraph per question.
- Long detailed pitch → можно матчить глубину, но не превышать их длину более чем в 2x.

## Structure inside the response

- They numbered questions → mirror the numbering, no extra structure on top.
- They asked one open question → flowing paragraph, not bullets.
- They sent bullet points → bullets ok in response.
- Never add headers, never add bold.

## What to include

1. Direct answers to every question. No dodging.
2. One concrete proof: a number, a specific project, a measurable result.
3. Honest line on availability if they ask.
4. End with a short human goodbye ("Спасибо за уделенное время, хорошего дня!" / "Спасибо, что прочитали!"). Don't use corporate filler like "буду рад любой обратной связи" or "I look forward to hearing from you".

## What NOT to include

- Self-praising adjectives ("сильный разработчик", "глубокий опыт").
- Verbatim copies from the resume.
- Apologies for gaps. If they ask about something you don't have, say what's adjacent that you do have.
- Volunteering gaps unprompted. Never mention missing skills or experience in a cold отклик or cover letter — this is for the interview. Only address a gap if the recruiter explicitly asks about it.
- Promises that overstate ("точно справлюсь", "решу любую задачу").

## Workflow

1. Read the vacancy + recruiter message together.
2. Pick the tone register from the cheatsheet above.
3. List what they actually asked (explicit + implicit).
4. Draft a response answering each, with one proof point.
5. **Cleanup pass:** search the draft for `—`, `→`, `**`, `;` in casual contexts, `#`, "не X, а Y", "в рамках", "стоит отметить". Rewrite each.
6. **Length pass:** trim until you're within ~2x their message length.
7. **Read-aloud test:** would a friend text this to another friend in chat? If it sounds like a press release, rewrite.
8. **Archive step (mandatory):** save the response to `applications/`. Don't ask permission, just do it as part of the workflow. Then output the response text to the user.

## Archive format

Filename: `applications/YYYY-MM-DD-{company-slug}-{role-slug}.md`
- Date prefix sorts entries chronologically.
- `{company-slug}` and `{role-slug}` are kebab-case, lowercase, transliterated to Latin if needed.
- If the company is hidden, use `unknown-company` or a project codename.

File contents (Markdown):

```markdown
# {Vacancy title} — {Company}

- **URL:** {full link to vacancy}
- **Источник:** {hh.ru / Habr / Hirify / Telegram / direct}
- **Дата отклика:** YYYY-MM-DD
- **Статус:** отправлен
- **ЗП:** {salary if known, else "не указана"}
- **Формат:** {remote / hybrid / office, employment type}

## Отклик

{the full response text, exactly as sent}

## Заметки

_(пусто; здесь можно добавить детали собеса, ответ рекрутёра и т.д.)_
```

The status field starts as "отправлен" and the user can update it later (ответили / собес / отказ / оффер). Don't overwrite the user's edits to existing files; if a file already exists for the same vacancy, ask before saving over it.

## Follow-up replies (recruiter sent another question after the first отклик)

When the user asks to write a follow-up reply for a vacancy that already has a file in `applications/`, **append** to the existing file rather than creating a new one.

Add a `## Переписка` section just before `## Заметки` (or append to it if it already exists). Each exchange uses this format:

```markdown
### YYYY-MM-DD — рекрутёр

> {their message, verbatim}

### YYYY-MM-DD — мой ответ

{your reply}
```

Order entries chronologically (oldest first). Don't repeat the original `## Отклик` content; that section stays as the first message. Also update the `Статус` field at the top if the situation changed (e.g., `ответили`, `собес`, `тестовое`, `отказ`, `оффер`).

## Resources

See `examples/responses.md` for paired tone-matching examples covering casual, direct, formal, screening, and cold-outreach cases.
