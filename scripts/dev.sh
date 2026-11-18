#!/usr/bin/env bash
# FastVPN Linux GUI - Live Development with Vite Hot Reload & Go Backend

set -e

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export FRONTEND_DEVSERVER_URL="http://localhost:5173"

echo "=========================================================="
echo "    FastVPN GUI - Live Dev Mode (Vite HMR + Go Backend)   "
echo "=========================================================="

# Cleanup child processes on exit
cleanup() {
  echo ""
  echo "==> Shutting down dev servers..."
  if [ -n "$VITE_PID" ] && kill -0 "$VITE_PID" 2>/dev/null; then
    kill "$VITE_PID" 2>/dev/null || true
  fi
  exit 0
}
trap cleanup SIGINT SIGTERM EXIT

# 1. Start Vite Dev Server with Hot Module Replacement (HMR)
echo "==> [1/2] Starting Vite Frontend on http://localhost:5173..."
cd "$REPO_DIR/frontend"
npm run dev -- --host 127.0.0.1 --port 5173 &
VITE_PID=$!

# 2. Wait for Vite to be ready
echo "==> Waiting for Vite dev server..."
for i in {1..30}; do
  if curl -s "http://127.0.0.1:5173" >/dev/null 2>&1; then
    echo "    Vite is ready!"
    break
  fi
  sleep 0.5
done

# 3. Launch Wails Go Desktop App connected to Vite HMR
echo "==> [2/2] Launching Go backend with live proxy..."
cd "$REPO_DIR"

if command -v air >/dev/null 2>&1; then
  echo "    Using 'air' for Go backend hot-reloading..."
  air
else
  echo "    Running 'go run main.go'..."
  go run main.go
fi

