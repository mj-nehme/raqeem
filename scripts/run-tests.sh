#!/usr/bin/env bash
set -euo pipefail

# Raqeem test runner (portable to macOS bash 3.x)
# Usage:
#   ./scripts/run-tests.sh [suite]
#   ./scripts/run-tests.sh -h | --help
#   RUN_TESTS_INTERACTIVE=1 ./scripts/run-tests.sh
#   ./scripts/run-tests.sh suite1 suite2 ... (aggregated summary)
# Suites:
#   devices-backend | devices-be | be-devices | dbe | db
#   mentor-backend  | mentor-be  | be-mentor  | mbe | mb
#   devices-frontend| devices-fe | fe-devices | dfe | df
#   mentor-frontend | mentor-fe  | fe-mentor  | mfe | mf
#   all | ci-prepush | ci-full | ci-battle | quit
# Notes:
# - Devices backend needs DATABASE_URL (postgres+asyncpg://...)
# - Frontends need npm deps installed (npm ci || npm install)
# - Multiple suites aggregate pass/fail instead of stopping early.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

log()  { echo -e "\n\033[1;36m==> $*\033[0m"; }
warn() { echo -e "\033[1;33m[warn]\033[0m $*"; }
err()  { echo -e "\033[1;31m[error]\033[0m $*"; }
has_cmd() { command -v "$1" >/dev/null 2>&1; }

print_usage() {
  cat <<'EOF'
Raqeem test runner

Usage:
  ./scripts/run-tests.sh [suite]
  ./scripts/run-tests.sh -h | --help
  RUN_TESTS_INTERACTIVE=1 ./scripts/run-tests.sh
  ./scripts/run-tests.sh suite1 suite2 ...

Suites:
  devices-backend | devices-be | be-devices | dbe | db
  mentor-backend  | mentor-be  | be-mentor  | mbe | mb
  devices-frontend| devices-fe | fe-devices | dfe | df
    mentor-frontend | mentor-fe  | fe-mentor  | mfe | mf | dashboard
  all | ci-prepush | ci-full | ci-battle | quit

Notes:
  - Devices backend tests require DATABASE_URL (postgres+asyncpg driver).
  - Frontends require Node >=18 and installed dependencies.
  - Multiple suites produce a combined summary.
  - RUN_TESTS_PREPULL=0 to skip Docker image pre-pulls for CI presets.
EOF
}

ensure_act() {
  if has_cmd act; then return 0; fi
  warn "'act' not found. Attempting Homebrew install..."
  if has_cmd brew; then
    brew install act || { err "Failed to install 'act'."; exit 1; }
  else
    err "Homebrew not found. Install manually: https://github.com/nektos/act"; exit 1
  fi
}

ensure_docker() {
  if ! has_cmd docker; then err "Docker is required."; exit 1; fi
  if docker info >/dev/null 2>&1; then return 0; fi
  warn "Docker not running; attempting start (macOS only)."
  if [[ "$(uname -s)" == "Darwin" ]]; then
    open -a Docker || true
    for i in {1..120}; do
      if docker info >/dev/null 2>&1; then log "Docker is up."; return 0; fi
      echo "Waiting for Docker... ($i/120)"; sleep 2
    done
    err "Docker did not become ready in time."; exit 1
  else
    err "Start Docker daemon and retry."; exit 1
  fi
}

pre_pull_ci_images() {
  local enable="${RUN_TESTS_PREPULL:-auto}"
  [[ "$enable" == "0" || "$enable" == "false" ]] && return 0
  local platform_flag=""
  if [[ "$(uname -s)" == "Darwin" && "$(uname -m)" == "arm64" ]]; then
    platform_flag="--platform=linux/amd64"
  fi
  local images=(
    ghcr.io/catthehacker/ubuntu:act-22.04
    docker.io/library/postgres:16
    postgres:15
    minio/minio:latest
  )
  log "Pre-pulling CI images (skip: RUN_TESTS_PREPULL=0)"
  for img in "${images[@]}"; do
    echo "Pulling $img ..."
    if ! docker pull $platform_flag "$img"; then warn "Failed to pull $img"; fi
  done
}

prompt_menu() {
  cat <<EOF
Select test suite:
  1) Devices Backend
  2) Mentor Backend
  3) Devices Frontend
  4) Mentor Frontend
  5) All (summary)
  6) CI Pre-Push (act)
  7) CI Full (act)
  8) CI Battle (act)
  q) Quit
EOF
  read -r -p "Choice [1-8/q]: " choice
  case "$choice" in
    1) echo devices-backend ;;
    2) echo mentor-backend ;;
    3) echo devices-frontend ;;
    4) echo mentor-frontend ;;
    5) echo all ;;
    6) echo ci-prepush ;;
    7) echo ci-full ;;
    8) echo ci-battle ;;
    q|Q) echo quit ;;
    *) echo invalid ;;
  esac
}

run_devices_backend() {
  log "Devices backend tests"
  pushd "$ROOT_DIR/devices/backend/src" >/dev/null
  if ! has_cmd pytest; then
    warn "pytest missing. Install with: pip install -r ../requirements.txt -r ../requirements-test.txt"
    popd >/dev/null; return 1
  fi
  if [[ -z "${DATABASE_URL:-}" ]]; then
    warn "DATABASE_URL not set. Example: export DATABASE_URL='postgres+asyncpg://monitor:password@127.0.0.1:5432/monitoring_db'"
    popd >/dev/null; return 1
  fi
  PYTHONPATH=src pytest --maxfail=1 --disable-warnings -q || { popd >/dev/null; return 1; }
  popd >/dev/null
}

