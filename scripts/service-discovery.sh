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
  local timeout=${3:-60}
  
  echo "⏳ Waiting for $service_name to be ready..."
  if kubectl wait --for=condition=ready pod -l app=$service_name -n "$namespace" --timeout=${timeout}s; then
    echo "✅ $service_name is ready"
    return 0
  else
    echo "❌ $service_name failed to become ready within ${timeout}s"
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

export -f find_available_port wait_for_service_ready get_nodeport register_service get_service_url cleanup_terminated_ports