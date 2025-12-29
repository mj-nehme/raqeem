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
echo "🛑 Stopping mentor-backend (namespace=$NAMESPACE)..."
helm uninstall mentor-backend -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ mentor-backend not found"

# Clean registry entry if present
if [[ -f "$ROOT_DIR/.deploy/registry/mentor-backend" ]]; then
  rm -f "$ROOT_DIR/.deploy/registry/mentor-backend"
fi

echo "✅ Mentor Backend stopped"
