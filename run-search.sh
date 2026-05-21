#!/bin/bash
set -e
set -o pipefail

REPO_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="$REPO_DIR/logs"
LOG_FILE="$LOG_DIR/search.log"
DEBUG_FILE="$LOG_DIR/search-debug.log"
RAW_FILE="$LOG_DIR/search-stream.jsonl"
PROMPT_FILE="$REPO_DIR/search-config.md"
TIMEOUT_SEC=900

mkdir -p "$LOG_DIR"

# Load .env into the environment so claude subprocess inherits it (HH_USER_EMAIL, etc.)
if [ -f "$REPO_DIR/.env" ]; then
  set -a
  # shellcheck disable=SC1091
  . "$REPO_DIR/.env"
  set +a
fi

START_TS=$(date +%s)

log() {
  local elapsed=$(( $(date +%s) - START_TS ))
  printf "[%s +%ds] %s\n" "$(date -u '+%H:%M:%S')" "$elapsed" "$*" | tee -a "$LOG_FILE"
}

if ! command -v claude >/dev/null 2>&1; then
  log "ERROR: claude CLI не найден. Установи: https://claude.com/claude-code"
  exit 1
fi

if [ ! -f "$PROMPT_FILE" ]; then
  log "ERROR: не найден prompt-файл $PROMPT_FILE"
  exit 1
fi

HAS_JQ=1
command -v jq >/dev/null 2>&1 || HAS_JQ=0

cd "$REPO_DIR"

log "=== Поиск вакансий запущен ==="
log "Claude: $(claude --version 2>&1 | head -n1)"
log "Промт:  $PROMPT_FILE ($(wc -c < "$PROMPT_FILE") байт)"

RESUME_COUNT=$(find "$REPO_DIR/resumes" -maxdepth 1 -type f \( -name '*.pdf' -o -name '*.md' \) 2>/dev/null | wc -l | tr -d ' ')
log "Резюме в resumes/: $RESUME_COUNT файл(ов)"
if [ "$RESUME_COUNT" -eq 0 ]; then
  log "WARNING: в resumes/ пусто. Агенту нечего читать для понимания профиля."
fi

JOBS_BEFORE=$(find "$REPO_DIR/jobs" -maxdepth 1 -type f -name '*.md' 2>/dev/null | wc -l | tr -d ' ')
log "Файлов в jobs/ до запуска: $JOBS_BEFORE"

if [ -z "${HH_USER_EMAIL:-}" ]; then
  log "WARNING: HH_USER_EMAIL не задан (см. .env.example). hh.ru API может вернуть 403 — добавь email в .env."
else
  log "HH_USER_EMAIL: $HH_USER_EMAIL"
fi

log "Таймаут: $((TIMEOUT_SEC / 60)) мин"
log "Лог пишется в: $LOG_FILE"
if [ "$HAS_JQ" -eq 1 ]; then
  log "Сырой JSON-stream: $RAW_FILE"
fi
log ""

