#!/usr/bin/env bash
set -euo pipefail

# Run GitHub Actions workflows locally using nektos/act
# Requirements: Docker Desktop and act (brew install act)
#
# Usage:
#   ./scripts/run-ci-local.sh [root|devices] [job]
#   ./scripts/run-ci-local.sh preset [prepush|full]
#   ./scripts/run-ci-local.sh battle               # run battle-tests job (uses release event)
# Examples:
#   ./scripts/run-ci-local.sh root                 # run default event for root CI
#   ./scripts/run-ci-local.sh root test-devices-backend
#   ./scripts/run-ci-local.sh root test-mentor-backend
#   ./scripts/run-ci-local.sh root test-devices-frontend
#   ./scripts/run-ci-local.sh root test-mentor-frontend
#   ./scripts/run-ci-local.sh devices test         # run devices/backend workflow job 'test'
#   ./scripts/run-ci-local.sh devices              # list jobs for devices workflow
#   ./scripts/run-ci-local.sh preset prepush       # lint + typecheck + all tests
#   ./scripts/run-ci-local.sh battle               # battle-tests job with release event

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

log() { echo -e "\n\033[1;36m==> $*\033[0m"; }
err() { echo -e "\033[1;31m[error]\033[0m $*"; }

has_cmd() { command -v "$1" >/dev/null 2>&1; }

print_usage() {
  cat <<USAGE
Run GitHub Actions locally with act.

Usage:
  ./scripts/run-ci-local.sh [root|devices] [job]
  ./scripts/run-ci-local.sh preset [prepush|full]
  ./scripts/run-ci-local.sh battle

Examples:
  ./scripts/run-ci-local.sh root                     # list jobs in root CI
  ./scripts/run-ci-local.sh root test-devices-backend
  ./scripts/run-ci-local.sh devices                  # list jobs in devices CI
  ./scripts/run-ci-local.sh devices test             # run 'test' job in devices workflow
  ./scripts/run-ci-local.sh preset prepush           # lint+typecheck+all tests
  ./scripts/run-ci-local.sh preset full              # includes service checks and build
  ./scripts/run-ci-local.sh battle                   # runs battle-tests job

Notes:
  - Apple Silicon is auto-detected; uses --container-architecture linux/amd64
  - Root CI uses dummy CODECOV_TOKEN/GITHUB_TOKEN locally (uploads are skipped)
  - Devices CI passes POSTGRES_PASSWORD=supersecret for its service job

Options:
  -h, --help   Show this help message and exit
USAGE
}

if ! has_cmd docker; then
  err "Docker is required. Install Docker Desktop and retry."
  exit 1
fi
if ! has_cmd act; then
  err "act is required. Install with: brew install act"
  exit 1
fi

# Use a modern Ubuntu image for act runners
# Use the GHCR runner image recommended by act
PLATFORM_MAP="ubuntu-latest=ghcr.io/catthehacker/ubuntu:act-22.04"
# Detect Apple Silicon and set container architecture to linux/amd64 (act flag)
ARCH_FLAG=()
if [[ $(uname -s) == "Darwin" ]]; then
  # sysctl can report arm64 on Apple Silicon
  if sysctl -n machdep.cpu.brand_string 2>/dev/null | grep -qi "Apple" || [[ $(uname -m) == "arm64" ]]; then
    ARCH_FLAG=("--container-architecture" "linux/amd64")
  fi
fi

scope="${1:-root}"
job="${2:-}"
shift || true

# Help flag
if [[ "$scope" == "-h" || "$scope" == "--help" ]]; then
  print_usage
  exit 0
fi

resolve_github_token() {
  # Try to use GitHub CLI auth token if available; fallback to env; else prompt
  local token=""
  if command -v gh >/dev/null 2>&1; then
    token="$(gh auth token 2>/dev/null || true)"
  fi
  if [[ -z "$token" && -n "${GITHUB_TOKEN:-}" ]]; then
    token="$GITHUB_TOKEN"
  fi
  echo "$token"
}