run_mentor_backend() {
  log "Mentor backend tests"
  pushd "$ROOT_DIR/mentor/backend/src" >/dev/null
  if ! has_cmd go; then err "Go toolchain not found."; popd >/dev/null; return 1; fi
  export POSTGRES_USER="${POSTGRES_USER:-monitor}"
  export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-password}"
  export POSTGRES_DB="${POSTGRES_DB:-monitoring_db}"
  export POSTGRES_HOST="${POSTGRES_HOST:-127.0.0.1}"
  export POSTGRES_PORT="${POSTGRES_PORT:-5432}"
  go test ./... -v -race -p 1 || { popd >/dev/null; return 1; }
  popd >/dev/null
}

run_devices_frontend() {
  log "Devices frontend tests"
  pushd "$ROOT_DIR/devices/web-simulator" >/dev/null
  if ! has_cmd npm; then err "npm not found (Node >=18 required)."; popd >/dev/null; return 1; fi
  npm test -- --coverage || npm test || { popd >/dev/null; return 1; }
  popd >/dev/null
}

run_mentor_frontend() {
  log "Dashboard tests"
  pushd "$ROOT_DIR/mentor/dashboard" >/dev/null
  if ! has_cmd npm; then err "npm not found (Node >=18 required)."; popd >/dev/null; return 1; fi
  npm test -- --coverage || npm test || { popd >/dev/null; return 1; }
  popd >/dev/null
}

run_all() {
    local order=(mentor-backend devices-backend devices-frontend mentor-frontend)
  log "Running all suites (summary at end)"
  local names=(); local statuses=(); local failures=0
  for s in "${order[@]}"; do
    echo "\n--- Suite: $s ---"
    if "$0" "$s"; then names+=("$s"); statuses+=(success); else names+=("$s"); statuses+=(fail); ((failures++)) || true; fi
  done
  echo "\n================ Summary ================"
  local i
  for ((i=0; i<${#names[@]}; i++)); do printf '%-20s %s\n' "${names[$i]}" "${statuses[$i]}"; done
  echo "-----------------------------------------"
  (( failures > 0 )) && { err "One or more suites failed ($failures)."; return 1; } || { log "All suites succeeded."; return 0; }
}

# --- Parse arguments / interactive -------------------------------------------
arg_count=$#
if (( arg_count == 0 )) && [[ "${RUN_TESTS_INTERACTIVE:-0}" != "0" ]]; then
  suite="$(prompt_menu)"
else
  suite="${1:-}"
fi

case "${suite:-}" in
  -h|--help|/?) print_usage; exit 0 ;;
esac

# Multi-suite path
if (( arg_count > 1 )); then
  suites=("$@")
  expanded=()
  for s in "${suites[@]}"; do
    slc="$(printf '%s' "$s" | tr '[:upper:]' '[:lower:]')"
    case "$slc" in
      all) expanded+=(mentor-backend devices-backend devices-frontend mentor-frontend) ;;
      -h|--help|/?) print_usage; exit 0 ;;
      *) expanded+=("$slc") ;;
    esac
  done
  unique=(); seen=""
  for s in "${expanded[@]}"; do
    if [[ " $seen " != *" $s "* ]]; then unique+=("$s"); seen+=" $s"; fi
  done
  log "Running multiple suites: ${unique[*]}"
  names=(); statuses=(); failures=0
  for s in "${unique[@]}"; do
    echo "\n--- Suite: $s ---"
    if "$0" "$s"; then names+=("$s"); statuses+=(success); else names+=("$s"); statuses+=(fail); ((failures++)) || true; fi
  done
  echo "\n================ Summary ================"
  for ((i=0; i<${#names[@]}; i++)); do printf '%-20s %s\n' "${names[$i]}" "${statuses[$i]}"; done
  echo "-----------------------------------------"
  (( failures > 0 )) && { err "One or more suites failed ($failures)."; exit 1; } || { log "All suites succeeded."; exit 0; }
fi

[[ -z "${suite}" ]] && suite="ci-prepush"
suite_lc="$(printf '%s' "$suite" | tr '[:upper:]' '[:lower:]')"

case "$suite_lc" in
  devices-backend|devices-be|be-devices|dbe|db) run_devices_backend ;;
  mentor-backend|mentor-be|be-mentor|mbe|mb)   run_mentor_backend ;;
  devices-frontend|devices-fe|fe-devices|dfe|df) run_devices_frontend ;;
  mentor-frontend|mentor-fe|fe-mentor|mfe|mf|dashboard) run_mentor_frontend ;;
  all) run_all ;;
  ci-prepush)
    log "Running CI pre-push preset via act"
    ensure_docker; ensure_act; pre_pull_ci_images
    "$ROOT_DIR/scripts/run-ci-local.sh" preset prepush ;;
  ci-full)
    log "Running CI full preset via act"
    ensure_docker; ensure_act; pre_pull_ci_images
    "$ROOT_DIR/scripts/run-ci-local.sh" preset full ;;
  ci-battle)
    log "Running CI battle-tests via act"
    ensure_docker; ensure_act; pre_pull_ci_images
    "$ROOT_DIR/scripts/run-ci-local.sh" battle ;;
  quit) exit 0 ;;
  invalid|*) err "Invalid selection. Use -h for help. For interactive: RUN_TESTS_INTERACTIVE=1 ./scripts/run-tests.sh"; exit 1 ;;
esac
