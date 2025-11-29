#!/usr/bin/env bash
set -e

# =============================================================================
# RELEASE SCRIPT - Creates and publishes official releases
# =============================================================================
# This script is for maintainers to create official releases.
# It builds Docker images and pushes them to GitHub Container Registry (GHCR).
#
# For local development, use:
#   ./scripts/build-local-images.sh  # Build images locally
#   ./start.sh                       # Start local Kubernetes cluster
#
# Prerequisites for this script:
#   - GitHub authentication (gh CLI or Personal Access Token)
#   - Docker login to ghcr.io
#   - Write access to the repository packages
# =============================================================================

# Color codes for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_info() {
  echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
  echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
  echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
  echo -e "${RED}❌ $1${NC}"
}

# Check if we're logged in to GHCR already
docker_logged_in_to_ghcr() {
  # Docker may not create config.json until the first login; treat as not logged in
  [[ -f "$HOME/.docker/config.json" ]] || return 1
  grep -q '"ghcr.io"' "$HOME/.docker/config.json" 2>/dev/null
}

# Ensure we are authenticated to GHCR, attempt non-interactive login if possible
ensure_ghcr_login() {
  if docker_logged_in_to_ghcr; then
    return 0
  fi

  # Non-interactive credential discovery: env vars first, then gh CLI token
  local USERNAME TOKEN
  USERNAME="${GHCR_USERNAME:-${GH_USERNAME:-${GITHUB_ACTOR}}}"
  TOKEN="${GHCR_PAT:-${GH_TOKEN:-${GITHUB_TOKEN}}}"

  if [[ -z "$USERNAME" ]] && command -v gh >/dev/null 2>&1; then
    USERNAME=$(gh api user -q .login 2>/dev/null || true)
  fi
  if [[ -z "$TOKEN" ]] && command -v gh >/dev/null 2>&1; then
    TOKEN=$(gh auth token 2>/dev/null || true)
  fi

  if [[ -n "$USERNAME" && -n "$TOKEN" ]]; then
    print_info "Attempting docker login to ghcr.io as '$USERNAME'..."
    if echo "$TOKEN" | docker login ghcr.io -u "$USERNAME" --password-stdin; then
      print_success "Authenticated to GHCR"
      return 0
    else
      print_error "Docker login to GHCR failed"
      return 1
    fi
  fi

  print_error "No GHCR credentials found. Set GHCR_USERNAME and GHCR_PAT, or login via 'gh auth login' and retry."
  return 1
}

# Force interactive GHCR login (replaces any existing login)
prompt_ghcr_login() {
  print_info "Re-authenticating with GHCR (requires write:packages scope)"
  docker logout ghcr.io >/dev/null 2>&1 || true
  if [[ -t 0 ]]; then
    local USERNAME TOKEN
    read -rp "GitHub username: " USERNAME
    read -rsp "GitHub token (write:packages): " TOKEN
    echo ""
    if [[ -n "$USERNAME" && -n "$TOKEN" ]]; then
      if echo "$TOKEN" | docker login ghcr.io -u "$USERNAME" --password-stdin; then
        print_success "Authenticated to GHCR"
        return 0
      else
        print_error "Docker login to GHCR failed"
        return 1
      fi
    else
      print_error "Username/token not provided"
      return 1
    fi
  else
    print_error "Cannot prompt for credentials (no TTY). Set GHCR_PAT and GHCR_USERNAME env vars."
    return 1
  fi
}

# Push with retry on authentication scope errors
push_image() {
  local IMAGE="$1"
  if docker push "$IMAGE"; then
    return 0
  fi
  print_warning "Push failed for $IMAGE. Attempting to re-authenticate to GHCR..."
  if prompt_ghcr_login; then
    if docker push "$IMAGE"; then
      return 0
    fi
  fi
  print_error "Failed to push $IMAGE after re-authentication"
  return 1
}

# Parse arguments
if [[ -z "$1" ]]; then
  print_error "Usage: $0 <version> [--skip-tests] [--yes|-y] [--non-interactive]"
  echo ""
  echo "Examples:"
  echo "  $0 v1.0.0"
  echo "  $0 v1.1.0 --skip-tests --yes --non-interactive"
  echo ""
  exit 1
