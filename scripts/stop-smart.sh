#!/usr/bin/env bash
set -e

echo "🛑 Stopping Smart Service Discovery Environment..."
echo ""

NAMESPACE=${NAMESPACE:-default}
CLEAN_DATA=${1:-""}

# Stop frontend processes
  if [[ -f ".deploy/smart.pids" ]]; then
    echo "🌐 Stopping frontend processes..."
  source .deploy/smart.pids
  
    if [[ -n "$MENTOR_FE_PID" ]] && kill -0 "$MENTOR_FE_PID" 2>/dev/null; then
      echo "  - Stopping Dashboard (PID $MENTOR_FE_PID)"
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

# Handle data cleanup based on flag
if [[ "$CLEAN_DATA" == "--clean" ]] || [[ "$CLEAN_DATA" == "-c" ]]; then
  echo ""
  echo "🗑️  Deleting persistent volumes (fresh database on next start)..."
  kubectl delete pvc --all -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ No PVCs to delete"
  echo "✅ Database will be recreated fresh on next start"
else
  echo ""
  echo "✅ Smart Service Discovery Environment Stopped!"
  echo ""
  echo "💾 Data volumes preserved - database will persist on restart"
  echo "💡 To start fresh with clean database, run: ./stop.sh --clean"
fi