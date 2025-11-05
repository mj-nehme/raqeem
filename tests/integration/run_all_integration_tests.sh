#!/usr/bin/env bash
# Comprehensive Integration Test Runner
# Runs all integration tests in sequence

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "=============================================================================="
echo -e "${BLUE}Raqeem Comprehensive Integration Test Suite${NC}"
echo "=============================================================================="
echo ""

cd "$PROJECT_ROOT"

# Check prerequisites
echo "Checking prerequisites..."
for cmd in docker python3; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo -e "${RED}✗ Missing required command: $cmd${NC}"
    exit 1
  fi
done

# Check for docker compose (v2) or docker-compose (v1)
if docker compose version >/dev/null 2>&1; then
  DOCKER_COMPOSE="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
  DOCKER_COMPOSE="docker-compose"
else
  echo -e "${RED}✗ Docker Compose not found${NC}"
  exit 1
fi

echo -e "${GREEN}✓ Prerequisites satisfied${NC}"
echo ""

# Clean up any existing containers
echo "Cleaning up existing containers..."
$DOCKER_COMPOSE -f .github/docker-compose.test.yml down -v 2>/dev/null || true
echo ""

# Start services
echo "Starting services with docker compose..."
$DOCKER_COMPOSE -f .github/docker-compose.test.yml up -d --build

if [ $? -ne 0 ]; then
  echo -e "${RED}✗ Failed to start services${NC}"
  exit 1
fi
echo -e "${GREEN}✓ Services started${NC}"
echo ""

# Wait for services to be healthy
echo "Waiting for services to be healthy..."
echo "This may take 30-60 seconds for backends to initialize..."

# Check health of all services
max_wait=60
waited=0
all_healthy=false

echo "Checking service health (timeout: ${max_wait}s)..."
while [ $waited -lt $max_wait ]; do
  # Check if services are healthy
  devices_status=$($DOCKER_COMPOSE -f .github/docker-compose.test.yml ps --format json devices-backend 2>/dev/null | grep -o '"Health":"[^"]*"' | cut -d'"' -f4 || echo "")
  mentor_status=$($DOCKER_COMPOSE -f .github/docker-compose.test.yml ps --format json mentor-backend 2>/dev/null | grep -o '"Health":"[^"]*"' | cut -d'"' -f4 || echo "")
  
  # Fallback to grep method if JSON parsing fails
  if [ -z "$devices_status" ]; then
    devices_health=$($DOCKER_COMPOSE -f .github/docker-compose.test.yml ps devices-backend 2>/dev/null | grep -c "healthy" || echo "0")
    [ "$devices_health" = "1" ] && devices_status="healthy"
  fi
  
  if [ -z "$mentor_status" ]; then
    mentor_health=$($DOCKER_COMPOSE -f .github/docker-compose.test.yml ps mentor-backend 2>/dev/null | grep -c "healthy" || echo "0")
    [ "$mentor_health" = "1" ] && mentor_status="healthy"
  fi
  
  if [ "$devices_status" = "healthy" ] && [ "$mentor_status" = "healthy" ]; then
    all_healthy=true
    break
  fi
  
  echo -n "."
  sleep 5
  waited=$((waited + 5))
done
echo ""

if [ "$all_healthy" = true ]; then
  echo -e "${GREEN}✓ All services are healthy${NC}"
else
  echo -e "${YELLOW}⚠ Services may not be fully ready, but proceeding with tests...${NC}"
fi
echo ""

# Install Python test dependencies
echo "Installing Python dependencies for tests..."
python3 -m pip install -q requests
echo ""

# Array to track test results
declare -a test_results
declare -a test_names

# Function to run a test
run_test() {
  local test_file=$1
  local test_name=$2
  
  echo ""
  echo "=============================================================================="
  echo -e "${BLUE}Running: $test_name${NC}"
  echo "=============================================================================="
  
  python3 "$test_file"
  local exit_code=$?
  
  test_results+=($exit_code)
  test_names+=("$test_name")
  
  if [ $exit_code -eq 0 ]; then
    echo -e "${GREEN}✓ $test_name PASSED${NC}"
  else
    echo -e "${RED}✗ $test_name FAILED${NC}"
  fi
  
  return $exit_code
}

