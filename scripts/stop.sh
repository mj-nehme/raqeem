#!/usr/bin/env bash
set -e

echo "🛑 Stopping Raqeem environment..."
echo ""

# Load global .env if present
if [[ -f ./.env ]]; then
  set -a; source ./.env; set +a
fi
NAMESPACE=${NAMESPACE:-default}

# Uninstall all Helm releases
echo "📦 Uninstalling Devices Backend..."
helm uninstall devices-backend -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️  devices-backend not found"

echo "📦 Uninstalling Mentor Backend..."
helm uninstall mentor-backend -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️  mentor-backend not found"

echo "📦 Uninstalling MinIO..."
helm uninstall minio -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️  minio not found"

echo "📦 Uninstalling PostgreSQL..."
helm uninstall postgres -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️  postgres not found"

echo ""
echo "⏳ Waiting for pods to terminate..."
sleep 3

echo ""
echo "📊 Remaining resources:"
kubectl get pods,svc -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️  No resources found"

echo ""
echo "💾 Note: PersistentVolumeClaims are NOT deleted (data preserved)"
echo "   To delete data volumes, run: kubectl delete pvc --all -n $NAMESPACE"
echo ""

# Stop frontends and port-forwards if running
if [[ -f ".deploy/frontend.pids" ]]; then
  echo "🌐 Stopping frontends and port-forwards..."
  while IFS='=' read -r key pid; do
    if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
      echo "  - Stopping $key (PID $pid)"
      kill "$pid" 2>/dev/null || true
    fi
  done < .deploy/frontend.pids
  rm -f .deploy/frontend.pids
  echo ""
fi

echo "✅ Environment stopped!"