run_root_job() {
  local job_name="$1"
  local gh_token
  gh_token="$(resolve_github_token)"
  local extra_args=("-s" "CODECOV_TOKEN=dummy")
  if [[ -n "$gh_token" ]]; then
    extra_args+=("-s" "GITHUB_TOKEN=$gh_token")
  else
    echo "[warn] No GitHub token found. Public actions may fail to fetch."
    echo "       Tip: run 'gh auth login' or export GITHUB_TOKEN to improve reliability."
    extra_args+=("-s" "GITHUB_TOKEN=dummy")
  fi
  act -P "$PLATFORM_MAP" "${ARCH_FLAG[@]}" -W "$ROOT_DIR/.github/workflows/ci.yml" -j "$job_name" "${extra_args[@]}"
}

run_devices_job() {
  local job_name="$1"
  act -P "$PLATFORM_MAP" "${ARCH_FLAG[@]}" -W "$ROOT_DIR/devices/backend/.github/workflows/ci.yml" -j "$job_name" -s POSTGRES_PASSWORD="supersecret"
}

run_preset_prepush() {
  # Mirrors core checks before pushing
  log "Running pre-push preset via act (root CI)"
  local jobs=(
    lint-python
    typecheck-python
    lint-go
    lint-devices-frontend
    build-artifacts
    test-devices-backend
    test-mentor-backend
    test-devices-frontend
    test-mentor-frontend
  )
  for j in "${jobs[@]}"; do
    log "act job: $j"
    run_root_job "$j"
  done
}

run_preset_full() {
  log "Running full preset via act (root CI)"
  # Run all major jobs except publish-images (which requires GHCR auth)
  local jobs=(
    check-services
    lint-python
    typecheck-python
    lint-go
    lint-devices-frontend
    build-artifacts
    test-devices-backend
    test-mentor-backend
    test-devices-frontend
    test-mentor-frontend
  )
  for j in "${jobs[@]}"; do
    log "act job: $j"
    run_root_job "$j"
  done
}

run_battle_tests() {
  # battle-tests job has an if: release condition; emulate release event
  log "Running battle-tests via act using a release event"
  local event_file
  event_file="$(mktemp -t act-release-event.XXXX.json)"
  cat >"$event_file" <<EOF
{
  "action": "published",
  "release": { "tag_name": "v0.0.0-local" }
}
EOF
  # Pass tokens as dummy to avoid Codecov auth failures
  act -P "$PLATFORM_MAP" "${ARCH_FLAG[@]}" -W "$ROOT_DIR/.github/workflows/ci.yml" -j battle-tests -e "$event_file" -s CODECOV_TOKEN=dummy -s GITHUB_TOKEN=dummy
  rm -f "$event_file"
}

case "$scope" in
  root)
    if [[ -n "$job" ]]; then
      log "Root CI, job: $job"
      run_root_job "$job"
    else
      log "Listing available jobs in root CI"
      act -P "$PLATFORM_MAP" "${ARCH_FLAG[@]}" -W "$ROOT_DIR/.github/workflows/ci.yml" -l || true
      echo "To run a job: ./scripts/run-ci-local.sh root <job>"
    fi
    ;;
  devices)
    if [[ -n "$job" ]]; then
      log "Devices CI, job: $job"
      run_devices_job "$job"
    else
      log "Listing available jobs in devices workflow"
      act -P "$PLATFORM_MAP" "${ARCH_FLAG[@]}" -W "$ROOT_DIR/devices/backend/.github/workflows/ci.yml" -l || true
      echo "To run a job: ./scripts/run-ci-local.sh devices <job>"
    fi
    ;;
  preset)
    case "$job" in
      prepush)
        run_preset_prepush ;;
      full)
        run_preset_full ;;
      *)
        err "Unknown preset: $job (use prepush|full)"
        exit 1 ;;
    esac
    ;;
  battle)
    run_battle_tests ;;
  *)
    err "Unknown scope: $scope (use 'root'|'devices'|'preset'|'battle')"
    print_usage
    exit 1 ;;
esac
