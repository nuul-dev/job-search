#!/bin/bash
set -e

REPO_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="$REPO_DIR/logs"
LOG_FILE="$LOG_DIR/search.log"
DEBUG_FILE="$LOG_DIR/search-debug.log"
PROMPT_FILE="$REPO_DIR/search-config.md"

mkdir -p "$LOG_DIR"

log() { echo "[$(date -u '+%H:%M:%S')] $*" | tee -a "$LOG_FILE"; }

if ! command -v claude >/dev/null 2>&1; then
  log "ERROR: claude CLI not found. Install: https://claude.com/claude-code"
  exit 1
fi

log "=== Search started ==="
log "Claude: $(which claude) ($(claude --version 2>&1))"
log "Prompt: $PROMPT_FILE ($(wc -c < "$PROMPT_FILE") bytes)"

cd "$REPO_DIR"

log "Starting claude (timeout 15m)..."
timeout 900 claude \
  --print \
  --dangerously-skip-permissions \
  --debug-file "$DEBUG_FILE" \
  "$(cat "$PROMPT_FILE")" >> "$LOG_FILE" 2>&1

EXIT_CODE=$?
log "Claude exited with code: $EXIT_CODE"
[ $EXIT_CODE -eq 124 ] && log "ERROR: Timed out after 15 minutes"
[ $EXIT_CODE -ne 0 ] && [ $EXIT_CODE -ne 124 ] && log "ERROR: See $DEBUG_FILE for details"

log "=== Search finished ==="
