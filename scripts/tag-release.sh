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

  # Try to login using provided credentials
  local USERNAME="${GHCR_USERNAME:-${GH_USERNAME:-${GITHUB_ACTOR}}}"
  local TOKEN="${GHCR_PAT:-${GH_TOKEN:-${GITHUB_TOKEN}}}"

  if [[ -n "$TOKEN" ]]; then
    if [[ -z "$USERNAME" ]]; then
      print_warning "GHCR username not provided. Set GHCR_USERNAME or GH_USERNAME (fallbacks to GITHUB_ACTOR in CI)."
      return 1
    fi
    print_info "Attempting non-interactive docker login to ghcr.io as '$USERNAME'..."
    if echo "$TOKEN" | docker login ghcr.io -u "$USERNAME" --password-stdin; then
      print_success "Authenticated to GHCR"
      return 0
    else
      print_error "Docker login to GHCR failed"
      return 1
    fi
  fi

  # No token available; require manual login
  return 1
}

# Check if version argument is provided
if [[ -z "$1" ]]; then
  print_error "Usage: $0 <version> [--skip-tests]"
  echo ""
  echo "Examples:"
  echo "  $0 v1.0.0"
  echo "  $0 v1.1.0 --skip-tests"
  echo ""
  exit 1
fi

VERSION=$1
SKIP_TESTS=${2:-""}

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
  
  # Test Helm charts
  print_info "  Validating Helm charts..."
  for chart in charts/*/; do
    chart_name=$(basename "$chart")
    if ! helm lint "$chart" > /dev/null 2>&1; then
      print_error "Helm chart validation failed: $chart_name"
      exit 1
    fi
  done
  
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
  echo ""
  echo "I couldn't log in automatically. Authenticate using one of these methods:"
  echo ""
  echo "1) Non-interactive env vars (recommended):"
  echo "   export GHCR_USERNAME=your-github-username"
  echo "   export GHCR_PAT=your-personal-access-token"
  echo "   # Required scopes: write:packages, read:packages"
  echo ""
  echo "   Or in GitHub Actions, prefer GITHUB_TOKEN and the docker/login-action."
  echo ""
  echo "2) GitHub CLI:"
  echo "   gh auth login"
  echo "   gh auth token | docker login ghcr.io -u \"$USER\" --password-stdin"
  echo ""
  echo "3) Manual PAT login:"
  echo "   echo YOUR_TOKEN | docker login ghcr.io -u your-github-username --password-stdin"
  echo ""
  exit 1
fi
echo ""

# Confirmation prompt before pushing to registry
print_warning "Ready to push Docker images to GitHub Container Registry (ghcr.io)"
echo ""
echo "This will publish the following images:"
echo "  • ghcr.io/mj-nehme/raqeem-devices-backend:${VERSION}"
echo "  • ghcr.io/mj-nehme/raqeem-devices-backend:${VERSION}-${GIT_COMMIT}"
echo "  • ghcr.io/mj-nehme/raqeem-devices-backend:latest"
echo "  • ghcr.io/mj-nehme/raqeem-mentor-backend:${VERSION}"
echo "  • ghcr.io/mj-nehme/raqeem-mentor-backend:${VERSION}-${GIT_COMMIT}"
echo "  • ghcr.io/mj-nehme/raqeem-mentor-backend:latest"
echo ""
read -p "Continue with push to GHCR? (yes/no): " confirm
if [[ "$confirm" != "yes" ]]; then
  print_warning "Push cancelled by user"
  echo ""
  echo "Images are built locally and ready for use with 'pullPolicy: Never'"
  exit 0
fi
echo ""

# Push images to registry
print_info "Pushing images to GitHub Container Registry..."
echo ""

print_info "  Pushing Devices Backend images..."
docker push ghcr.io/mj-nehme/raqeem-devices-backend:${VERSION}
docker push ghcr.io/mj-nehme/raqeem-devices-backend:${VERSION}-${GIT_COMMIT}
docker push ghcr.io/mj-nehme/raqeem-devices-backend:latest

print_info "  Pushing Mentor Backend images..."
docker push ghcr.io/mj-nehme/raqeem-mentor-backend:${VERSION}
docker push ghcr.io/mj-nehme/raqeem-mentor-backend:${VERSION}-${GIT_COMMIT}
docker push ghcr.io/mj-nehme/raqeem-mentor-backend:latest

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

# Create git tag
print_info "Creating git tag..."
git add charts/devices-backend/values.yaml charts/mentor-backend/values.yaml
git commit -m "chore: release ${VERSION}

- Built and tagged Docker images: ${VERSION}
- Git commit: ${GIT_COMMIT}
- Updated Helm charts to use ${VERSION}
- Images pushed to GitHub Container Registry (GHCR)
"

git tag -a "${VERSION}" -m "Release ${VERSION}

Docker Images:
- ghcr.io/mj-nehme/raqeem-devices-backend:${VERSION}
- ghcr.io/mj-nehme/raqeem-mentor-backend:${VERSION}

Git Commit: ${GIT_COMMIT}
"

print_success "Git tag created: ${VERSION}"
echo ""

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
echo "  1. Push the tag: git push origin ${VERSION}"
echo "  2. Push the commit: git push"
echo "  3. Deploy with: ./start.sh (will use ${VERSION} automatically)"
echo ""
echo "🔄 To rollback to a specific version:"
echo "  • Update .deploy/tag.env with desired version"
echo "  • Run: ./start.sh"
echo ""
