#!/usr/bin/env bash
set -euo pipefail

NAMESPACE=${NAMESPACE:-default}

# Load .env if present
if [[ -f ./.env ]]; then
  set -a
  # shellcheck disable=SC1091
  source ./.env
  set +a
fi

# Require ports to be set (no hardcoded defaults)
if [[ -z "${DEVICES_BACKEND_PORT:-}" ]]; then
  fail_then_exit "DEVICES_BACKEND_PORT not set. Define it in .env"
fi
if [[ -z "${MENTOR_BACKEND_PORT:-}" ]]; then
  fail_then_exit "MENTOR_BACKEND_PORT not set. Define it in .env"
fi
if [[ -z "${DEVICES_FRONTEND_PORT:-}" ]]; then
  warn "DEVICES_FRONTEND_PORT not set. Skipping devices frontend checks"
  DEVICES_FRONTEND_PORT=""
fi
if [[ -z "${MENTOR_FRONTEND_PORT:-}" ]]; then
  warn "MENTOR_FRONTEND_PORT not set. Skipping mentor frontend checks"
  MENTOR_FRONTEND_PORT=""
fi
TIMEOUT=${TIMEOUT:-120}
SLEEP=${SLEEP:-2}

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0

pass() { echo -e "${GREEN}✔${NC} $1"; PASS_COUNT=$((PASS_COUNT+1)); }
fail() { echo -e "${RED}✘${NC} $1"; FAIL_COUNT=$((FAIL_COUNT+1)); }
warn() { echo -e "${YELLOW}⚠${NC} $1"; WARN_COUNT=$((WARN_COUNT+1)); }
info() { echo -e "${BLUE}ℹ${NC} $1"; }

fail_then_exit() {
  fail "$1"
  echo ""; echo "===== Summary ====="; echo "Pass: $PASS_COUNT  Warn: $WARN_COUNT  Fail: $FAIL_COUNT"
  exit 1
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    fail "Missing required command: $1"
    exit 1
  fi
}

require_cmd kubectl
require_cmd curl
require_cmd nc

# Track background PF PIDs and clean up
PF_PIDS=()
cleanup() {
  for pid in "${PF_PIDS[@]:-}"; do
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      wait "$pid" 2>/dev/null || true
    fi
  done
}
trap cleanup EXIT

pf_svc() {
  local svc=$1
  local local_port=$2
  local remote_port=$3
  kubectl -n "$NAMESPACE" port-forward "svc/${svc}" "${local_port}:${remote_port}" >/dev/null 2>&1 &
  local pid=$!
  PF_PIDS+=("$pid")
  # Wait until localhost port is ready
  for _ in $(seq 1 50); do
    if nc -z localhost "$local_port" 2>/dev/null; then
      sleep 0.2
      return 0
    fi
    sleep 0.2
  done
  return 1
}

pf_pod() {
  local pod=$1
  local local_port=$2
  local container_port=$3
  kubectl -n "$NAMESPACE" port-forward "pod/${pod}" "${local_port}:${container_port}" >/dev/null 2>&1 &
  local pid=$!
  PF_PIDS+=("$pid")
  for _ in $(seq 1 50); do
    if nc -z localhost "$local_port" 2>/dev/null; then
      sleep 0.2
      return 0
    fi
    sleep 0.2
  done
  return 1
}

http_retry() {
  local url=$1
  local attempts=${2:-20}
  local delay=${3:-0.3}
  for _ in $(seq 1 "$attempts"); do
    if curl -fsS "$url" -o /dev/null; then
      return 0
    fi
    sleep "$delay"
  done
  return 1
}

get_running_pod_by_pattern() {
  local pattern=$1
  kubectl -n "$NAMESPACE" get pods -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.phase}{"\n"}{end}' \
    | awk -v pat="$pattern" '$1 ~ pat && $2 == "Running" {print $1; exit}'
}

check_deploy_ready() {
  local deploy=$1
  local desired available
  desired=$(kubectl -n "$NAMESPACE" get deploy "$deploy" -o jsonpath='{.status.replicas}' 2>/dev/null || echo 0)
  available=$(kubectl -n "$NAMESPACE" get deploy "$deploy" -o jsonpath='{.status.availableReplicas}' 2>/dev/null || echo 0)
  if [[ "$desired" == "$available" && "$available" != "0" ]]; then
    pass "Deployment $deploy available ($available/$desired)"
  else
    fail "Deployment $deploy not ready (available=$available desired=$desired)"
  fi
}

echo ""
echo "===== Verifying Kubernetes deployments ====="
check_deploy_ready postgres || true
check_deploy_ready minio || true
check_deploy_ready mentor-backend || true
check_deploy_ready devices-backend || true

