#!/usr/bin/env bash
# Service Discovery Client - Query services dynamically

REGISTRY_DIR=".deploy/registry"

get_service() {
  local service_name=$1
  local property=${2:-"URL"}
  
  if [[ -f "$REGISTRY_DIR/$service_name" ]]; then
    source "$REGISTRY_DIR/$service_name"
    case $property in
      "URL"|"url") echo "$URL" ;;
      "PORT"|"port") echo "$PORT" ;;
      "STATUS"|"status") echo "$STATUS" ;;
      "TIMESTAMP"|"timestamp") echo "$TIMESTAMP" ;;
      *) echo "Unknown property: $property" >&2; return 1 ;;
    esac
  else
    echo "Service not found: $service_name" >&2
    return 1
  fi
}

list_services() {
  echo "📋 Registered Services:"
  if [[ -d "$REGISTRY_DIR" ]]; then
    for service_file in "$REGISTRY_DIR"/*; do
      if [[ -f "$service_file" ]]; then
        local service_name=$(basename "$service_file")
        local url=$(get_service "$service_name" "URL")
        local status=$(get_service "$service_name" "STATUS")
        printf "  %-20s %s [%s]\n" "$service_name" "$url" "$status"
      fi
    done
  else
    echo "  No services registered"
  fi
}

check_service_health() {
  local service_name=$1
  local url=$(get_service "$service_name" "URL" 2>/dev/null)
  
  if [[ -n "$url" ]]; then
    if curl -sf "$url/health" >/dev/null 2>&1; then
      echo "✅ $service_name is healthy"
      return 0
    else
      echo "❌ $service_name is not responding"
      return 1
    fi
  else
    echo "❓ $service_name not registered"
    return 1
  fi
}

# CLI interface
case "${1:-list}" in
  "get")
    get_service "$2" "$3"
    ;;
  "list")
    list_services
    ;;
  "health")
    if [[ -n "$2" ]]; then
      check_service_health "$2"
    else
      echo "Checking all services..."
      for service_file in "$REGISTRY_DIR"/*; do
        if [[ -f "$service_file" ]]; then
          service_name=$(basename "$service_file")
          check_service_health "$service_name"
        fi
      done
    fi
    ;;
  "wait")
    service_name="$2"
    timeout="${3:-30}"
    echo "⏳ Waiting for $service_name to be available..."
    for ((i=0; i<timeout; i++)); do
      if check_service_health "$service_name" >/dev/null 2>&1; then
        echo "✅ $service_name is ready"
        exit 0
      fi
      sleep 1
    done
    echo "❌ Timeout waiting for $service_name"
    exit 1
    ;;
  *)
    echo "Usage: $0 {list|get <service> [property]|health [service]|wait <service> [timeout]}"
    echo ""
    echo "Examples:"
    echo "  $0 list                           # List all services"
    echo "  $0 get devices-backend           # Get service URL"
    echo "  $0 get devices-backend PORT      # Get service port"
    echo "  $0 health                        # Check all services"
    echo "  $0 health mentor-backend         # Check specific service"
    echo "  $0 wait devices-backend 60       # Wait for service to be ready"
    exit 1
    ;;
esac