#!/usr/bin/env bash
set -e

NAMESPACE=${NAMESPACE:-default}
CLEAN_DATA=${1:-""}

echo "🛑 Stopping Infra (PostgreSQL + MinIO) in namespace '$NAMESPACE'..."

helm uninstall minio -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ minio not found"
helm uninstall postgres -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ postgres not found"

if [[ "$CLEAN_DATA" == "--clean" ]] || [[ "$CLEAN_DATA" == "-c" ]]; then
  echo "🗑️  Deleting persistent volumes (fresh database on next start)..."
  kubectl delete pvc --all -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ No PVCs to delete"
  echo "⏳ Waiting for PVCs to be fully removed..."
  for i in $(seq 1 30); do
    REMAINING=$(kubectl get pvc -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l | tr -d ' ')
    if [[ "$REMAINING" == "0" ]]; then
      break
    fi
    sleep 2
  done
  if kubectl get pvc postgres-pvc -n "$NAMESPACE" >/dev/null 2>&1; then
    echo "⚠️  postgres-pvc still present; attempting to remove finalizers and force delete..."
    kubectl patch pvc postgres-pvc -n "$NAMESPACE" -p '{"metadata":{"finalizers":null}}' --type=merge >/dev/null 2>&1 || true
    kubectl delete pvc postgres-pvc -n "$NAMESPACE" --grace-period=0 >/dev/null 2>&1 || true
  fi
  echo "✅ Database volumes cleaned"
fi

echo "✅ Infra stopped"