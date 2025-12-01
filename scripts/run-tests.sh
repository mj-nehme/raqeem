#!/usr/bin/env bash
set -euo pipefail

# Raqeem test runner
# Usage:
#   ./scripts/run-tests.sh [suite]
#
# Suites:
#   devices-backend | devices-be | be-devices
#   mentor-backend  | mentor-be  | be-mentor
#   devices-frontend| devices-fe | fe-devices
#   mentor-frontend | mentor-fe  | fe-mentor
#   all
#
# Notes:
# - Devices backend tests require PostgreSQL. Set DATABASE_URL or start a local DB.
# - Frontend tests require npm dependencies installed in their folders.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

log() { echo -e "\n\033[1;36m==> $*\033[0m"; }
warn() { echo -e "\033[1;33m[warn]\033[0m $*"; }
err() { echo -e "\033[1;31m[error]\033[0m $*"; }

has_cmd() { command -v "$1" >/dev/null 2>&1; }

# --- Self-containment helpers -------------------------------------------------
ensure_act() {
  if has_cmd act; then return 0; fi
  warn "'act' not found. Attempting to install via Homebrew..."
  if has_cmd brew; then
    brew install act || {
      err "Failed to install 'act' via Homebrew. Install manually: https://github.com/nektos/act"
      exit 1
    }
  else
    err "Homebrew not found. Please install 'act' manually: https://github.com/nektos/act"
    exit 1
  fi
}

ensure_docker() {
  if ! has_cmd docker; then
    err "Docker is required. Install Docker Desktop and retry."
    exit 1
  fi
  if docker info >/dev/null 2>&1; then return 0; fi
  warn "Docker is not running. Attempting to start Docker Desktop..."
  if [[ "$(uname -s)" == "Darwin" ]]; then
    # Start Docker.app on macOS and wait until ready
    open -a Docker || true
    for i in {1..120}; do
      if docker info >/dev/null 2>&1; then
        log "Docker is up."
        return 0
      fi
      echo "Waiting for Docker... ($i/120)"; sleep 2
    done
    err "Docker did not become ready in time. Please start Docker Desktop manually."
    exit 1
  else
    err "Docker not running. Please start the Docker daemon and retry."
    exit 1
  fi
}

# Pre-pull common images used by CI to show progress and avoid act timeouts
pre_pull_ci_images() {
  # Only pre-pull for CI presets unless explicitly requested
  local enable="${RUN_TESTS_PREPULL:-auto}"
  if [[ "$enable" == "0" || "$enable" == "false" ]]; then
    return 0
  fi

  # Detect Apple Silicon and prefer amd64 images for compatibility with act
  local platform_flag=""
  if [[ "$(uname -s)" == "Darwin" && "$(uname -m)" == "arm64" ]]; then
    platform_flag="--platform=linux/amd64"
  fi

  # Act runner image and common service images from workflows
  local images=(
    "ghcr.io/catthehacker/ubuntu:act-22.04"
    "docker.io/library/postgres:16"
    "postgres:15"
    "minio/minio:latest"
  )

  log "Pre-pulling CI images to show progress (can skip with RUN_TESTS_PREPULL=0)"
  for img in "${images[@]}"; do
    echo "Pulling $img ..."
    # shellcheck disable=SC2086
    if ! docker pull $platform_flag "$img"; then
      warn "Failed to pull $img. act may attempt to pull during run."
    fi
  done
}

prompt_menu() {
  cat <<EOF
Select test suite to run:
  1) Devices Backend (Python/FastAPI)
  2) Mentor Backend (Go/Gin)
  3) Devices Frontend (Node/Vite)
  4) Mentor Frontend (Node/Vite)
  5) All
  6) CI Pre-Push (act preset)
  7) CI Full (act preset)
  8) CI Battle Tests (act)
  q) Quit
EOF
  read -r -p "Enter choice [1-8/q]: " choice
  case "$choice" in
    1) echo "devices-backend" ;;
    2) echo "mentor-backend" ;;
    3) echo "devices-frontend" ;;
    4) echo "mentor-frontend" ;;
    5) echo "all" ;;
    6) echo "ci-prepush" ;;
    7) echo "ci-full" ;;
    8) echo "ci-battle" ;;
    q|Q) echo "quit" ;;
    *) echo "invalid" ;;
  esac
}

