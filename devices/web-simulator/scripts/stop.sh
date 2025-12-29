#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(dirname "$0")"
COMPONENT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ROOT_DIR="$(cd "$COMPONENT_DIR/../.." && pwd)"

PID_FILE="$ROOT_DIR/.deploy/devices-web-simulator.pid"

echo "🛑 Stopping Devices Web Simulator frontend..."
if [[ -f "$PID_FILE" ]]; then
  source "$PID_FILE"
  if [[ -n "$PID" ]] && kill -0 "$PID" 2>/dev/null; then
    kill "$PID" 2>/dev/null || true
    sleep 1
    kill -9 "$PID" 2>/dev/null || true
  fi
  rm -f "$PID_FILE"
fi
echo "✅ Devices Web Simulator stopped"
