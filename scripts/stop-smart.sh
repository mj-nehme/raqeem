#!/usr/bin/env bash
set -e

echo "🛑 Stopping Smart Service Discovery Environment..."
echo ""

NAMESPACE=${NAMESPACE:-default}

# Stop frontend processes
if [[ -f ".deploy/smart.pids" ]]; then
  echo "🌐 Stopping frontend processes..."
  source .deploy/smart.pids
  
  if [[ -n "$MENTOR_FE_PID" ]] && kill -0 "$MENTOR_FE_PID" 2>/dev/null; then
    echo "  - Stopping Mentor Frontend (PID $MENTOR_FE_PID)"
    kill "$MENTOR_FE_PID" 2>/dev/null || true
  fi
  
  if [[ -n "$DEVICES_FE_PID" ]] && kill -0 "$DEVICES_FE_PID" 2>/dev/null; then
    echo "  - Stopping Devices Frontend (PID $DEVICES_FE_PID)"
    kill "$DEVICES_FE_PID" 2>/dev/null || true
  fi
  
  rm -f .deploy/smart.pids
fi

# Clean up service registry
if [[ -d ".deploy/registry" ]]; then
  echo "🗂️  Cleaning up service registry..."
  rm -rf .deploy/registry
fi

# Uninstall Helm releases
echo "📦 Uninstalling Kubernetes services..."
helm uninstall devices-backend -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ devices-backend not found"
helm uninstall mentor-backend -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ mentor-backend not found"
helm uninstall minio -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ minio not found"
helm uninstall postgres -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ postgres not found"

echo ""
echo "✅ Smart Service Discovery Environment Stopped!"
echo ""
echo "💾 Note: Data volumes preserved"
echo "   To delete all data: kubectl delete pvc --all -n $NAMESPACE"