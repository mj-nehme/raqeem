#!/usr/bin/env bash
# Pre-Release Testing Checklist
# Run this BEFORE creating v1.0.0 release

set -e

echo "🧪 Pre-Release Testing - Raqeem v1.0.0"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_pass() {
  echo -e "${GREEN}✅ $1${NC}"
}

print_fail() {
  echo -e "${RED}❌ $1${NC}"
}

print_test() {
  echo -e "${YELLOW}🧪 Testing: $1${NC}"
}

# Test 1: Environment is stopped
print_test "Checking clean state"
if kubectl get pods -n default 2>/dev/null | grep -q "raqeem\|devices-backend\|mentor-backend"; then
  print_fail "Services are already running. Run ./stop.sh first"
  exit 1
fi
print_pass "Clean state confirmed"
echo ""

# Test 2: Start environment
print_test "Starting complete environment"
if ! ./start.sh; then
  print_fail "Startup failed"
  exit 1
fi
print_pass "Environment started successfully"
echo ""

# Wait a bit for services to stabilize
sleep 5

# Test 3: Check all pods are running
print_test "Verifying all pods are ready"
PODS=$(kubectl get pods -n default -o json | jq -r '.items[] | select(.metadata.labels.app != null) | "\(.metadata.labels.app) \(.status.phase)"')
echo "$PODS" | while read -r app phase; do
  if [ "$phase" != "Running" ]; then
    print_fail "Pod $app is $phase (expected Running)"
    exit 1
  fi
  print_pass "Pod $app is $phase"
done
echo ""

# Test 4: Check backend health
print_test "Testing backend health"

# Get ports from .env
if [ -f .env ]; then
  source .env
fi

sleep 2

# Test Devices Backend
if curl -sf "http://localhost:${DEVICES_BACKEND_PORT}/docs" > /dev/null; then
  print_pass "Devices Backend responding on port ${DEVICES_BACKEND_PORT}"
else
  print_fail "Devices Backend not responding on port ${DEVICES_BACKEND_PORT}"
  exit 1
fi

# Test Mentor Backend  
if curl -sf "http://localhost:${MENTOR_BACKEND_PORT}/activities" > /dev/null; then
  print_pass "Mentor Backend responding on port ${MENTOR_BACKEND_PORT}"
else
  print_fail "Mentor Backend not responding on port ${MENTOR_BACKEND_PORT}"
  exit 1
fi
echo ""

# Test 5: Check frontend accessibility
print_test "Testing frontend accessibility"

if curl -sf "http://localhost:${MENTOR_FRONTEND_PORT}" | grep -q "<!doctype html"; then
  print_pass "Mentor Frontend responding on port ${MENTOR_FRONTEND_PORT}"
else
  print_fail "Mentor Frontend not responding on port ${MENTOR_FRONTEND_PORT}"
  exit 1
fi

if curl -sf "http://localhost:${DEVICES_FRONTEND_PORT}" | grep -q "<!doctype html"; then
  print_pass "Devices Frontend responding on port ${DEVICES_FRONTEND_PORT}"
else
  print_fail "Devices Frontend not responding on port ${DEVICES_FRONTEND_PORT}"
  exit 1
fi
echo ""

# Test 6: Check environment variables are set in pods
print_test "Verifying PORT environment variables in pods"

DEVICES_PORT=$(kubectl exec -n default $(kubectl get pod -l app=devices-backend -n default -o jsonpath='{.items[0].metadata.name}') -- env | grep -E '^PORT=' | cut -d= -f2)
if [ "$DEVICES_PORT" = "8081" ]; then
  print_pass "Devices Backend PORT=$DEVICES_PORT (correct)"
else
  print_fail "Devices Backend PORT=$DEVICES_PORT (expected 8081)"
  exit 1
fi

MENTOR_PORT=$(kubectl exec -n default $(kubectl get pod -l app=mentor-backend -n default -o jsonpath='{.items[0].metadata.name}') -- env | grep -E '^PORT=' | cut -d= -f2)
if [ "$MENTOR_PORT" = "8080" ]; then
  print_pass "Mentor Backend PORT=$MENTOR_PORT (correct)"
else
  print_fail "Mentor Backend PORT=$MENTOR_PORT (expected 8080)"
  exit 1
fi
echo ""

# Test 7: Check service discovery
print_test "Verifying Kubernetes service discovery"
SERVICES="postgres-service minio-service devices-backend mentor-backend"
for svc in $SERVICES; do
  if kubectl get service "$svc" -n default > /dev/null 2>&1; then
    print_pass "Service $svc exists and is discoverable"
  else
    print_fail "Service $svc not found"
    exit 1
  fi
done
echo ""

# Summary
echo "========================================"
echo -e "${GREEN}🎉 All Pre-Release Tests PASSED!${NC}"
echo "========================================"
echo ""
echo "✅ Environment starts successfully"
echo "✅ All pods are running and healthy"
echo "✅ Backends are responding correctly"
echo "✅ Frontends are accessible"
echo "✅ Environment variables configured"
echo "✅ Service discovery working"
echo ""
echo "📦 Current Images Being Used:"
kubectl get pods -n default -o jsonpath='{range .items[*]}{.metadata.labels.app}{"\t"}{.spec.containers[0].image}{"\n"}{end}' | grep -E "devices-backend|mentor-backend"
echo ""
echo "🏷️  Ready to create release v1.0.0!"
echo ""
echo "Next steps:"
echo "  1. Keep environment running and test manually (optional)"
echo "  2. When satisfied, stop: ./stop.sh"
echo "  3. Create release: ./scripts/tag-release.sh v1.0.0"
echo "  4. Push tag: git push origin v1.0.0 && git push"
echo ""