fi

VERSION=$1; shift

SKIP_TESTS=""
AUTO_CONFIRM_PUSH=1
NON_INTERACTIVE=1

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-tests)
      SKIP_TESTS="--skip-tests";
      shift
      ;;
    --yes|-y)
      AUTO_CONFIRM_PUSH=1;
      NON_INTERACTIVE=1; # imply non-interactive when auto-confirming
      shift
      ;;
    --non-interactive)
      NON_INTERACTIVE=1;
      shift
      ;;
    *)
      print_warning "Unknown argument: $1"; shift
      ;;
  esac
done

# Validate version format (vX.Y.Z)
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  print_error "Version must be in format vX.Y.Z (e.g., v1.0.0)"
  exit 1
fi

print_info "Creating release $VERSION"
echo ""

# Check for uncommitted changes
if [[ -n $(git status --porcelain) ]]; then
  print_warning "You have uncommitted changes. Commit them first!"
  git status --short
  exit 1
fi

# Get current git commit
GIT_COMMIT=$(git rev-parse --short HEAD)
print_info "Current commit: $GIT_COMMIT"

# Run tests unless --skip-tests is specified
if [[ "$SKIP_TESTS" != "--skip-tests" ]]; then
  print_info "Running validation tests..."
  
  # Test Go backend compilation
  print_info "  Testing Mentor Backend (Go)..."
  if ! (cd mentor/backend/src && go build -o /dev/null .); then
    print_error "Mentor Backend compilation failed"
    exit 1
  fi
  
  # Test Python backend syntax
  print_info "  Testing Devices Backend (Python)..."
  if ! python3 -m py_compile devices/backend/src/app/main.py; then
    print_error "Devices Backend has syntax errors"
    exit 1
  fi
  
    # Test Helm charts (if Helm is installed)
    if command -v helm >/dev/null 2>&1; then
      print_info "  Validating Helm charts..."
      for chart in charts/*/; do
        chart_name=$(basename "$chart")
        if ! helm lint "$chart" > /dev/null 2>&1; then
          print_error "Helm chart validation failed: $chart_name"
          exit 1
        fi
      done
    else
      print_warning "Helm not found; skipping 'helm lint' validation"
    fi
  
  print_success "All validation tests passed"
  echo ""
fi

# Build Docker images with version tags
print_info "Building Docker images..."
echo ""

print_info "  Building Devices Backend..."
docker build -t ghcr.io/mj-nehme/raqeem-devices-backend:${VERSION} \
             -t ghcr.io/mj-nehme/raqeem-devices-backend:${VERSION}-${GIT_COMMIT} \
             -t ghcr.io/mj-nehme/raqeem-devices-backend:latest \
             devices/backend/

print_info "  Building Mentor Backend..."
docker build -t ghcr.io/mj-nehme/raqeem-mentor-backend:${VERSION} \
             -t ghcr.io/mj-nehme/raqeem-mentor-backend:${VERSION}-${GIT_COMMIT} \
             -t ghcr.io/mj-nehme/raqeem-mentor-backend:latest \
             mentor/backend/

print_success "Images built successfully"
echo ""

# Check or establish Docker authentication before pushing
print_info "Ensuring GitHub Container Registry authentication..."
if ! ensure_ghcr_login; then
  print_error "Not authenticated with GitHub Container Registry (ghcr.io)"
  exit 1
fi
echo ""

# Confirmation prompt before pushing to registry
print_info "Pushing Docker images to GitHub Container Registry (no prompt)"

# Push images to registry
print_info "Pushing images to GitHub Container Registry..."
echo ""

print_info "  Pushing Devices Backend images..."
push_image ghcr.io/mj-nehme/raqeem-devices-backend:${VERSION}
push_image ghcr.io/mj-nehme/raqeem-devices-backend:${VERSION}-${GIT_COMMIT}
push_image ghcr.io/mj-nehme/raqeem-devices-backend:latest

print_info "  Pushing Mentor Backend images..."
push_image ghcr.io/mj-nehme/raqeem-mentor-backend:${VERSION}
push_image ghcr.io/mj-nehme/raqeem-mentor-backend:${VERSION}-${GIT_COMMIT}
push_image ghcr.io/mj-nehme/raqeem-mentor-backend:latest

print_success "Images pushed successfully"
echo ""

# Update Helm chart values to use the new version
print_info "Updating Helm chart values..."

# Update devices-backend values.yaml
sed -i.bak "s/tag: .*/tag: ${VERSION}/" charts/devices-backend/values.yaml
rm -f charts/devices-backend/values.yaml.bak

