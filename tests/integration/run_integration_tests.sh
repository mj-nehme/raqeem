#!/usr/bin/env bash
# Integration test runner script
# Starts services via docker-compose and runs E2E tests

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=========================================="
echo "Raqeem E2E Integration Test Runner"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

cd "$PROJECT_ROOT"

# Check prerequisites
echo "Checking prerequisites..."
for cmd in docker docker-compose python3; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo -e "${RED}✗ Missing required command: $cmd${NC}"
    exit 1
  fi
done
echo -e "${GREEN}✓ Prerequisites satisfied${NC}"
echo ""

# Clean up any existing containers
echo "Cleaning up existing containers..."
docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true
echo ""

# Start services
echo "Starting services with docker-compose..."
docker-compose -f docker-compose.test.yml up -d --build

if [ $? -ne 0 ]; then
  echo -e "${RED}✗ Failed to start services${NC}"
  exit 1
fi
echo -e "${GREEN}✓ Services started${NC}"
echo ""

# Wait for services to be healthy
echo "Waiting for services to be healthy..."
sleep 10

# Install Python test dependencies
echo "Installing Python dependencies for tests..."
python3 -m pip install -q requests

# Run integration test
echo ""
echo "Running integration tests..."
echo "=========================================="
python3 tests/integration/test_alert_flow.py
TEST_EXIT_CODE=$?

echo ""
echo "=========================================="
if [ $TEST_EXIT_CODE -eq 0 ]; then
  echo -e "${GREEN}✓ Integration tests passed!${NC}"
else
  echo -e "${RED}✗ Integration tests failed${NC}"
fi

# Show service logs if tests failed
if [ $TEST_EXIT_CODE -ne 0 ]; then
  echo ""
  echo "Service logs (last 50 lines):"
  echo "=========================================="
  echo ""
  echo "Devices Backend:"
  docker-compose -f docker-compose.test.yml logs --tail=50 devices-backend
  echo ""
  echo "Mentor Backend:"
  docker-compose -f docker-compose.test.yml logs --tail=50 mentor-backend
fi

# Cleanup
echo ""
read -p "Keep services running for manual testing? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Stopping services..."
  docker-compose -f docker-compose.test.yml down -v
  echo -e "${GREEN}✓ Services stopped and cleaned up${NC}"
else
  echo ""
  echo "Services are still running. Access them at:"
  echo "  - Devices Backend: http://localhost:8081"
  echo "  - Mentor Backend:  http://localhost:8080"
  echo ""
  echo "To stop services later, run:"
  echo "  docker-compose -f docker-compose.test.yml down -v"
fi

exit $TEST_EXIT_CODE
