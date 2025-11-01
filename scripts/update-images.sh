#!/usr/bin/env bash
set -euo pipefail

# update-images.sh
# Build and push backend images, then upgrade Helm releases to use the new tag.
# Usage:
#   ./scripts/update-images.sh [-t TAG] [--only devices|mentor] [--no-build] [--no-push]
#
# Defaults:
#   TAG defaults to current git short SHA or timestamp if git not available.

# Determine TAG
TAG=""
ONLY=""
DO_BUILD=1
DO_PUSH=1

while [[ $# -gt 0 ]]; do
  case "$1" in
    -t|--tag)
      TAG="$2"; shift 2;;
    --only)
      ONLY="$2"; shift 2;;
    --no-build)
      DO_BUILD=0; shift;;
    --no-push)
      DO_PUSH=0; shift;;
    -h|--help)
      sed -n '1,40p' "$0"; exit 0;;
    *)
      echo "Unknown argument: $1" >&2; exit 1;;
  esac
done

if [[ -z "$TAG" ]]; then
  if command -v git >/dev/null 2>&1 && git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    TAG="$(git rev-parse --short HEAD)"
  else
    TAG="v$(date +%Y%m%d%H%M%S)"
  fi
fi

# Read repositories from Helm values
DEVICES_REPO=$(awk -F': ' '/^image:/{f=1} f && /repository:/{print $2; exit}' charts/devices-backend/values.yaml)
MENTOR_REPO=$(awk -F': ' '/^image:/{f=1} f && /repository:/{print $2; exit}' charts/mentor-backend/values.yaml)

if [[ -z "$DEVICES_REPO" || -z "$MENTOR_REPO" ]]; then
  echo "Failed to read image repositories from Helm values." >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

# Build images
if [[ $DO_BUILD -eq 1 ]]; then
  if [[ -z "$ONLY" || "$ONLY" == "devices" ]]; then
    echo "🧱 Building devices backend image: $DEVICES_REPO:$TAG"
    docker build -t "$DEVICES_REPO:$TAG" "$ROOT_DIR/devices/backend"
  fi
  if [[ -z "$ONLY" || "$ONLY" == "mentor" ]]; then
    echo "🧱 Building mentor backend image: $MENTOR_REPO:$TAG"
    docker build -t "$MENTOR_REPO:$TAG" "$ROOT_DIR/mentor/backend"
  fi
fi

# Push images
if [[ $DO_PUSH -eq 1 ]]; then
  if [[ -z "$ONLY" || "$ONLY" == "devices" ]]; then
    echo "🚀 Pushing $DEVICES_REPO:$TAG"
    docker push "$DEVICES_REPO:$TAG"
  fi
  if [[ -z "$ONLY" || "$ONLY" == "mentor" ]]; then
    echo "🚀 Pushing $MENTOR_REPO:$TAG"
    docker push "$MENTOR_REPO:$TAG"
  fi
fi

# Persist tag for future starts
STATE_DIR="$ROOT_DIR/.deploy"
mkdir -p "$STATE_DIR"
echo "IMAGE_TAG=$TAG" > "$STATE_DIR/tag.env"

# Upgrade Helm releases to new tag (immediate rollout)
if [[ -z "$ONLY" || "$ONLY" == "devices" ]]; then
  echo "🔧 Upgrading devices-backend to tag $TAG"
  helm upgrade --install devices-backend ./charts/devices-backend --namespace default --set image.tag="$TAG"
fi
if [[ -z "$ONLY" || "$ONLY" == "mentor" ]]; then
  echo "🔧 Upgrading mentor-backend to tag $TAG"
  helm upgrade --install mentor-backend ./charts/mentor-backend --namespace default --set image.tag="$TAG"
fi

# Show status
echo ""
echo "📊 Deployments status (default namespace):"
kubectl get deploy,po,svc -n default | sed -n '1,100p'

echo ""
echo "✅ Done. Image tag: $TAG"
echo "ℹ️  Next: ./scripts/stop.sh && ./scripts/start.sh (if you want a clean restart)"
