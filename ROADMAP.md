# Job Search — Roadmap

## О проекте

Персональный AI-ассистент для поиска работы. Фетчер на Go вытягивает свежие вакансии с hh.ru, Hirify и Habr Career, ранкер на Claude API скорит их по профилю кандидата, отдельный шаг генерирует черновики откликов через тот же Claude по skill-промту.

Сейчас всё работает на файловой системе: Markdown-отчёты в `jobs/`, отклики в `applications/`, профиль в `config/user-profile.md`. Этот документ описывает миграцию на полноценный сервис: PostgreSQL как single source of truth, Go-бекенд с HTTP API, React-фронт для inbox/kanban.

## Стек

- **Backend:** Go (chi, pgx, slog), Anthropic SDK
- **DB:** PostgreSQL 16 (jsonb для сырых ответов API, плоские колонки для всего что джойнится)
- **Infra:** docker-compose, goose для миграций
- **Frontend (план):** Vite + React + TS + Tailwind + TanStack Query

## Архитектурные решения

- **Auth model:** single-user сейчас, multi-user-ready схема (user_id во всех таблицах, basic auth позже)
- **DB layer:** pgx напрямую, сырые SQL-запросы, без ORM
- **Source of truth:** PostgreSQL. Markdown — только экспорт по запросу
- **Local env:** docker-compose (postgres + adminer)

---

## Этап 1: Инфра

- [x] `docker-compose.yml` в корне: postgres:16 + adminer (порты 5432, 8080)
- [x] Миграционный инструмент: **goose** (один бинарь, SQL-миграции, нет магии)
- [x] `db/migrations/` — папка под версионируемые миграции
- [x] `.env.example` с `DATABASE_URL`, `CLAUDE_API_KEY`, `HH_USER_EMAIL`
- [x] `Makefile` или `scripts/dev.sh`: `db-up`, `db-down`, `db-reset`, `migrate-up`, `migrate-status`

## Этап 2: Схема БД

- [ ] `users` (id, email, created_at) — single row пока, чтобы FK работали
- [ ] `profiles` (user_id PK, priorities jsonb, search_queries text[], stack text[], location text, work_format jsonb, blacklist_companies text[], exclusions jsonb, updated_at)
- [ ] `vacancies` (id, source, source_external_id, url, title, company, salary, location, work_format, description, raw jsonb, fetched_at, published_at, UNIQUE(source, source_external_id))
- [ ] `vacancy_scores` (vacancy_id, user_id, score, group_name, notes, ranked_at) — отдельно, чтобы ранкер мог пересчитывать без потери вакансии
- [ ] `applications` (id, vacancy_id, user_id, status enum [draft, sent, replied, screen, test, offer, rejected, archived], draft_body, sent_at, created_at, updated_at)
- [ ] `application_messages` (id, application_id, author enum [me, recruiter], body, sent_at) — переписка
- [ ] `fetch_runs` (id, started_at, finished_at, source, fetched_count, error) — заменит логи
- [ ] Индексы: `vacancies(fetched_at DESC)`, `vacancies(source, source_external_id)`, `vacancy_scores(user_id, score DESC)`, `applications(user_id, status)`

## Этап 3: Бекенд-скелет

- [ ] Новый Go-модуль в `backend/`: `github.com/nuul/job-search/backend`
- [ ] Структура: `cmd/api/main.go`, `internal/db`, `internal/handlers`, `internal/config`
- [ ] `pgxpool.Pool` через DATABASE_URL, graceful shutdown
- [ ] Роутер `chi`
- [ ] `GET /health` — пинг + проверка DB
- [ ] `slog` для логов в stdout, JSON-формат
- [ ] Конфиг через `env` (godotenv для локалки)

## Этап 4: Миграция данных

Одноразовый скрипт `cmd/migrate-md/main.go`.

