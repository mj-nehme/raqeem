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

NAMESPACE=${NAMESPACE:-default}
BUILD_IMAGES=${BUILD_IMAGES:-true}
DEVICES_FRONTEND_START_PORT=${DEVICES_FRONTEND_PORT_RANGE_START:-4000}
MENTOR_FRONTEND_START_PORT=${MENTOR_FRONTEND_PORT_RANGE_START:-5000}
DEVICES_BACKEND_PORT_PREF=${DEVICES_BACKEND_PREFERRED_NODEPORT:-30080}

source "$ROOT_DIR/scripts/service-discovery.sh"

echo "🚀 Starting Devices Backend (namespace=$NAMESPACE)..."

for cmd in kubectl helm docker; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "❌ Missing $cmd"; exit 1; }
done
kubectl cluster-info >/dev/null 2>&1 || { echo "❌ kubectl cannot reach cluster"; exit 1; }

if [[ "$BUILD_IMAGES" == "true" ]]; then
  echo "🔨 Building local backend images..."
  "$ROOT_DIR/scripts/build-local-images.sh"
fi

# CORS regex for FE ranges
DEVICES_RANGE_START=$DEVICES_FRONTEND_START_PORT
DEVICES_RANGE_END=$((DEVICES_RANGE_START + 4))
MENTOR_RANGE_START=$MENTOR_FRONTEND_START_PORT
MENTOR_RANGE_END=$((MENTOR_RANGE_START + 4))
CORS_REGEX="^http://localhost:($(seq -s'|' $DEVICES_RANGE_START $DEVICES_RANGE_END)|$(seq -s'|' $MENTOR_RANGE_START $MENTOR_RANGE_END))\$"

# NodePort selection
DEVICES_BACKEND_PORT=$(find_available_backend_port "$DEVICES_BACKEND_PORT_PREF" 5)
[[ $? -eq 0 ]] || { echo "❌ No available NodePort near $DEVICES_BACKEND_PORT_PREF"; exit 1; }

# Mentor API URL required
if [[ -z "$MENTOR_API_URL" ]]; then
  MENTOR_NODEPORT=$(get_nodeport "mentor-backend" "$NAMESPACE")
  if [[ -n "$MENTOR_NODEPORT" ]]; then
    MENTOR_API_URL="http://localhost:$MENTOR_NODEPORT"
  fi
fi
[[ -n "$MENTOR_API_URL" ]] || { echo "❌ Mentor API URL not found; start mentor-backend or set MENTOR_API_URL"; exit 1; }

echo "📦 Helm upgrade/install devices-backend..."
helm upgrade --install devices-backend "$ROOT_DIR/charts/devices-backend" \
  --namespace "$NAMESPACE" \
  --set service.nodePort="$DEVICES_BACKEND_PORT" \
  --set image.tag="latest" \
  --set image.pullPolicy="IfNotPresent" \
  --set-string mentorApiUrl="$MENTOR_API_URL" \
  --set-string frontendOriginRegex="$CORS_REGEX"

wait_for_service_ready "devices-backend" "$NAMESPACE" 300

DEVICES_NODEPORT=$(get_nodeport "devices-backend" "$NAMESPACE")
register_service "devices-backend" "http://localhost:$DEVICES_NODEPORT" "$DEVICES_NODEPORT"

echo "✅ Devices Backend ready at http://localhost:$DEVICES_NODEPORT"
