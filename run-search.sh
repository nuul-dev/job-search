#!/bin/bash
set -e
set -o pipefail

REPO_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="$REPO_DIR/logs"
LOG_FILE="$LOG_DIR/search.log"
DEBUG_FILE="$LOG_DIR/search-debug.log"
PROMPT_FILE="$REPO_DIR/search-config.md"
TIMEOUT_SEC=900

mkdir -p "$LOG_DIR"

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

log "Таймаут: $((TIMEOUT_SEC / 60)) мин"
log "Лог пишется в: $LOG_FILE"
log ""
log "Запускаю claude. Вывод агента в реальном времени:"
log "──────────────────────────────────────────────────────"

set +e
timeout "$TIMEOUT_SEC" claude \
  --print \
  --dangerously-skip-permissions \
  --debug-file "$DEBUG_FILE" \
  "$(cat "$PROMPT_FILE")" 2>&1 | tee -a "$LOG_FILE"
EXIT_CODE=${PIPESTATUS[0]}
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
