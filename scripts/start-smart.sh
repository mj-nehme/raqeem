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

# Check and optionally pre-pull Docker images
check_and_pull_images

# Build local backend images
echo "🔨 Building local backend images..."
"$(dirname "$0")/build-local-images.sh"
echo ""

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
echo "💡 Note: First run after clearing Docker images may take 5-10 minutes for image downloads"

# Deploy PostgreSQL
echo "🐘 Deploying PostgreSQL..."
helm upgrade --install postgres ./charts/postgres --namespace "$NAMESPACE" --create-namespace
wait_for_service_ready "postgres" "$NAMESPACE" 300  # 5 minutes for fresh image pulls

# Deploy MinIO  
echo "🗄️ Deploying MinIO..."
helm upgrade --install minio ./charts/minio --namespace "$NAMESPACE"
wait_for_service_ready "minio" "$NAMESPACE" 300  # 5 minutes for fresh image pulls

echo ""
echo "🌐 Detecting available frontend ports..."

# Find available ports for frontends (avoiding NodePorts)
# Search within 5-port range for each application
DEVICES_FRONTEND_PORT=$(find_available_port $DEVICES_FRONTEND_START_PORT 5)
if [[ $? -ne 0 ]]; then
  echo "❌ ERROR: Could not find available port for Devices Frontend in range $DEVICES_FRONTEND_START_PORT-$((DEVICES_FRONTEND_START_PORT + 4))"
  exit 1
fi

MENTOR_FRONTEND_PORT=$(find_available_port $MENTOR_FRONTEND_START_PORT 5)
if [[ $? -ne 0 ]]; then
  echo "❌ ERROR: Could not find available port for Mentor Frontend in range $MENTOR_FRONTEND_START_PORT-$((MENTOR_FRONTEND_START_PORT + 4))"
  exit 1
fi

echo "  - Detected available ports: Devices=$DEVICES_FRONTEND_PORT, Mentor=$MENTOR_FRONTEND_PORT"

# Build CORS regex patterns to allow any port in the 5-port range for each frontend
# This solves the CORS problem by allowing dynamic ports within a limited range
DEVICES_FRONTEND_CORS_REGEX="^http://localhost:([4-9][0-9]{3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$"
MENTOR_FRONTEND_CORS_REGEX="^http://localhost:([4-9][0-9]{3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$"

# For more specific CORS (only the 5-port ranges we're using):
# Devices frontend: ports 4000-4004
# Mentor frontend: ports 5000-5004
DEVICES_RANGE_START=$DEVICES_FRONTEND_START_PORT
DEVICES_RANGE_END=$((DEVICES_FRONTEND_START_PORT + 4))
MENTOR_RANGE_START=$MENTOR_FRONTEND_START_PORT
MENTOR_RANGE_END=$((MENTOR_FRONTEND_START_PORT + 4))

# Build regex to match both frontend port ranges
# Regex: ^http://localhost:(4000|4001|4002|4003|4004|5000|5001|5002|5003|5004)$
CORS_REGEX="^http://localhost:($(seq -s'|' $DEVICES_RANGE_START $DEVICES_RANGE_END)|$(seq -s'|' $MENTOR_RANGE_START $MENTOR_RANGE_END))\$"

echo "  - CORS Regex Pattern: $CORS_REGEX"
echo "  - This allows ports: $DEVICES_RANGE_START-$DEVICES_RANGE_END (Devices) and $MENTOR_RANGE_START-$MENTOR_RANGE_END (Mentor)"

# Check if backend ports are available and find alternatives if needed
echo ""
echo "🔍 Checking backend port availability..."

DEVICES_BACKEND_PORT=$(find_available_backend_port 30080 5)
if [[ $? -ne 0 ]]; then
  echo "❌ ERROR: Could not find available port for Devices Backend in range 30080-30084"
  exit 1
fi

MENTOR_BACKEND_PORT=$(find_available_backend_port 30090 5)
if [[ $? -ne 0 ]]; then
  echo "❌ ERROR: Could not find available port for Mentor Backend in range 30090-30094"
  exit 1
fi

echo "  - Backend ports: Devices=$DEVICES_BACKEND_PORT, Mentor=$MENTOR_BACKEND_PORT"

# Get actual NodePort assignments (Kubernetes auto-assigns if not specified)
DEVICES_NODEPORT=$(get_nodeport "devices-backend" "$NAMESPACE")
MENTOR_NODEPORT=$(get_nodeport "mentor-backend" "$NAMESPACE")

# Deploy mentor backend first to get its URL for devices backend
if [[ -z "$MENTOR_NODEPORT" ]]; then
  helm upgrade --install mentor-backend ./charts/mentor-backend \
    --namespace "$NAMESPACE" \
    --set service.nodePort="$MENTOR_BACKEND_PORT" \
    --set-string frontendOriginRegex="$CORS_REGEX"
  wait_for_service_ready "mentor-backend" "$NAMESPACE"
  MENTOR_NODEPORT=$(get_nodeport "mentor-backend" "$NAMESPACE")
fi

# Set mentor API URL for devices backend
MENTOR_API_URL="http://localhost:$MENTOR_NODEPORT"

# Deploy devices backend with MENTOR_API_URL and CORS regex
if [[ -z "$DEVICES_NODEPORT" ]]; then
  helm upgrade --install devices-backend ./charts/devices-backend \
    --namespace "$NAMESPACE" \
    --set service.nodePort="$DEVICES_BACKEND_PORT" \
    --set-string mentorApiUrl="$MENTOR_API_URL" \
    --set-string frontendOriginRegex="$CORS_REGEX"
  wait_for_service_ready "devices-backend" "$NAMESPACE"
  DEVICES_NODEPORT=$(get_nodeport "devices-backend" "$NAMESPACE")
else
  # If devices-backend already exists, upgrade it with mentor URL and CORS
  helm upgrade --install devices-backend ./charts/devices-backend \
    --namespace "$NAMESPACE" \
    --set service.nodePort="$DEVICES_BACKEND_PORT" \
    --set-string mentorApiUrl="$MENTOR_API_URL" \
    --set-string frontendOriginRegex="$CORS_REGEX"
  wait_for_service_ready "devices-backend" "$NAMESPACE"
fi

# Update mentor backend with CORS regex if it was already deployed
if [[ -n "$MENTOR_NODEPORT" ]]; then
  helm upgrade --install mentor-backend ./charts/mentor-backend \
    --namespace "$NAMESPACE" \
    --set service.nodePort="$MENTOR_BACKEND_PORT" \
    --set-string frontendOriginRegex="$CORS_REGEX"
  wait_for_service_ready "mentor-backend" "$NAMESPACE"
fi

# Register backend services in discovery registry
register_service "devices-backend" "http://localhost:$DEVICES_NODEPORT" "$DEVICES_NODEPORT"
register_service "mentor-backend" "http://localhost:$MENTOR_NODEPORT" "$MENTOR_NODEPORT"

echo ""
echo "🌐 Starting frontends with detected ports..."

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
echo "🔐 CORS Configuration:"
echo "  - Both backends use CORS regex: $CORS_REGEX"
echo "  - Allowed port ranges: $DEVICES_RANGE_START-$DEVICES_RANGE_END (Devices) and $MENTOR_RANGE_START-$MENTOR_RANGE_END (Mentor)"
echo ""
echo "📂 Service Registry: .deploy/registry/"
echo "💡 To stop: ./scripts/stop-smart.sh"