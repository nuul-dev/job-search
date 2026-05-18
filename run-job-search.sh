#!/bin/bash
set -e

export HOME=/home/nuul
export PATH="/home/nuul/.local/bin:$PATH"

REPO_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="$REPO_DIR/logs"
LOG_FILE="$LOG_DIR/job-search.log"
DEBUG_FILE="$LOG_DIR/job-search-debug.log"
PROMPT_FILE="$REPO_DIR/daily-search-jobs.md"

mkdir -p "$LOG_DIR"

log() { echo "[$(date -u '+%H:%M:%S')] $*" | tee -a "$LOG_FILE"; }

[ -f "$REPO_DIR/.env" ] && set -a && source "$REPO_DIR/.env" && set +a

if [ -z "$GITHUB_PAT" ]; then
  log "ERROR: GITHUB_PAT is not set. Add it to .env or export before running."
  exit 1
fi

log "=== Job search started ==="
log "Claude: $(which claude) — $(claude --version 2>&1)"
log "Prompt: $PROMPT_FILE ($(wc -c < "$PROMPT_FILE") bytes)"

cd "$REPO_DIR"

log "Starting claude (timeout 15m)..."
timeout 900 claude \
  --print \
  --dangerously-skip-permissions \
  --debug-file "$DEBUG_FILE" \
  "$(envsubst < "$PROMPT_FILE")" >> "$LOG_FILE" 2>&1

EXIT_CODE=$?
log "Claude exited with code: $EXIT_CODE"
[ $EXIT_CODE -eq 124 ] && log "ERROR: Timed out after 15 minutes"
[ $EXIT_CODE -ne 0 ] && [ $EXIT_CODE -ne 124 ] && log "ERROR: See $DEBUG_FILE for details"

log "=== Job search finished ==="