# jq-фильтр: превращает каждую строку stream-json в одну короткую человекочитаемую строку
JQ_FILTER='
  . as $raw
  | (fromjson? // null) as $e
  | if $e == null then
      if ($raw | length) > 0 then "raw: " + $raw else empty end
    elif $e.type == "system" and $e.subtype == "init" then
      "init: модель=\($e.model // "?"), инструментов=\(($e.tools // []) | length), cwd=\($e.cwd // "?")"
    elif $e.type == "assistant" then
      ($e.message.content[]? |
        if .type == "text" and ((.text // "") | length > 0) then
          "думает: " + ((.text | gsub("\\s+"; " "))[:220])
        elif .type == "thinking" then
          "размышление..."
        elif .type == "tool_use" then
          "вызов " + .name + ": " + (
            if .name == "WebFetch" then "GET " + (.input.url // "?")
            elif .name == "WebSearch" then "поиск \"" + ((.input.query // "?")[:120]) + "\""
            elif .name == "Bash" then ((.input.description // .input.command // "?")[:160])
            elif .name == "Read" then (.input.file_path // "?")
            elif .name == "Write" then "запись -> " + (.input.file_path // "?")
            elif .name == "Edit" then "правка " + (.input.file_path // "?")
            elif .name == "Grep" then "/" + (.input.pattern // "?") + "/" + (if .input.path then " в " + .input.path else "" end)
            elif .name == "Glob" then (.input.pattern // "?")
            elif .name == "Task" or .name == "Agent" then "подагент: " + ((.input.description // "?")[:160])
            else ((.input | tostring)[:160])
            end
          )
        else empty end
      )
    elif $e.type == "user" then
      ($e.message.content[]? |
        select(.type == "tool_result") |
        if (.is_error // false) == true then
          "ОШИБКА инструмента: " + ((.content | tostring | gsub("\\s+"; " "))[:240])
        else empty end
      )
    elif $e.type == "result" then
      if $e.subtype == "success" then
        "агент закончил: " + ((($e.result // "") | tostring | gsub("\\s+"; " "))[:240])
      else
        "агент остановлен (\($e.subtype)): " + ((($e.error // "") | tostring | gsub("\\s+"; " "))[:240])
      end
    else empty end
'

set +e
if [ "$HAS_JQ" -eq 1 ]; then
  log "Запускаю claude (stream-json + jq). Прогресс в реальном времени:"
  log "──────────────────────────────────────────────────────"
  timeout "$TIMEOUT_SEC" claude \
    --print \
    --verbose \
    --model claude-haiku-4-5 \
    --output-format stream-json \
    --dangerously-skip-permissions \
    --debug-file "$DEBUG_FILE" \
    "$(cat "$PROMPT_FILE")" 2>&1 \
    | tee -a "$RAW_FILE" \
    | jq -Rrc --unbuffered "$JQ_FILTER" \
    | while IFS= read -r line; do log "$line"; done
  EXIT_CODE=${PIPESTATUS[0]}
else
  log "WARNING: jq не найден — детальный прогресс отключён. Установи jq для построчного лога."
  log "Запускаю claude. Вывод в реальном времени:"
  log "──────────────────────────────────────────────────────"
  timeout "$TIMEOUT_SEC" claude \
    --print \
    --model claude-haiku-4-5 \
    --dangerously-skip-permissions \
    --debug-file "$DEBUG_FILE" \
    "$(cat "$PROMPT_FILE")" 2>&1 | tee -a "$LOG_FILE"
  EXIT_CODE=${PIPESTATUS[0]}
fi
set -e

log "──────────────────────────────────────────────────────"

ELAPSED=$(( $(date +%s) - START_TS ))
ELAPSED_MIN=$((ELAPSED / 60))
ELAPSED_SEC=$((ELAPSED % 60))

case "$EXIT_CODE" in
  0)
    JOBS_AFTER=$(find "$REPO_DIR/jobs" -maxdepth 1 -type f -name '*.md' 2>/dev/null | wc -l | tr -d ' ')
    NEW_FILES=$((JOBS_AFTER - JOBS_BEFORE))
    log "OK: Поиск завершён за ${ELAPSED_MIN}m ${ELAPSED_SEC}s"
    log "Новых отчётов в jobs/: $NEW_FILES"
    LATEST=$(find "$REPO_DIR/jobs" -maxdepth 1 -type f -name '*.md' -printf '%T@ %p\n' 2>/dev/null | sort -nr | head -n1 | cut -d' ' -f2-)
    [ -n "$LATEST" ] && log "Свежий отчёт: $LATEST"
    ;;
  124)
    log "ERROR: Таймаут после $((TIMEOUT_SEC / 60)) мин. Детали: $DEBUG_FILE"
    exit 124
    ;;
  *)
    log "ERROR: claude exited with code $EXIT_CODE. Детали: $DEBUG_FILE"
    exit "$EXIT_CODE"
    ;;
esac