# Run all integration tests
echo "=============================================================================="
echo -e "${BLUE}Starting Integration Test Suite${NC}"
echo "=============================================================================="

# Test 1: Devices Backend ↔ DB & S3
run_test "tests/integration/test_devices_backend_db_s3.py" "Devices Backend ↔ DB & S3"
TEST1_EXIT=$?

# Test 2: Mentor Backend ↔ DB & S3
run_test "tests/integration/test_mentor_backend_db_s3.py" "Mentor Backend ↔ DB & S3"
TEST2_EXIT=$?

# Test 3: Backend-to-Backend Communication
run_test "tests/integration/test_backend_communication.py" "Backend-to-Backend Communication"
TEST3_EXIT=$?

# Test 4: Original Alert Flow Test (backward compatibility)
run_test "tests/integration/test_alert_flow.py" "Alert Flow (Original)"
TEST4_EXIT=$?

# Test 5: End-to-End System Flow
run_test "tests/integration/test_e2e_system_flow.py" "End-to-End System Flow"
TEST5_EXIT=$?

# Summary
echo ""
echo "=============================================================================="
echo -e "${BLUE}Test Summary${NC}"
echo "=============================================================================="

total_tests=${#test_results[@]}
passed_tests=0
failed_tests=0

for i in "${!test_results[@]}"; do
  if [ "${test_results[$i]}" -eq 0 ]; then
    echo -e "${GREEN}✓${NC} ${test_names[$i]}"
    passed_tests=$((passed_tests + 1))
  else
    echo -e "${RED}✗${NC} ${test_names[$i]}"
    failed_tests=$((failed_tests + 1))
  fi
done

echo ""
echo "Total: $total_tests | Passed: $passed_tests | Failed: $failed_tests"
echo "=============================================================================="

# Show service logs if any tests failed
if [ $failed_tests -gt 0 ]; then
  echo ""
  echo -e "${YELLOW}Some tests failed. Showing service logs...${NC}"
  echo "=============================================================================="
  echo ""
  echo "Devices Backend logs (last 50 lines):"
  echo "---"
  $DOCKER_COMPOSE -f .github/docker-compose.test.yml logs --tail=50 devices-backend
  echo ""
  echo "Mentor Backend logs (last 50 lines):"
  echo "---"
  $DOCKER_COMPOSE -f .github/docker-compose.test.yml logs --tail=50 mentor-backend
  echo "=============================================================================="
fi

# Cleanup
echo ""
read -p "Keep services running for manual testing? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Stopping services..."
  $DOCKER_COMPOSE -f .github/docker-compose.test.yml down -v
  echo -e "${GREEN}✓ Services stopped and cleaned up${NC}"
else
  echo ""
  echo -e "${YELLOW}Services are still running. Access them at:${NC}"
  echo "  - Devices Backend: http://localhost:8081"
  echo "  - Devices API Docs: http://localhost:8081/docs"
  echo "  - Mentor Backend:  http://localhost:8080"
  echo "  - PostgreSQL:      localhost:5432"
  echo "  - MinIO:           localhost:9000"
  echo ""
  echo "To stop services later, run:"
  echo "  $DOCKER_COMPOSE -f .github/docker-compose.test.yml down -v"
fi

echo ""
if [ $failed_tests -eq 0 ]; then
  echo -e "${GREEN}=============================================================================="
  echo "✓ ALL INTEGRATION TESTS PASSED!"
  echo "==============================================================================${NC}"
  exit 0
else
  echo -e "${RED}=============================================================================="
  echo "✗ SOME INTEGRATION TESTS FAILED"
  echo "==============================================================================${NC}"
  exit 1
fi
