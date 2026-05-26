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

if [ -f "$REPO_DIR/.env" ]; then
  set -a; . "$REPO_DIR/.env"; set +a
fi

START_TS=$(date +%s)

# ── Colors & terminal ────────────────────────────────────────────────────
IS_TTY=0; [ -t 1 ] && IS_TTY=1
RED='\033[0;31m' YEL='\033[0;33m' GRN='\033[0;32m'
CYN='\033[0;36m' DIM='\033[2m'    BLD='\033[1m'    RST='\033[0m'

# ── Spinner + phase bar ──────────────────────────────────────────────────
# Precomputed bars for phases 0-4 (20 chars each, 5 per phase)
_BARS=("░░░░░░░░░░░░░░░░░░░░" "█████░░░░░░░░░░░░░░░" "██████████░░░░░░░░░░" "███████████████░░░░░" "████████████████████")

SPIN_PID_FILE="$(mktemp)"
SPIN_MSG_FILE="$(mktemp)"
PHASE_FILE="$(mktemp)"
echo "0" > "$PHASE_FILE"

_spinner() {
  local frames='⠋⠙⠹⠸⠼⠴⠦⠧⠣⠏' i=0 c p bar msg
  while true; do
    c="${frames:$(( i % 10 )):1}"
    p=$(cat "$PHASE_FILE" 2>/dev/null || echo "0")
    bar="${_BARS[$p]:-${_BARS[0]}}"
    msg=$(cat "$SPIN_MSG_FILE" 2>/dev/null || echo "...")
    printf "\r  ${CYN}%s${RST}  [%s] ${DIM}%d/4  %s${RST}\033[K" \
      "$c" "$bar" "$p" "$msg" > /dev/tty 2>/dev/null
    i=$(( i + 1 ))
    sleep 0.08
  done
}

_spin_pause() {
  local spid; spid=$(cat "$SPIN_PID_FILE" 2>/dev/null)
  [ -n "$spid" ] && kill -STOP "$spid" 2>/dev/null || true
  printf "\r\033[2K" > /dev/tty 2>/dev/null
}
_spin_resume() {
  local spid; spid=$(cat "$SPIN_PID_FILE" 2>/dev/null)
  [ -n "$spid" ] && kill -CONT "$spid" 2>/dev/null || true
}

cleanup() {
  local spid; spid=$(cat "$SPIN_PID_FILE" 2>/dev/null)
  [ -n "$spid" ] && { kill "$spid" 2>/dev/null; wait "$spid" 2>/dev/null; }
  printf "\r\033[2K" > /dev/tty 2>/dev/null
  rm -f "$SPIN_PID_FILE" "$SPIN_MSG_FILE" "$PHASE_FILE"
}
trap cleanup EXIT INT TERM

# ── log() ────────────────────────────────────────────────────────────────
log() {
  local elapsed=$(( $(date +%s) - START_TS ))
  local ts; ts="$(date -u '+%H:%M:%S')"
  printf '[%s +%ds] %s\n' "$ts" "$elapsed" "$*" >> "$LOG_FILE"

  if [ "$IS_TTY" = 0 ]; then
    printf '[%s +%ds] %s\n' "$ts" "$elapsed" "$*"; return
  fi

  local label; label=$(printf '%s' "$*" | awk '{print $1}')
  local rest;  rest="$(printf '%s' "$*" | cut -c$(( ${#label} + 1 ))-)"

  # Advance phase + spinner message based on what's happening
  case "$label" in
    init|read)    printf '1' > "$PHASE_FILE"; printf '%s' "$rest" > "$SPIN_MSG_FILE" ;;
    search|fetch) printf '2' > "$PHASE_FILE"; printf '%s' "$rest" > "$SPIN_MSG_FILE" ;;
    report)       printf '3' > "$PHASE_FILE"; printf '%s' > "$SPIN_MSG_FILE" ;;
    letter)       printf '4' > "$PHASE_FILE"; printf '%s' > "$SPIN_MSG_FILE" ;;
  esac

  _spin_pause

  case "$label" in
    error|failed) printf "  ${RED}%-8s${RST}%s\n" "$label" "$rest" ;;
    warn)         printf "  ${YEL}%-8s${RST}%s\n" "$label" "$rest" ;;
    done)         printf "  ${GRN}✓${RST}  ${BLD}%s${RST}\n" "$rest" ;;
    ─*)           printf "  ${DIM}%s${RST}\n" "$*" ;;
    =*)           printf "\n  ${BLD}%s${RST}\n" "$*" ;;
    *)            printf "  %-8s%s\n" "$label" "$rest" ;;
  esac

  _spin_resume
}

# ── Preflight ────────────────────────────────────────────────────────────
if ! command -v claude >/dev/null 2>&1; then
  log "error   claude CLI not found — https://claude.com/claude-code"; exit 1
fi
if [ ! -f "$PROMPT_FILE" ]; then
  log "error   prompt file not found: $PROMPT_FILE"; exit 1
fi

HAS_JQ=1; command -v jq >/dev/null 2>&1 || HAS_JQ=0
cd "$REPO_DIR"

RESUME_COUNT=$(find "$REPO_DIR/resumes" -maxdepth 1 -type f \( -name '*.pdf' -o -name '*.md' \) 2>/dev/null | wc -l | tr -d ' ')
JOBS_BEFORE=$(find  "$REPO_DIR/jobs"    -maxdepth 1 -type f -name '*.md'                           2>/dev/null | wc -l | tr -d ' ')

