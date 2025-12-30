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
DEVICES_BACKEND_PORT_PREF=${DEVICES_BACKEND_PREFERRED_NODEPORT:-30080}

source "$ROOT_DIR/scripts/service-discovery.sh"
source "$ROOT_DIR/scripts/preflight.sh"

echo "🚀 Starting Devices Backend (namespace=$NAMESPACE)..."

# Preflight
ensure_docker_running || exit 1
ensure_helm_ready || exit 1
ensure_kube_ready || exit 1

if [[ "$BUILD_IMAGES" == "true" ]]; then
  echo "🔨 Building local backend images..."
  "$ROOT_DIR/scripts/build-local-images.sh"
fi

# CORS regex: default allows any localhost port; override via CORS_ORIGIN_REGEX
CORS_REGEX=${CORS_ORIGIN_REGEX:-^http://localhost(:[0-9]+)?$}

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
