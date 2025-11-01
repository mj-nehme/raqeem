#!/usr/bin/env bash
set -e

# Track what has been started for error reporting
STARTED_SERVICES=()

error_exit() {
  echo ""
  echo "❌ ERROR: $1"
  echo ""
  if [ ${#STARTED_SERVICES[@]} -gt 0 ]; then
    echo "✅ Successfully started:"
    for service in "${STARTED_SERVICES[@]}"; do
      echo "   - $service"
    done
  fi
  echo ""
  echo "🛑 Startup failed at: $2"
  echo ""
  echo "� To clean up, run: ./scripts/stop.sh"
  exit 1
}

echo "�🚀 Starting Raqeem environment..."
echo ""

# Load global .env if present
if [[ -f ./.env ]]; then
  echo "🔧 Loading environment from .env"
  set -a
  # shellcheck disable=SC1091
  source ./.env
  set +a
fi

# Namespace default
NAMESPACE=${NAMESPACE:-default}

# Require ports to be set in .env (no hardcoded defaults)
if [[ -z "${DEVICES_BACKEND_PORT:-}" ]]; then
  error_exit "DEVICES_BACKEND_PORT not set. Define it in .env" "Configuration"
fi
if [[ -z "${MENTOR_BACKEND_PORT:-}" ]]; then
  error_exit "MENTOR_BACKEND_PORT not set. Define it in .env" "Configuration"
fi
if [[ -z "${DEVICES_FRONTEND_PORT:-}" ]]; then
  error_exit "DEVICES_FRONTEND_PORT not set. Define it in .env" "Configuration"
fi
if [[ -z "${MENTOR_FRONTEND_PORT:-}" ]]; then
  error_exit "MENTOR_FRONTEND_PORT not set. Define it in .env" "Configuration"
fi

# Derived frontend API URLs (exported for Vite)
export VITE_DEVICES_API_URL=${VITE_DEVICES_API_URL:-"http://localhost:${DEVICES_BACKEND_PORT}/api/v1"}
export VITE_MENTOR_API_URL=${VITE_MENTOR_API_URL:-"http://localhost:${MENTOR_BACKEND_PORT}"}

# Backend CORS origins (comma-separated list of frontend origins)
export FRONTEND_ORIGIN=${FRONTEND_ORIGIN:-"http://localhost:${MENTOR_FRONTEND_PORT}"}
export FRONTEND_ORIGINS=${FRONTEND_ORIGINS:-"http://localhost:${DEVICES_FRONTEND_PORT}"}
echo ""

# Sanity checks for required commands
for cmd in kubectl helm node npm lsof; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    error_exit "Missing required command: $cmd" "Prerequisites"
  fi
done

# Check cluster connectivity
if ! kubectl cluster-info >/dev/null 2>&1; then
  error_exit "kubectl cannot reach a Kubernetes cluster (is your cluster running?)" "Cluster connectivity"
fi

# Step 1: Deploy PostgreSQL
echo "📦 [1/6] Deploying PostgreSQL..."
if ! helm upgrade --install postgres ./charts/postgres --namespace "$NAMESPACE" --create-namespace; then
  error_exit "Failed to deploy PostgreSQL" "PostgreSQL deployment"
fi
STARTED_SERVICES+=("PostgreSQL")

echo "⏳ Waiting for PostgreSQL pod to be ready..."
if ! kubectl wait --for=condition=ready pod -l app=postgres -n "$NAMESPACE" --timeout=60s; then
  error_exit "PostgreSQL pod failed to become ready" "PostgreSQL readiness check"
fi
echo "✅ PostgreSQL is ready"
echo ""

# Step 2: Deploy MinIO
echo "📦 [2/6] Deploying MinIO..."
if ! helm upgrade --install minio ./charts/minio --namespace "$NAMESPACE"; then
  error_exit "Failed to deploy MinIO" "MinIO deployment"
fi
STARTED_SERVICES+=("MinIO")

echo "⏳ Waiting for MinIO pod to be ready..."
if ! kubectl wait --for=condition=ready pod -l app=minio -n "$NAMESPACE" --timeout=60s; then
  error_exit "MinIO pod failed to become ready" "MinIO readiness check"
fi
echo "✅ MinIO is ready"
echo ""

# Step 3: Deploy Devices Backend
echo "📦 [3/6] Deploying Devices Backend..."
if [[ -f "./.deploy/tag.env" ]]; then
  # shellcheck disable=SC1091
  source ./.deploy/tag.env
  echo "  Using persisted image tag: ${IMAGE_TAG}"
  if ! helm upgrade --install devices-backend ./charts/devices-backend --namespace "$NAMESPACE" \
    --set image.tag="${IMAGE_TAG}" \
    --set frontendOrigins="${FRONTEND_ORIGINS}" \
    --set mentorApiUrl="http://mentor-backend.${NAMESPACE}.svc.cluster.local:8080"; then
    error_exit "Failed to deploy Devices Backend" "Devices Backend deployment"
  fi
else
  if ! helm upgrade --install devices-backend ./charts/devices-backend --namespace "$NAMESPACE" \
    --set frontendOrigins="${FRONTEND_ORIGINS}" \
    --set mentorApiUrl="http://mentor-backend.${NAMESPACE}.svc.cluster.local:8080"; then
    error_exit "Failed to deploy Devices Backend" "Devices Backend deployment"
  fi
fi
STARTED_SERVICES+=("Devices Backend")

echo "⏳ Waiting for Devices Backend pod to be ready..."
if ! kubectl wait --for=condition=ready pod -l app=devices-backend -n "$NAMESPACE" --timeout=60s; then
  error_exit "Devices Backend pod failed to become ready" "Devices Backend readiness check"
fi
echo "✅ Devices Backend is ready"
echo ""

# Step 4: Deploy Mentor Backend
echo "📦 [4/6] Deploying Mentor Backend..."
if [[ -f "./.deploy/tag.env" ]]; then
  # shellcheck disable=SC1091
  source ./.deploy/tag.env
  echo "  Using persisted image tag: ${IMAGE_TAG}"
  if ! helm upgrade --install mentor-backend ./charts/mentor-backend --namespace "$NAMESPACE" \
    --set image.tag="${IMAGE_TAG}" \
    --set frontendOrigin="${FRONTEND_ORIGIN}"; then
    error_exit "Failed to deploy Mentor Backend" "Mentor Backend deployment"
  fi
else
  if ! helm upgrade --install mentor-backend ./charts/mentor-backend --namespace "$NAMESPACE" \
    --set frontendOrigin="${FRONTEND_ORIGIN}"; then
    error_exit "Failed to deploy Mentor Backend" "Mentor Backend deployment"
  fi
fi
STARTED_SERVICES+=("Mentor Backend")

echo "⏳ Waiting for Mentor Backend pod to be ready..."
if ! kubectl wait --for=condition=ready pod -l app=mentor-backend -n "$NAMESPACE" --timeout=60s; then
  error_exit "Mentor Backend pod failed to become ready" "Mentor Backend readiness check"
fi
echo "✅ Mentor Backend is ready"
echo ""

echo "✅ Mentor Backend is ready"
echo ""

# Step 5: Start port-forwards for backends
echo "🔗 [5/6] Starting port-forwards for backends..."
# Kill any existing port-forwards on our ports
lsof -ti :${DEVICES_BACKEND_PORT} | xargs kill -9 2>/dev/null || true
lsof -ti :${MENTOR_BACKEND_PORT} | xargs kill -9 2>/dev/null || true
sleep 1

# Start backend port-forwards in background
kubectl port-forward svc/devices-backend ${DEVICES_BACKEND_PORT}:80 -n "$NAMESPACE" >/dev/null 2>&1 &
PF_DEVICES_PID=$!
kubectl port-forward svc/mentor-backend ${MENTOR_BACKEND_PORT}:80 -n "$NAMESPACE" >/dev/null 2>&1 &
PF_MENTOR_PID=$!

# Wait for port-forwards to be ready and verify
sleep 2
if ! kill -0 $PF_DEVICES_PID 2>/dev/null; then
  error_exit "Devices Backend port-forward failed to start" "Devices Backend port-forward"
fi
if ! kill -0 $PF_MENTOR_PID 2>/dev/null; then
  error_exit "Mentor Backend port-forward failed to start" "Mentor Backend port-forward"
fi
echo "  - Devices Backend PF: http://localhost:${DEVICES_BACKEND_PORT} (PID: $PF_DEVICES_PID)"
echo "  - Mentor Backend PF:  http://localhost:${MENTOR_BACKEND_PORT} (PID: $PF_MENTOR_PID)"
STARTED_SERVICES+=("Backend Port-forwards")
echo "✅ Port-forwards are ready"
echo ""

# Step 6: Start frontends
echo "🌐 [6/6] Starting frontends..."
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Kill any stuck frontend processes on our ports
lsof -ti :${DEVICES_FRONTEND_PORT} | xargs kill -9 2>/dev/null || true
lsof -ti :${MENTOR_FRONTEND_PORT} | xargs kill -9 2>/dev/null || true
pkill -9 -f 'vite.*devices/frontend' 2>/dev/null || true
pkill -9 -f 'vite.*mentor/frontend' 2>/dev/null || true
sleep 2

# Start mentor frontend
echo "  - Starting Mentor Frontend (port ${MENTOR_FRONTEND_PORT})..."
cd "$ROOT_DIR/mentor/frontend"
if ! npm install --silent 2>/dev/null; then
  error_exit "Failed to install Mentor Frontend dependencies" "Mentor Frontend npm install"
fi
VITE_MENTOR_FRONTEND_PORT=${MENTOR_FRONTEND_PORT} VITE_MENTOR_API_URL=${VITE_MENTOR_API_URL} npm run dev >/dev/null 2>&1 &
MENTOR_FE_PID=$!
if ! kill -0 $MENTOR_FE_PID 2>/dev/null; then
  error_exit "Failed to start Mentor Frontend" "Mentor Frontend startup"
fi
echo "    Mentor Frontend PID: $MENTOR_FE_PID"
STARTED_SERVICES+=("Mentor Frontend")

# Start devices frontend
echo "  - Starting Devices Frontend (port ${DEVICES_FRONTEND_PORT})..."
cd "$ROOT_DIR/devices/frontend"
if ! npm install --silent 2>/dev/null; then
  error_exit "Failed to install Devices Frontend dependencies" "Devices Frontend npm install"
fi
VITE_DEVICES_FRONTEND_PORT=${DEVICES_FRONTEND_PORT} VITE_DEVICES_API_URL=${VITE_DEVICES_API_URL} npm run dev >/dev/null 2>&1 &
DEVICES_FE_PID=$!
if ! kill -0 $DEVICES_FE_PID 2>/dev/null; then
  error_exit "Failed to start Devices Frontend" "Devices Frontend startup"
fi
echo "    Devices Frontend PID: $DEVICES_FE_PID"
STARTED_SERVICES+=("Devices Frontend")

cd "$ROOT_DIR"

# Save PIDs for stop script
mkdir -p .deploy
{
  echo "PF_DEVICES_PID=$PF_DEVICES_PID"
  echo "PF_MENTOR_PID=$PF_MENTOR_PID"
  echo "MENTOR_FE_PID=$MENTOR_FE_PID"
  echo "DEVICES_FE_PID=$DEVICES_FE_PID"
} > .deploy/frontend.pids

# Wait for frontends to start
echo "⏳ Waiting for frontends to start..."
sleep 5

# Verify frontends are still running
if ! kill -0 $MENTOR_FE_PID 2>/dev/null; then
  error_exit "Mentor Frontend died after startup" "Mentor Frontend verification"
fi
if ! kill -0 $DEVICES_FE_PID 2>/dev/null; then
  error_exit "Devices Frontend died after startup" "Devices Frontend verification"
fi
echo "✅ Frontends are ready"
echo ""

echo "🎉 Environment is ready!"
echo ""
echo "📱 Access URLs:"
echo "  - Devices Backend:   http://localhost:${DEVICES_BACKEND_PORT}/docs"
echo "  - Mentor Backend:    http://localhost:${MENTOR_BACKEND_PORT}/activities"
echo "  - Mentor Dashboard:  http://localhost:${MENTOR_FRONTEND_PORT}"
echo "  - Device Simulator:  http://localhost:${DEVICES_FRONTEND_PORT}"
echo ""
echo "💡 To stop everything: ./scripts/stop.sh"
