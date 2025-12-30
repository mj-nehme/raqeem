#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(dirname "$0")"
COMPONENT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ROOT_DIR="$(cd "$COMPONENT_DIR/../.." && pwd)"

# Load env
if [[ -f "$COMPONENT_DIR/.env" ]]; then
  set -o allexport
  source "$COMPONENT_DIR/.env"
  set +o allexport
fi

source "$ROOT_DIR/scripts/service-discovery.sh"
source "$ROOT_DIR/scripts/preflight.sh"

NAMESPACE=${NAMESPACE:-default}
PORT_START=${MENTOR_FRONTEND_START_PORT:-5000}

# Preflight: ensure Kubernetes is reachable; check node/npm locally
ensure_kube_ready || exit 1
for cmd in node npm; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "❌ Missing $cmd"; exit 1; }
done

MENTOR_NODEPORT=$(get_nodeport "mentor-backend" "$NAMESPACE")
if [[ -z "$MENTOR_API_URL" && -n "$MENTOR_NODEPORT" ]]; then
  MENTOR_API_URL="http://localhost:$MENTOR_NODEPORT"
fi
[[ -n "$MENTOR_API_URL" ]] || { echo "❌ Mentor API URL not found; start mentor-backend or set MENTOR_API_URL"; exit 1; }

PORT=$(find_available_port $PORT_START 5)
[[ $? -eq 0 ]] || { echo "❌ No available port in range $PORT_START-$((PORT_START+4))"; exit 1; }

mkdir -p "$ROOT_DIR/.deploy"
echo "  - Starting Dashboard on port $PORT..."
cd "$COMPONENT_DIR"
npm install --silent
VITE_MENTOR_FRONTEND_PORT=$PORT \
VITE_MENTOR_API_URL="$MENTOR_API_URL" \
nohup npm run dev > "$ROOT_DIR/.deploy/dashboard.log" 2>&1 &
PID=$!

register_service "dashboard" "http://localhost:$PORT" "$PORT"
echo "PID=$PID" > "$ROOT_DIR/.deploy/dashboard.pid"

sleep 3
kill -0 $PID 2>/dev/null || { echo "❌ Dashboard failed - check .deploy/dashboard.log"; exit 1; }
echo "✅ Dashboard ready at http://localhost:$PORT"