# Update mentor-backend values.yaml
sed -i.bak "s/tag: .*/tag: ${VERSION}/" charts/mentor-backend/values.yaml
rm -f charts/mentor-backend/values.yaml.bak

print_success "Helm charts updated"
echo ""

# Persist the version tag for start.sh to use
mkdir -p .deploy
echo "IMAGE_TAG=${VERSION}" > .deploy/tag.env
echo "GIT_COMMIT=${GIT_COMMIT}" >> .deploy/tag.env

print_success "Version persisted to .deploy/tag.env"
echo ""

# Create commit if there are chart changes, then always create tag
print_info "Preparing git commit and tag..."

if ! git diff --quiet -- charts/devices-backend/values.yaml charts/mentor-backend/values.yaml; then
  git add charts/devices-backend/values.yaml charts/mentor-backend/values.yaml
  git commit -m "chore: release ${VERSION}

- Built and tagged Docker images: ${VERSION}
- Git commit: ${GIT_COMMIT}
- Updated Helm charts to use ${VERSION}
- Images pushed to GitHub Container Registry (GHCR)
"
else
  print_warning "No Helm chart changes to commit"
fi

if git rev-parse -q --verify "refs/tags/${VERSION}" >/dev/null; then
  print_warning "Git tag ${VERSION} already exists locally"
else
  git tag -a "${VERSION}" -m "Release ${VERSION}

Docker Images:
- ghcr.io/mj-nehme/raqeem-devices-backend:${VERSION}
- ghcr.io/mj-nehme/raqeem-mentor-backend:${VERSION}

Git Commit: ${GIT_COMMIT}
"
  print_success "Git tag created: ${VERSION}"
fi
echo ""

# Automatically push commit and tag to origin (no prompt)
print_info "Pushing commit and tag to GitHub (origin)"

# Safety check: ensure tag not already on remote
if git ls-remote --tags origin | grep -q "refs/tags/${VERSION}$"; then
  print_warning "Tag ${VERSION} already exists on remote origin. Skipping tag push."
  print_info "Pushing commit..."
  git push origin HEAD || { print_error "Failed to push commit"; exit 1; }
else
  print_info "Pushing commit..."
  git push origin HEAD || { print_error "Failed to push commit"; exit 1; }
  print_info "Pushing tag ${VERSION}..."
  git push origin "${VERSION}" || { print_error "Failed to push tag"; exit 1; }
fi

print_success "Remote push step completed"

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
print_success "Release ${VERSION} created successfully!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📦 Docker Images Tagged:"
echo "  • ghcr.io/mj-nehme/raqeem-devices-backend:${VERSION}"
echo "  • ghcr.io/mj-nehme/raqeem-devices-backend:${VERSION}-${GIT_COMMIT}"
echo "  • ghcr.io/mj-nehme/raqeem-devices-backend:latest"
echo ""
echo "  • ghcr.io/mj-nehme/raqeem-mentor-backend:${VERSION}"
echo "  • ghcr.io/mj-nehme/raqeem-mentor-backend:${VERSION}-${GIT_COMMIT}"
echo "  • ghcr.io/mj-nehme/raqeem-mentor-backend:latest"
echo ""
echo "📝 Changes Committed:"
echo "  • Helm charts updated to use ${VERSION}"
echo "  • Git commit: $(git rev-parse HEAD)"
echo ""
echo "🏷️  Git Tag: ${VERSION}"
echo ""
echo "📌 Next Steps:"
echo "  1. Deploy with: ./start.sh (uses ${VERSION})"
echo "  2. (If not auto-pushed) Run: git push && git push origin ${VERSION}"
echo ""
echo "🔄 To rollback to a specific version:"
echo "  • Update .deploy/tag.env with desired version"
echo "  • Run: ./start.sh"
echo ""
