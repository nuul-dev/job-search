---
name: application
description: >
  Skill for writing personal job application responses (отклики) and replies
  to recruiters. Output is plain-text chat messages for Telegram, hh.ru, or
  email, that sound like a human wrote them — not AI. Mirrors the recruiter's
  tone exactly (casual / friendly / business / direct).
  Trigger when user asks to write an отклик, ответ рекрутёру, response to a
  vacancy, or a chat message to an HR.
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

Roughly match their message length. Rules:
- 1-line ping → 2-4 line answer.
- 3-question screening → 1 short paragraph per question.
- Long detailed pitch → you can match the depth, but never exceed their length 2x.

## Structure inside the response

- They numbered questions → mirror the numbering, no extra structure on top.
- They asked one open question → flowing paragraph, not bullets.
- They sent bullet points → bullets ok in response.
- Never add headers, never add bold.

## What to include

1. Direct answers to every question. No dodging.
2. One concrete proof: a number, a specific project, a measurable result.
3. Honest line on availability if they ask.
4. End naturally. No "буду рад любой обратной связи" or "I look forward to hearing from you".

## What NOT to include

- Self-praising adjectives ("сильный разработчик", "глубокий опыт").
- Verbatim copies from the resume.
- Apologies for gaps. If they ask about something you don't have, say what's adjacent that you do have.
- Promises that overstate ("точно справлюсь", "решу любую задачу").

## Workflow

1. Read the vacancy + recruiter message together.
2. Pick the tone register from the cheatsheet above.
3. List what they actually asked (explicit + implicit).
4. Draft a response answering each, with one proof point.
5. **Cleanup pass:** search the draft for `—`, `→`, `**`, `;` in casual contexts, `#`, "не X, а Y", "в рамках", "стоит отметить". Rewrite each.
6. **Length pass:** trim until you're within ~2x their message length.
7. **Read-aloud test:** would a friend text this to another friend in chat? If it sounds like a press release, rewrite.

Output the final response as plain text, ready to paste into a chat. No surrounding commentary unless the user asks.

## Resources

See `examples/responses.md` for paired tone-matching examples covering casual, direct, formal, screening, and cold-outreach cases.
