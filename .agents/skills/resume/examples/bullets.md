# Bullet Examples — Good vs Bad

Collection of before/after rewrites following the Google XYZ formula:
**Accomplished [X], as measured by [Y], by doing [Z]**

Use these as patterns when rewriting any bullet. Match the *shape* — your specifics will differ.

---

## English examples (industry-standard)

### Latency / performance

❌ "Worked on backend services and improved performance"
✅ "Reduced API latency by 71% by introducing Redis caching layer, serving 2M+ daily requests"

❌ "Optimized slow queries"
✅ "Cut P95 query time from 800ms to 120ms by adding composite indexes and rewriting N+1 JOINs across the order pipeline"

❌ "Improved system performance"
✅ "Refactored backend services, reducing API response time by 43% and improving uptime to 99.97% across 12M+ user sessions monthly"

### Architecture / migrations

❌ "Migrated services to Kubernetes"
✅ "Led migration of 14 services to Kubernetes, cutting deployment errors by 88% and enabling zero-downtime rollouts"

❌ "Designed and built out the backend for a client application"
✅ "Designed a 6-service backend (NodeJS, Postgres, Redis) that scaled to 200K users and generated $1.4M annual revenue"

### Features / impact

❌ "Built a feature flag system"
✅ "Implemented LaunchDarkly feature flags adopted by 4 product teams, reducing rollback time from 30 min to under 1 min"

❌ "Improved frontend performance"
✅ "Increased mobile app Lighthouse score from 45 to 92 via lazy rendering + image optimization — added 500K monthly active users"

### Cost / efficiency

❌ "Reduced cloud costs"
✅ "Led infrastructure rightsizing across 40 services, saving $180K/year in AWS spend without impacting SLOs"

---

## Russian examples (для русскоязычного рынка)

### Backend / архитектура

❌ "Разрабатывал backend-сервисы"
✅ "Разработал 8+ Go-микросервисов (Clean Architecture, gRPC + RabbitMQ) — независимый деплой каждого компонента без координации между командами"

❌ "Работал с микросервисами"
✅ "Спроектировал с нуля backend из 10+ Go-микросервисов; сократил цикл выкатки фич с 2 дней до 4 часов через независимый CI/CD на сервис"

### Performance / оптимизация

❌ "Снизил латентность горячих эндпоинтов через Redis-кэширование"
✅ "Снизил P95 latency критичных API с 800мс до 120мс через композитные индексы, JOIN-рефакторинг и Redis-кэш горячих эндпоинтов"

❌ "Оптимизировал запросы к базе"
✅ "Оптимизировал тяжёлые PostgreSQL-запросы через EXPLAIN ANALYZE и переписывание JOIN; сократил среднее время ответа orders-API в 6 раз"

### AI / агенты

❌ "Работал с AI-агентами"
✅ "Спроектировал мультиагентный pipeline на Claude Code (orchestrator / implementer / reviewer) — сократил время написания нового CRUD-сервиса с 2 недель до 2 дней"

❌ "Использовал Anthropic API в продакшене"
✅ "Внедрил мультиагентную систему с eval-слоем (reviewer-агент валидирует архитектуру) и retry/timeout-стратегиями — автоматизировал ~80% boilerplate"

### Воркеры / надёжность

❌ "Реализовал фоновые воркеры"
✅ "Реализовал фоновые воркеры с retry + DLQ для пайплайнов обработки данных — гарантированная доставка ~1 000+ задач/день при сбоях downstream"

❌ "Делал интеграции с внешними API"
✅ "Реализовал интеграции с внешними API (rate limiting, exponential backoff, timeout handling) — нулевые потери запросов при нестабильности upstream"

### Команда / стандарты

❌ "Помогал команде с DevOps"
✅ "Стандартизировал Docker + CI/CD-шаблоны для команды из ~20 разработчиков (health-checks, structured logging, graceful degradation) — онбординг нового сервиса по единому шаблону"

❌ "Делал ревью и менторил коллег"
✅ "Внедрил парные ревью и архитектурные RFC для команды из 12 backend-инженеров — снизил баг-rate в проде на ~40% за полгода"

---

## Anti-patterns — never write

| Pattern | Why bad |
|---------|---------|
| "Responsible for X" | Describes a role, not an achievement |
| "Helped with X" / "Participated in X" | Hides your actual ownership |
| "Worked on X" | Vague; everyone "works on" things |
| "Various technologies" | Be specific, list them |
| "Improved performance" without a number | Show the number or don't claim improvement |
| "Wrote clean code" / "Followed best practices" | Self-praise without evidence — every engineer claims this |
| "Team player" / "Detail-oriented" | Soft claims without proof; cut |
| Listing 15+ skills without context | Recruiter doesn't believe you're expert in all 15 |

---

## Quick "is this bullet ready?" checklist

- [ ] Starts with a past-tense action verb (Designed / Built / Reduced / Led / Migrated)?
- [ ] Has at least one number (%, users, $, ms, requests/sec, hours saved)?
- [ ] Names the specific tech stack used inside the bullet?
- [ ] Answers "so what?" — what did this enable for the business?
- [ ] Under 2 lines?

If any box is empty — rewrite it.
