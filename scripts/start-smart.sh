#!/usr/bin/env bash
set -e

# Source service discovery functions
source "$(dirname "$0")/service-discovery.sh"

echo "🚀 Starting Raqeem with Smart Service Discovery..."
echo ""

# Configuration - smart defaults, no .env file needed
NAMESPACE=${NAMESPACE:-default}
DEVICES_FRONTEND_START_PORT=${DEVICES_FRONTEND_START_PORT:-4000}
MENTOR_FRONTEND_START_PORT=${MENTOR_FRONTEND_START_PORT:-5000}

# Clean up any terminated processes first
cleanup_terminated_ports

# Validate tools
for cmd in kubectl helm node npm; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "❌ ERROR: Missing required command: $cmd"
    exit 1
  fi
done

if ! kubectl cluster-info >/dev/null 2>&1; then
  echo "❌ ERROR: kubectl cannot reach Kubernetes cluster"
  exit 1
fi

echo "📦 Deploying backend services with NodePort discovery..."

# Deploy PostgreSQL
helm upgrade --install postgres ./charts/postgres --namespace "$NAMESPACE" --create-namespace
wait_for_service_ready "postgres" "$NAMESPACE"

# Deploy MinIO  
helm upgrade --install minio ./charts/minio --namespace "$NAMESPACE"
wait_for_service_ready "minio" "$NAMESPACE"

# Get actual NodePort assignments (Kubernetes auto-assigns if not specified)
DEVICES_NODEPORT=$(get_nodeport "devices-backend" "$NAMESPACE")
MENTOR_NODEPORT=$(get_nodeport "mentor-backend" "$NAMESPACE")

# If services don't exist yet, deploy them and get ports
if [[ -z "$DEVICES_NODEPORT" ]]; then
  helm upgrade --install devices-backend ./charts/devices-backend --namespace "$NAMESPACE"
  wait_for_service_ready "devices-backend" "$NAMESPACE"
  DEVICES_NODEPORT=$(get_nodeport "devices-backend" "$NAMESPACE")
fi

if [[ -z "$MENTOR_NODEPORT" ]]; then
  helm upgrade --install mentor-backend ./charts/mentor-backend --namespace "$NAMESPACE"
  wait_for_service_ready "mentor-backend" "$NAMESPACE"
  MENTOR_NODEPORT=$(get_nodeport "mentor-backend" "$NAMESPACE")
fi

# Register backend services in discovery registry
register_service "devices-backend" "http://localhost:$DEVICES_NODEPORT" "$DEVICES_NODEPORT"
register_service "mentor-backend" "http://localhost:$MENTOR_NODEPORT" "$MENTOR_NODEPORT"

echo ""
echo "🌐 Starting frontends with auto-detected ports..."

# Find available ports for frontends (avoiding NodePorts)
DEVICES_FRONTEND_PORT=$(find_available_port $DEVICES_FRONTEND_START_PORT)
MENTOR_FRONTEND_PORT=$(find_available_port $MENTOR_FRONTEND_START_PORT)

echo "  - Detected available ports: Devices=$DEVICES_FRONTEND_PORT, Mentor=$MENTOR_FRONTEND_PORT"

# Update CORS settings for backends with actual frontend ports
FRONTEND_ORIGINS="http://localhost:$DEVICES_FRONTEND_PORT\\,http://localhost:$MENTOR_FRONTEND_PORT"
FRONTEND_ORIGIN="http://localhost:$MENTOR_FRONTEND_PORT"

# Redeploy backends with correct CORS settings
helm upgrade devices-backend ./charts/devices-backend --namespace "$NAMESPACE" \
  --set "frontendOrigins=$FRONTEND_ORIGINS" \
  --set "mentorApiUrl=http://mentor-backend.${NAMESPACE}.svc.cluster.local:8080" \
  --reuse-values

helm upgrade mentor-backend ./charts/mentor-backend --namespace "$NAMESPACE" \
  --set "frontendOrigin=$FRONTEND_ORIGIN" \
  --reuse-values

# Start mentor frontend
echo "  - Starting Mentor Frontend on port $MENTOR_FRONTEND_PORT..."
cd mentor/frontend
npm install --silent
VITE_MENTOR_FRONTEND_PORT=$MENTOR_FRONTEND_PORT \
VITE_MENTOR_API_URL="http://localhost:$MENTOR_NODEPORT" \
nohup npm run dev > ../../.deploy/mentor-frontend.log 2>&1 &
MENTOR_FE_PID=$!

# Start devices frontend
echo "  - Starting Devices Frontend on port $DEVICES_FRONTEND_PORT..."
cd ../../devices/frontend
npm install --silent
VITE_DEVICES_FRONTEND_PORT=$DEVICES_FRONTEND_PORT \
VITE_DEVICES_API_URL="http://localhost:$DEVICES_NODEPORT/api/v1" \
nohup npm run dev > ../../.deploy/devices-frontend.log 2>&1 &
DEVICES_FE_PID=$!

cd ../..

# Register frontend services
register_service "devices-frontend" "http://localhost:$DEVICES_FRONTEND_PORT" "$DEVICES_FRONTEND_PORT"
register_service "mentor-frontend" "http://localhost:$MENTOR_FRONTEND_PORT" "$MENTOR_FRONTEND_PORT"

# Save PIDs and ports for cleanup
mkdir -p .deploy
cat > .deploy/smart.pids <<EOF
MENTOR_FE_PID=$MENTOR_FE_PID
DEVICES_FE_PID=$DEVICES_FE_PID
DEVICES_FRONTEND_PORT=$DEVICES_FRONTEND_PORT
MENTOR_FRONTEND_PORT=$MENTOR_FRONTEND_PORT
DEVICES_NODEPORT=$DEVICES_NODEPORT
MENTOR_NODEPORT=$MENTOR_NODEPORT
EOF

# Wait and verify
echo "⏳ Waiting for frontends to initialize..."
sleep 5

if ! kill -0 $MENTOR_FE_PID 2>/dev/null; then
  echo "❌ ERROR: Mentor Frontend failed - check .deploy/mentor-frontend.log"
  exit 1
fi
if ! kill -0 $DEVICES_FE_PID 2>/dev/null; then
  echo "❌ ERROR: Devices Frontend failed - check .deploy/devices-frontend.log"
  exit 1
fi

echo ""
echo "🎉 Smart Service Discovery Ready!"
echo ""
echo "📱 Discovered Services:"
echo "  - Devices Backend:   http://localhost:$DEVICES_NODEPORT/docs"
echo "  - Mentor Backend:    http://localhost:$MENTOR_NODEPORT/health"
echo "  - Mentor Dashboard:  http://localhost:$MENTOR_FRONTEND_PORT"
echo "  - Device Simulator:  http://localhost:$DEVICES_FRONTEND_PORT"
echo ""
echo "🗂️  Service Registry: .deploy/registry/"
echo "💡 To stop: ./scripts/stop-smart.sh"