#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(dirname "$0")"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

source "$SCRIPT_DIR/service-discovery.sh"
source "$SCRIPT_DIR/preflight.sh"

TARGET=${1:-all}

echo "🚀 Starting Infra deployment (target=$TARGET)..."

NAMESPACE=${NAMESPACE:-default}

# Validate tools and services
ensure_docker_running || exit 1
ensure_helm_ready || exit 1
ensure_kube_ready || exit 1

# Pre-pull images to speed up deployment
check_and_pull_images

deploy_postgres() {
  echo "🐘 Deploying PostgreSQL to namespace '$NAMESPACE'..."
# Preflight: check for existing PVC without Helm ownership
if kubectl get pvc postgres-pvc -n "$NAMESPACE" >/dev/null 2>&1; then
  OWNED_BY_HELM=$(kubectl get pvc postgres-pvc -n "$NAMESPACE" -o jsonpath='{.metadata.annotations.meta\.helm\.sh/release-name}' 2>/dev/null || echo "")
  if [[ -z "$OWNED_BY_HELM" ]]; then
    echo "⚠️  Found existing PVC 'postgres-pvc' without Helm ownership."
    echo "   Helm cannot adopt existing PVCs. Either:"
    echo "   - Run ./scripts/stop-infra.sh --clean to delete PVCs, or"
    echo "   - kubectl delete pvc postgres-pvc -n $NAMESPACE"
    if [[ "${AUTO_CLEAN_PVC:-}" == "true" ]]; then
      echo "   AUTO_CLEAN_PVC=true -> deleting PVC automatically..."
      kubectl delete pvc postgres-pvc -n "$NAMESPACE" || true
    else
      echo "   Aborting to prevent partial install. Set AUTO_CLEAN_PVC=true to auto-delete."
      exit 1
    fi
  fi
fi

  helm upgrade --install postgres "$ROOT_DIR/charts/postgres" --namespace "$NAMESPACE" --create-namespace
  wait_for_service_ready "postgres" "$NAMESPACE" 300
}

deploy_minio() {
  echo "🗄️ Deploying MinIO to namespace '$NAMESPACE'..."
MINIO_IMAGE_REPO_ENV=${MINIO_IMAGE_REPO:-}
MINIO_IMAGE_TAG_ENV=${MINIO_IMAGE_TAG:-}
if [[ -n "$MINIO_IMAGE_REPO_ENV" || -n "$MINIO_IMAGE_TAG_ENV" ]]; then
  echo "  - Using image override: repo='${MINIO_IMAGE_REPO_ENV:-<chart default>}' tag='${MINIO_IMAGE_TAG_ENV:-<chart default>}" 
  helm upgrade --install minio "$ROOT_DIR/charts/minio" \
    --namespace "$NAMESPACE" \
    --set image.repository="${MINIO_IMAGE_REPO_ENV:-quay.io/minio/minio}" \
    --set image.tag="${MINIO_IMAGE_TAG_ENV:-latest}"
else
  helm upgrade --install minio "$ROOT_DIR/charts/minio" --namespace "$NAMESPACE"
fi
  wait_for_service_ready "minio" "$NAMESPACE" 300
}

case "$TARGET" in
  all)
    # Postgres
    # Preflight: check for existing PVC without Helm ownership
    if kubectl get pvc postgres-pvc -n "$NAMESPACE" >/dev/null 2>&1; then
      OWNED_BY_HELM=$(kubectl get pvc postgres-pvc -n "$NAMESPACE" -o jsonpath='{.metadata.annotations.meta\.helm\.sh/release-name}' 2>/dev/null || echo "")
      if [[ -z "$OWNED_BY_HELM" ]]; then
        echo "⚠️  Found existing PVC 'postgres-pvc' without Helm ownership."
        echo "   Helm cannot adopt existing PVCs. Either:"
        echo "   - Run ./scripts/stop-infra.sh --clean to delete PVCs, or"
        echo "   - kubectl delete pvc postgres-pvc -n $NAMESPACE"
        if [[ "${AUTO_CLEAN_PVC:-}" == "true" ]]; then
          echo "   AUTO_CLEAN_PVC=true -> deleting PVC automatically..."
          kubectl patch pvc postgres-pvc -n "$NAMESPACE" -p '{"metadata":{"finalizers":null}}' --type=merge >/dev/null 2>&1 || true
          kubectl delete pvc postgres-pvc -n "$NAMESPACE" --grace-period=0 || true
        else
          echo "   Aborting to prevent partial install. Set AUTO_CLEAN_PVC=true to auto-delete."
          exit 1
        fi
      fi
    fi
    deploy_postgres
    deploy_minio
    ;;
  postgres)
    # Preflight for Postgres PVC
    if kubectl get pvc postgres-pvc -n "$NAMESPACE" >/dev/null 2>&1; then
      OWNED_BY_HELM=$(kubectl get pvc postgres-pvc -n "$NAMESPACE" -o jsonpath='{.metadata.annotations.meta\.helm\.sh/release-name}' 2>/dev/null || echo "")
      if [[ -z "$OWNED_BY_HELM" ]]; then
        echo "⚠️  Found existing PVC 'postgres-pvc' without Helm ownership."
        if [[ "${AUTO_CLEAN_PVC:-}" == "true" ]]; then
          echo "   AUTO_CLEAN_PVC=true -> deleting PVC automatically..."
          kubectl patch pvc postgres-pvc -n "$NAMESPACE" -p '{"metadata":{"finalizers":null}}' --type=merge >/dev/null 2>&1 || true
          kubectl delete pvc postgres-pvc -n "$NAMESPACE" --grace-period=0 || true
        else
          echo "   Aborting. Set AUTO_CLEAN_PVC=true or run stop-infra.sh --clean"
          exit 1
        fi
      fi
    fi
    deploy_postgres
    ;;
  minio)
    deploy_minio
    ;;
  *)
    echo "❌ Unknown target: $TARGET"
    echo "Valid targets: all | postgres | minio"
    exit 1
    ;;
esac

echo "✅ Infra ready: PostgreSQL and MinIO are running in namespace '$NAMESPACE'"
echo "💡 You can skip this locally and deploy these on AWS when desired."