log "=== job search ==="
log "model:   claude-haiku-4-5  timeout: $((TIMEOUT_SEC / 60))m"
log "resumes: $RESUME_COUNT   jobs before: $JOBS_BEFORE"
[ -z "${HH_USER_EMAIL:-}" ] && log "warn    HH_USER_EMAIL not set — hh.ru may 403 (check .env)"
log "──────────────────────────────────────────────────────"

# ── JQ filter: only meaningful events, no thinking/reasoning ────────────
JQ_FILTER='
  (fromjson? // null) as $e
  | if $e == null then empty
    elif $e.type == "system" and $e.subtype == "init" then
      "init    \($e.model // "?")  ·  \(($e.tools // []) | length) tools"
    elif $e.type == "assistant" then
      ($e.message.content[]? |
        if .type == "tool_use" then
          .name as $n | .input as $in |
          if $n == "WebFetch" then
            ($in.url // "") as $url |
            if   ($url | test("hh\\.ru"))       then "search  hh.ru        · " + ($url | gsub(".*text=|&.*";"") | @uri | gsub("%2[0Bb]";" ") | .[0:55])
            elif ($url | test("career\\.habr"))  then "search  Habr Career  · " + ($url | gsub(".*[?&]q=|&.*";"") | .[0:55])
            elif ($url | test("hirify"))         then "search  Hirify       · " + ($url | gsub(".*search=|&.*";"") | .[0:55])
            elif ($url | test("getmatch"))       then "search  GetMatch     · " + ($url | gsub(".*[?&]q=|&.*";"") | .[0:55])
            elif ($url | test("web3\\.career"))  then "search  web3.career"
            elif ($url | test("bondex"))         then "search  Bondex"
            else empty end
          elif $n == "Bash" then
            ($in.command // "") as $cmd |
            if   ($cmd | test("pdftotext|pdfplumber")) then "read    resumes/"
            elif ($cmd | test("curl.*hh\\.ru"))        then
              (($cmd | capture("text=(?<q>[^&\" ]+)") | .q) // "?") as $q |
              "search  hh.ru API   · " + ($q | @uri | gsub("%2[0Bb]";" ") | .[0:55])
            else empty end
          elif $n == "Read" then
            ($in.file_path // "") as $p |
            if ($p | test("resumes/")) then "read    " + ($p | split("/")[-1]) else empty end
          elif $n == "Write" then
            ($in.file_path // "") as $p |
            if   ($p | test("/jobs/"))         then "report  " + ($p | split("/")[-1])
            elif ($p | test("/applications/")) then "letter  " + ($p | split("/")[-1])
            else empty end
          elif $n == "Edit" then
            ($in.file_path // "") as $p |
            if ($p | test("/applications/")) then "letter  " + ($p | split("/")[-1]) + " (edit)" else empty end
          else empty end
        else empty end
      )
    elif $e.type == "user" then
      ($e.message.content[]? |
        select(.type == "tool_result") |
        select((.is_error // false) == true) |
        "error   " + ((.content | tostring | gsub("\\s+"; " "))[:180])
      )
    elif $e.type == "result" then
      if $e.subtype == "success" then empty
      else "failed  \($e.subtype): " + (($e.error // "") | tostring | gsub("\\s+"; " "))[:180]
      end
    else empty end
'

# ── Run ──────────────────────────────────────────────────────────────────
[ "$IS_TTY" = 1 ] && { _spinner & echo $! > "$SPIN_PID_FILE"; }

set +e
if [ "$HAS_JQ" -eq 1 ]; then
  timeout "$TIMEOUT_SEC" claude \
    --print --verbose \
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
  log "warn    jq not found — no live progress. Install jq."
  timeout "$TIMEOUT_SEC" claude \
    --print --model claude-haiku-4-5 \
    --dangerously-skip-permissions \
    --debug-file "$DEBUG_FILE" \
    "$(cat "$PROMPT_FILE")" 2>&1 | tee -a "$LOG_FILE"
  EXIT_CODE=${PIPESTATUS[0]}
fi
set -e

log "──────────────────────────────────────────────────────"

ELAPSED=$(( $(date +%s) - START_TS ))

case "$EXIT_CODE" in
  0)
    JOBS_AFTER=$(find "$REPO_DIR/jobs" -maxdepth 1 -type f -name '*.md' 2>/dev/null | wc -l | tr -d ' ')
    NEW_LETTERS=$(find "$REPO_DIR/applications" -maxdepth 1 -name '*.md' -newer "$LOG_FILE" 2>/dev/null | wc -l | tr -d ' ')
    LATEST=$(find "$REPO_DIR/jobs" -maxdepth 1 -type f -name '*.md' -printf '%T@ %p\n' 2>/dev/null | sort -nr | head -n1 | cut -d' ' -f2-)
    log "done    $((ELAPSED/60))m $((ELAPSED%60))s  ·  +$((JOBS_AFTER - JOBS_BEFORE)) report(s)  ·  +${NEW_LETTERS} letter(s)"
    [ -n "$LATEST" ] && log "report  $LATEST"
    ;;
  124) log "error   timeout after $((TIMEOUT_SEC/60))m — see $DEBUG_FILE"; exit 124 ;;
  *)   log "error   claude exited $EXIT_CODE — see $DEBUG_FILE";           exit "$EXIT_CODE" ;;
esac
