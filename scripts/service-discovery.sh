#!/usr/bin/env bash
# Smart Service Discovery - Auto-detect available ports and register services

find_available_port() {
  local start_port=$1
  local max_attempts=${2:-50}
  
  for ((i=0; i<max_attempts; i++)); do
    local port=$((start_port + i))
    if ! lsof -i ":$port" >/dev/null 2>&1; then
      echo $port
      return 0
    fi
  done
  
  echo "ERROR: No available port found starting from $start_port" >&2
  return 1
}

wait_for_service_ready() {
  local service_name=$1
  local namespace=$2
  local timeout=${3:-300}  # Increased default timeout for image pulling
  
  echo "⏳ Waiting for $service_name to be ready..."
  
  # First, check if we need to pull images (longer timeout needed)
  local pod_name=$(kubectl get pods -l app=$service_name -n "$namespace" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
  
  if [[ -n "$pod_name" ]]; then
    # Check if pod is in image pulling state
    local pod_status=$(kubectl get pod "$pod_name" -n "$namespace" -o jsonpath='{.status.containerStatuses[0].state.waiting.reason}' 2>/dev/null || echo "")
    
    if [[ "$pod_status" == "ContainerCreating" ]]; then
      echo "  📦 Detecting if images need to be pulled..."
      
      # Check if we're pulling images
      local events=$(kubectl get events --field-selector involvedObject.name="$pod_name" -n "$namespace" --sort-by='.lastTimestamp' -o custom-columns=TYPE:.type,REASON:.reason,MESSAGE:.message --no-headers 2>/dev/null || echo "")
      
      if echo "$events" | grep -q "Pulling"; then
        echo "  ⬇️ Images are being pulled from Docker Hub (this may take several minutes for fresh installations)..."
        echo "  💡 Tip: Pre-pull images with 'docker pull postgres:15' and 'docker pull minio/minio:RELEASE.2023-09-04T19-57-37Z' to speed up future starts"
      fi
    fi
  fi
  
  if kubectl wait --for=condition=ready pod -l app=$service_name -n "$namespace" --timeout=${timeout}s; then
    echo "✅ $service_name is ready"
    return 0
  else
    echo "❌ $service_name failed to become ready within ${timeout}s"
    
    # Provide helpful debugging info
    if [[ -n "$pod_name" ]]; then
      echo "🔍 Debug info for $service_name:"
      kubectl get pod "$pod_name" -n "$namespace" 2>/dev/null || echo "  - Pod not found"
      echo "  Last events:"
      kubectl get events --field-selector involvedObject.name="$pod_name" -n "$namespace" --sort-by='.lastTimestamp' -o custom-columns=MESSAGE:.message --no-headers 2>/dev/null | tail -3 | sed 's/^/    /'
    fi
    return 1
  fi
}

get_nodeport() {
  local service_name=$1
  local namespace=$2
  
  kubectl get svc "$service_name" -n "$namespace" -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo ""
}

register_service() {
  local service_name=$1
  local url=$2
  local port=$3
  
  mkdir -p .deploy/registry
  cat > ".deploy/registry/$service_name" <<EOF
URL=$url
PORT=$port
STATUS=ready
TIMESTAMP=$(date +%s)
EOF
  echo "📝 Registered $service_name at $url"
}

get_service_url() {
  local service_name=$1
  
  if [[ -f ".deploy/registry/$service_name" ]]; then
    source ".deploy/registry/$service_name"
    echo "$URL"
  else
    echo ""
  fi
}

cleanup_terminated_ports() {
  echo "🧹 Cleaning up terminated processes on common ports..."
  
  # Find and kill any processes in terminating state on our port ranges
  for port in $(seq 3000 3010) $(seq 4000 4010) $(seq 5000 5010) $(seq 8080 8090); do
    local pids=$(lsof -ti ":$port" 2>/dev/null || true)
    if [[ -n "$pids" ]]; then
      echo "  - Checking port $port..."
      for pid in $pids; do
        local status=$(ps -o stat= -p "$pid" 2>/dev/null || echo "")
        if [[ "$status" =~ [TZ] ]]; then
          echo "  - Killing terminated process $pid on port $port"
          kill -9 "$pid" 2>/dev/null || true
        fi
      done
    fi
  done
  
  # Wait for cleanup
  sleep 2
}

check_and_pull_images() {
  local required_images=("postgres:15" "minio/minio:RELEASE.2023-09-04T19-57-37Z")
  local missing_images=()
  
  echo "🔍 Checking required Docker images..."
  
  for image in "${required_images[@]}"; do
    if ! docker image inspect "$image" >/dev/null 2>&1; then
      missing_images+=("$image")
    fi
  done
  
  if [[ ${#missing_images[@]} -gt 0 ]]; then
    echo "📦 Missing images detected: ${missing_images[*]}"
    echo "⚡ Pre-pulling images to speed up deployment..."
    
    for image in "${missing_images[@]}"; do
      echo "  ⬇️ Pulling $image..."
      if docker pull "$image"; then
        echo "    ✅ $image pulled successfully"
      else
        echo "    ⚠️ Failed to pull $image - will retry during Kubernetes deployment"
      fi
    done
  else
    echo "✅ All required images are available locally"
  fi
}

export -f check_and_pull_images