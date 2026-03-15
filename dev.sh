#!/usr/bin/env bash
# Runs the backend (:8080) and frontend (:3000) together for local dev.
# Logs go to /tmp/pocket-backend.log and /tmp/pocket-frontend.log.
# Ctrl+C stops both. This has no extra dependencies (no concurrently/foreman
# required) — see Procfile for the same two commands if you do have a
# process manager like `honcho start` or `goreman start` installed.
set -euo pipefail
cd "$(dirname "$0")"

BACKEND_LOG=/tmp/pocket-backend.log
FRONTEND_LOG=/tmp/pocket-frontend.log

cleanup() {
  echo "Stopping backend and frontend..."
  kill "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null || true
  wait "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

echo "Starting backend (:8080) -> $BACKEND_LOG"
go run ./cmd/server > "$BACKEND_LOG" 2>&1 &
BACKEND_PID=$!

echo "Waiting for backend health check..."
for i in $(seq 1 30); do
  if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
    echo "Backend is up."
    break
  fi
  if [ "$i" -eq 30 ]; then
    echo "Backend did not become healthy in time. Check $BACKEND_LOG" >&2
    exit 1
  fi
  sleep 1
done

echo "Starting frontend (:3000) -> $FRONTEND_LOG"
pnpm --dir frontend dev > "$FRONTEND_LOG" 2>&1 &
FRONTEND_PID=$!

echo "Both servers running. Backend PID $BACKEND_PID, frontend PID $FRONTEND_PID."
echo "Tail logs with: tail -f $BACKEND_LOG $FRONTEND_LOG"
echo "Press Ctrl+C to stop both."

wait "$BACKEND_PID" "$FRONTEND_PID"
