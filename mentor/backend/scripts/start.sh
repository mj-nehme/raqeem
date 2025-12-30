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
MENTOR_BACKEND_PORT_PREF=${MENTOR_BACKEND_PREFERRED_NODEPORT:-30090}

source "$ROOT_DIR/scripts/service-discovery.sh"
source "$ROOT_DIR/scripts/preflight.sh"

echo "🚀 Starting Mentor Backend (namespace=$NAMESPACE)..."

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

# Optional object storage overrides from .env (BUCKET_*)
if [[ -n "$BUCKET_ENDPOINT" ]]; then
  HELM_ARGS+=( --set-string minio.endpoint="$BUCKET_ENDPOINT" )
fi
if [[ -n "$BUCKET_PUBLIC_ENDPOINT" ]]; then
  HELM_ARGS+=( --set-string minio.publicEndpoint="$BUCKET_PUBLIC_ENDPOINT" )
fi
if [[ -n "$BUCKET_NAME" ]]; then
  HELM_ARGS+=( --set-string minio.bucket="$BUCKET_NAME" )
fi

helm upgrade --install mentor-backend "$ROOT_DIR/charts/mentor-backend" "${HELM_ARGS[@]}"
wait_for_service_ready "mentor-backend" "$NAMESPACE" 300

MENTOR_NODEPORT=$(get_nodeport "mentor-backend" "$NAMESPACE")
register_service "mentor-backend" "http://localhost:$MENTOR_NODEPORT" "$MENTOR_NODEPORT"

echo "✅ Mentor Backend ready at http://localhost:$MENTOR_NODEPORT"