echo "" 
echo "===== Verifying PostgreSQL ====="
PG_POD=$(get_running_pod_by_pattern 'postgres') || true
if [[ -n "${PG_POD:-}" ]]; then
  if kubectl -n "$NAMESPACE" exec "$PG_POD" -- sh -lc 'pg_isready' >/dev/null 2>&1; then
    pass "PostgreSQL pg_isready reported ready in pod $PG_POD"
  else
    fail_then_exit "PostgreSQL not ready (pg_isready failed) in pod $PG_POD"
  fi
else
  fail_then_exit "PostgreSQL pod not found"
fi

echo ""
echo "===== Verifying MinIO API health ====="
MINIO_API_PORT=${MINIO_API_PORT:-9000}
if pf_svc minio-service ${MINIO_API_PORT} 9000; then
  if curl -fsS "http://localhost:${MINIO_API_PORT}/minio/health/ready" >/dev/null; then
    pass "MinIO health endpoint OK"
  else
    fail_then_exit "MinIO health endpoint failed"
  fi
else
  fail_then_exit "Failed to port-forward MinIO service"
fi

echo ""
echo "===== Verifying Devices Backend (${DEVICES_BACKEND_PORT}) ====="
DEV_POD=$(get_running_pod_by_pattern 'devices-backend') || true
if [[ -z "${DEV_POD:-}" ]]; then
  fail_then_exit "Devices backend pod not found"
fi
if pf_pod "$DEV_POD" ${DEVICES_BACKEND_PORT} ${DEVICES_BACKEND_PORT}; then
  if http_retry http://localhost:${DEVICES_BACKEND_PORT}/docs 10 0.3; then
    pass "Devices backend docs reachable"
  else
    warn "Devices backend /docs not reachable (continuing)"
  fi
  if http_retry http://localhost:${DEVICES_BACKEND_PORT}/api/v1/devices 20 0.3; then
    pass "Devices backend API list devices reachable"
  else
    fail_then_exit "Devices backend API /api/v1/devices not reachable"
  fi
  # Optional DB write check: attempt to register a test device
  DID="verify-$(date +%s)"
  if curl -fsS -X POST http://localhost:${DEVICES_BACKEND_PORT}/api/v1/devices/register \
      -H 'Content-Type: application/json' \
      -d "{\"id\":\"$DID\",\"name\":\"Verifier\",\"type\":\"laptop\",\"os\":\"macOS\",\"current_user\":\"verifier\"}" >/dev/null; then
    pass "Devices backend device registration succeeded ($DID)"
  else
    warn "Devices backend device registration failed (DB path might be down)"
  fi
else
  fail_then_exit "Failed to port-forward to devices-backend pod $DEV_POD"
fi

echo ""
echo "===== Verifying Mentor Backend (${MENTOR_BACKEND_PORT}) ====="
MEN_POD=$(get_running_pod_by_pattern 'mentor-backend') || true
if [[ -z "${MEN_POD:-}" ]]; then
  fail_then_exit "Mentor backend pod not found"
fi
if pf_pod "$MEN_POD" ${MENTOR_BACKEND_PORT} ${MENTOR_BACKEND_PORT}; then
  if http_retry http://localhost:${MENTOR_BACKEND_PORT}/activities 20 0.3; then
    pass "Mentor backend activities endpoint reachable"
  else
    fail_then_exit "Mentor backend /activities not reachable"
  fi
else
  fail_then_exit "Failed to port-forward to mentor-backend pod $MEN_POD"
fi

echo ""
echo "===== Checking local frontends (best-effort) ====="
if curl -fsS http://localhost:${DEVICES_FRONTEND_PORT} >/dev/null; then
  pass "Devices frontend reachable at http://localhost:${DEVICES_FRONTEND_PORT}"
else
  warn "Devices frontend not reachable at http://localhost:${DEVICES_FRONTEND_PORT} (start Vite dev server to use it)"
fi
if curl -fsS http://localhost:${MENTOR_FRONTEND_PORT} >/dev/null; then
  pass "Mentor frontend reachable at http://localhost:${MENTOR_FRONTEND_PORT}"
else
  warn "Mentor frontend not reachable at http://localhost:${MENTOR_FRONTEND_PORT} (start Vite dev server to use it)"
fi

echo ""
echo "===== Summary ====="
echo "Pass: $PASS_COUNT  Warn: $WARN_COUNT  Fail: $FAIL_COUNT"
if [[ $FAIL_COUNT -gt 0 ]]; then
  exit 1
fi
exit 0
