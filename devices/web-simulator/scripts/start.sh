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
PORT_START=${DEVICES_FRONTEND_START_PORT:-4000}

# Preflight: ensure Kubernetes is reachable; check node/npm locally
ensure_kube_ready || exit 1
for cmd in node npm; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "❌ Missing $cmd"; exit 1; }
done

DEVICES_NODEPORT=$(get_nodeport "devices-backend" "$NAMESPACE")
if [[ -z "$DEVICES_API_URL" && -n "$DEVICES_NODEPORT" ]]; then
  DEVICES_API_URL="http://localhost:$DEVICES_NODEPORT/api/v1"
fi
[[ -n "$DEVICES_API_URL" ]] || { echo "❌ Devices API URL not found; start devices-backend or set DEVICES_API_URL"; exit 1; }

PORT=$(find_available_port $PORT_START 5)
[[ $? -eq 0 ]] || { echo "❌ No available port in range $PORT_START-$((PORT_START+4))"; exit 1; }

mkdir -p "$ROOT_DIR/.deploy"
echo "  - Starting Devices Web Simulator on port $PORT..."
cd "$COMPONENT_DIR"
npm install --silent
VITE_DEVICES_FRONTEND_PORT=$PORT \
VITE_DEVICES_API_URL="$DEVICES_API_URL" \
nohup npm run dev > "$ROOT_DIR/.deploy/devices-web-simulator.log" 2>&1 &
PID=$!

register_service "devices-web-simulator" "http://localhost:$PORT" "$PORT"
echo "PID=$PID" > "$ROOT_DIR/.deploy/devices-web-simulator.pid"

sleep 3
kill -0 $PID 2>/dev/null || { echo "❌ Devices Web Simulator failed - check .deploy/devices-web-simulator.log"; exit 1; }
echo "✅ Devices Web Simulator ready at http://localhost:$PORT"