run_devices_backend() {
  log "Devices backend tests"
  pushd "$ROOT_DIR/devices/backend/src" >/dev/null

  if ! has_cmd pytest; then
    warn "pytest not found. Install dev deps first, e.g.:"
    echo "    pip install -r ../requirements.txt -r ../requirements-test.txt"
    exit 1
  fi

  if [[ -z "${DATABASE_URL:-}" ]]; then
    warn "DATABASE_URL is not set. Example:"
    echo "    export DATABASE_URL='postgres+asyncpg://monitor:password@127.0.0.1:5432/monitoring_db'"
    echo "    (You can start a local Postgres via Docker; see docs/TESTING.md)"
    exit 1
  fi

  PYTHONPATH=src pytest --maxfail=1 --disable-warnings -q
  popd >/dev/null
}

run_mentor_backend() {
  log "Mentor backend tests"
  pushd "$ROOT_DIR/mentor/backend/src" >/dev/null

  if ! has_cmd go; then
    err "Go toolchain not found. Install Go and retry."
    exit 1
  fi

  # Reasonable defaults (override in env if needed)
  export POSTGRES_USER="${POSTGRES_USER:-monitor}"
  export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-password}"
  export POSTGRES_DB="${POSTGRES_DB:-monitoring_db}"
  export POSTGRES_HOST="${POSTGRES_HOST:-127.0.0.1}"
  export POSTGRES_PORT="${POSTGRES_PORT:-5432}"

  go test ./... -v -race -p 1
  popd >/dev/null
}

run_devices_frontend() {
  log "Devices frontend tests"
  pushd "$ROOT_DIR/devices/frontend" >/dev/null
  if ! has_cmd npm; then
    err "npm not found. Install Node.js (>=18) and retry."
    exit 1
  fi
  npm test -- --coverage || npm test
  popd >/dev/null
}

run_mentor_frontend() {
  log "Mentor frontend tests"
  pushd "$ROOT_DIR/mentor/frontend" >/dev/null
  if ! has_cmd npm; then
    err "npm not found. Install Node.js (>=18) and retry."
    exit 1
  fi
  npm test -- --coverage || npm test
  popd >/dev/null
}

run_all() {
  run_mentor_backend
  run_devices_backend
  run_devices_frontend
  run_mentor_frontend
}

# Default behavior: run CI pre-push preset non-interactively for parity
suite="${1:-}"
if [[ -z "$suite" ]]; then
  suite="ci-prepush"
fi
# macOS ships bash 3.x which doesn't support ${var,,}; use tr for portability
suite_lc="$(printf '%s' "$suite" | tr '[:upper:]' '[:lower:]')"

case "$suite_lc" in
  devices-backend|devices-be|be-devices)
    run_devices_backend ;;
  mentor-backend|mentor-be|be-mentor)
    run_mentor_backend ;;
  devices-frontend|devices-fe|fe-devices)
    run_devices_frontend ;;
  mentor-frontend|mentor-fe|fe-mentor)
    run_mentor_frontend ;;
  all)
    run_all ;;
  ci-prepush)
    log "Running CI pre-push preset via act (self-contained)"
    ensure_docker
    ensure_act
    pre_pull_ci_images
    "$ROOT_DIR/scripts/run-ci-local.sh" preset prepush ;;
  ci-full)
    log "Running CI full preset via act (self-contained)"
    ensure_docker
    ensure_act
    pre_pull_ci_images
    "$ROOT_DIR/scripts/run-ci-local.sh" preset full ;;
  ci-battle)
    log "Running CI battle-tests via act (self-contained)"
    ensure_docker
    ensure_act
    pre_pull_ci_images
    "$ROOT_DIR/scripts/run-ci-local.sh" battle ;;
  quit)
    exit 0 ;;
  invalid|*)
    err "Invalid selection. See usage in header. To use interactive mode, run: RUN_TESTS_INTERACTIVE=1 ./scripts/run-tests.sh"
    exit 1 ;;
esac
