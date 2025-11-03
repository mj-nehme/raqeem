#!/usr/bin/env bash
# Smart Service Discovery Health Check

set -e

print_test() { echo "🧪 $1"; }
print_pass() { echo "✅ $1"; }
print_fail() { echo "❌ $1"; }

echo "🔍 Testing Raqeem with Smart Service Discovery..."
echo ""

# Use discovery CLI for health checks
print_test "Checking service discovery registry"
if [ -d ".deploy/registry" ] && [ "$(ls -A .deploy/registry 2>/dev/null)" ]; then
  print_pass "Service registry exists and populated"
  echo ""
  ./scripts/discover.sh list
else
  print_fail "Service registry not found - run ./start.sh first"
  exit 1
fi

echo ""
print_test "Running health checks on all discovered services"
if ./scripts/discover.sh health; then
  print_pass "All services are healthy"
else
  print_fail "Some services are unhealthy"
  exit 1
fi

echo ""
print_test "Testing Kubernetes pod status"
if kubectl get pods -n default --no-headers | grep -E "(devices-backend|mentor-backend|postgres|minio)" | grep -v Running; then
  print_fail "Some pods are not running"
  kubectl get pods -n default
  exit 1
else
  print_pass "All backend pods are running"
fi

echo ""
echo "🎉 All tests passed! System is healthy."