- [ ] Парсер `config/user-profile.md` → `profiles`
- [ ] Парсер `jobs/raw/*.json` → `vacancies`
- [ ] Парсер `jobs/*.md` — извлечь только score/group/notes → `vacancy_scores`
- [ ] Парсер `applications/*.md` — фронт-блок + `## Отклик` + `## Переписка` → `applications` + `application_messages`
- [ ] Парсер `seen-vacancies.txt` — игнорировать (UNIQUE-constraint его заменит)
- [ ] Идемпотентность: `ON CONFLICT DO NOTHING` везде
- [ ] После миграции `jobs/` и `applications/` оставить, но `run-search.sh` больше в них не пишет

## Этап 5: Фетчер в БД

- [ ] Перенести `backend/fetch/` → `backend/internal/fetcher/`
- [ ] Сменить выходной слой: вместо `jobs/raw/*.json` — `INSERT INTO vacancies ... ON CONFLICT (source, source_external_id) DO UPDATE SET ...`
- [ ] Каждый запуск — строка в `fetch_runs`
- [ ] Команда: `cmd/fetch/main.go` (standalone бинарь для cron)
- [ ] Удалить `seen-vacancies.txt`

## Этап 6: HTTP API — read-only

- [ ] `GET /vacancies?status=&group=&source=&since=&limit=&offset=` — отчёт-замена
- [ ] `GET /vacancies/:id` — детали + связанная application
- [ ] `GET /applications?status=` — kanban-данные
- [ ] `GET /applications/:id` — отклик + вся переписка
- [ ] `GET /profile`
- [ ] Тесты: `httptest` + `testcontainers/postgres` (или mock pgx)

## Этап 7: HTTP API — write

- [ ] `POST /applications` (vacancy_id, draft_body) — создать черновик
- [ ] `PATCH /applications/:id` — сменить status, обновить body
- [ ] `POST /applications/:id/messages` — добавить запись в переписку
- [ ] `PATCH /profile` — обновить приоритеты/запросы
- [ ] `POST /fetch/trigger` — запустить фетчер вне cron-а

## Этап 8: Ранкер как сервис

- [ ] Вызывать Claude через Anthropic Go SDK напрямую, не через CLI
- [ ] Промт ранкера: вместо «напиши MD» → «верни JSON: `[{vacancy_id, score, group, notes}]`»
- [ ] `POST /rank/trigger` — берёт все vacancies без score для user_id, отправляет в Claude батчами по 50
- [ ] Запись результатов в `vacancy_scores`
- [ ] `GET /reports/:date` — рендер MD-отчёта из БД (опционально, для совместимости)

## Этап 9: Драфт сопроводов через API

- [ ] `POST /applications/:vacancy_id/draft` — Claude SDK + skill-промт из `.agents/skills/application/SKILL.md` → возвращает текст
- [ ] Сохраняет в `applications.draft_body`, status=draft
- [ ] Опционально: `regenerate` со своим хинтом

## Этап 10: Auth (перед публикацией)

- [ ] Magic-link через email (resend) — без регистрации, whitelist в config
- [ ] JWT в cookie, middleware на все ручки кроме `/health`
- [ ] `user_id` из контекста, везде где сейчас хардкод

## Этап 11: React frontend

- [ ] Vite + React + TS, Tailwind, TanStack Query
- [ ] 3 экрана: **Inbox** (свежие вакансии с скорами), **Applications** (kanban по статусам), **Profile** (редактор приоритетов и запросов)
- [ ] Детальная страница вакансии с кнопками «Сгенерировать драфт» / «Отправлено» / «Скип»
- [ ] Кнопка «Запустить fetch» / «Запустить ранкер»

---

## Порядок работы

Критическая цепочка: **1 → 2 → 3 → 4 → 5**, после неё уже можно работать через psql/Adminer и фетчер пишет в БД.

Далее параллельно:
- **6 → 7** (API)
- **8** (ранкер на SDK)

После — **9** (драфты), **10** (auth), **11** (фронт).

Минимальный полезный срез — этапы 1-5 + 6 read-only.
