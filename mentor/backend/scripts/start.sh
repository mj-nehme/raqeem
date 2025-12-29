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
MENTOR_FRONTEND_START_PORT=${MENTOR_FRONTEND_PORT_RANGE_START:-5000}
DEVICES_FRONTEND_START_PORT=${DEVICES_FRONTEND_PORT_RANGE_START:-4000}
MENTOR_BACKEND_PORT_PREF=${MENTOR_BACKEND_PREFERRED_NODEPORT:-30090}

source "$ROOT_DIR/scripts/service-discovery.sh"

echo "🚀 Starting Mentor Backend (namespace=$NAMESPACE)..."

for cmd in kubectl helm docker; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "❌ Missing $cmd"; exit 1; }
done
kubectl cluster-info >/dev/null 2>&1 || { echo "❌ kubectl cannot reach cluster"; exit 1; }

if [[ "$BUILD_IMAGES" == "true" ]]; then
  echo "🔨 Building local backend images..."
  "$ROOT_DIR/scripts/build-local-images.sh"
fi

# CORS regex for FE ranges
MENTOR_RANGE_START=$MENTOR_FRONTEND_START_PORT
MENTOR_RANGE_END=$((MENTOR_RANGE_START + 4))
DEVICES_RANGE_START=$DEVICES_FRONTEND_START_PORT
DEVICES_RANGE_END=$((DEVICES_RANGE_START + 4))
CORS_REGEX="^http://localhost:($(seq -s'|' $DEVICES_RANGE_START $DEVICES_RANGE_END)|$(seq -s'|' $MENTOR_RANGE_START $MENTOR_RANGE_END))\$"

# NodePort selection
MENTOR_BACKEND_PORT=$(find_available_backend_port "$MENTOR_BACKEND_PORT_PREF" 5)
[[ $? -eq 0 ]] || { echo "❌ No available NodePort near $MENTOR_BACKEND_PORT_PREF"; exit 1; }

# Devices API URL discovery if not set
DEVICES_NODEPORT=$(get_nodeport "devices-backend" "$NAMESPACE")
if [[ -z "$DEVICES_API_URL" && -n "$DEVICES_NODEPORT" ]]; then
  DEVICES_API_URL="http://localhost:$DEVICES_NODEPORT/api/v1"
fi

echo "📦 Helm upgrade/install mentor-backend..."
HELM_ARGS=(
  --namespace "$NAMESPACE"
  --set service.nodePort="$MENTOR_BACKEND_PORT"
  --set image.tag="latest"
  --set image.pullPolicy="IfNotPresent"
  --set-string frontendOriginRegex="$CORS_REGEX"
)
if [[ -n "$DEVICES_API_URL" ]]; then
  HELM_ARGS+=( --set-string devicesApiUrl="$DEVICES_API_URL" )
fi

helm upgrade --install mentor-backend "$ROOT_DIR/charts/mentor-backend" "${HELM_ARGS[@]}"
wait_for_service_ready "mentor-backend" "$NAMESPACE" 300

MENTOR_NODEPORT=$(get_nodeport "mentor-backend" "$NAMESPACE")
register_service "mentor-backend" "http://localhost:$MENTOR_NODEPORT" "$MENTOR_NODEPORT"

echo "✅ Mentor Backend ready at http://localhost:$MENTOR_NODEPORT"
