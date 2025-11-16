#!/bin/bash
# Battle Test Runner Script
# Runs all battle test suites and generates comprehensive report

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_DIR="${SCRIPT_DIR}/results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="${RESULTS_DIR}/battle_test_report_${TIMESTAMP}.json"

# Test configuration (use smaller values for quicker runs)
STRESS_DEVICES=${STRESS_DEVICES:-100}
STRESS_DURATION=${STRESS_DURATION:-60}
LOAD_USERS=${LOAD_USERS:-50}
LOAD_DURATION=${LOAD_DURATION:-60}
BENCHMARK_SAMPLES=${BENCHMARK_SAMPLES:-100}

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Raqeem Battle Test Suite Runner${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Timestamp: $(date)"
echo "Results will be saved to: ${REPORT_FILE}"
echo ""

# Create results directory
mkdir -p "${RESULTS_DIR}"

# Check if services are running
echo -e "${YELLOW}Checking service health...${NC}"
if ! curl -sf http://localhost:8081/health > /dev/null 2>&1; then
    echo -e "${RED}✗ Devices backend is not running at localhost:8081${NC}"
    echo "Please start services first:"
    echo "  docker-compose -f .github/docker-compose.test.yml up -d"
    echo "  OR"
    echo "  ./start.sh"
    exit 1
fi

if ! curl -sf http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${RED}✗ Mentor backend is not running at localhost:8080${NC}"
    echo "Please start services first:"
    echo "  docker-compose -f .github/docker-compose.test.yml up -d"
    echo "  OR"
    echo "  ./start.sh"
    exit 1
fi

echo -e "${GREEN}✓ Services are healthy${NC}"
echo ""

# Initialize report
cat > "${REPORT_FILE}" <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "configuration": {
    "stress_devices": ${STRESS_DEVICES},
    "stress_duration": ${STRESS_DURATION},
    "load_users": ${LOAD_USERS},
    "load_duration": ${LOAD_DURATION},
    "benchmark_samples": ${BENCHMARK_SAMPLES}
  },
  "tests": {}
}
EOF

# Track overall status
ALL_PASSED=true

# Function to run a test and capture results
run_test() {
    local test_name=$1
    local test_script=$2
    shift 2
    local test_args=("$@")
    
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Running: ${test_name}${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    local test_result_file="${RESULTS_DIR}/${test_name}_${TIMESTAMP}.json"
    
    if python3 "${test_script}" "${test_args[@]}" > "${test_result_file}" 2>&1; then
        echo -e "${GREEN}✓ ${test_name} PASSED${NC}"
        local status="PASS"
    else
        echo -e "${RED}✗ ${test_name} FAILED${NC}"
        local status="FAIL"
        ALL_PASSED=false
    fi
    
    echo ""
    
    # Add to report (simplified - just record status)
    echo "  ${test_name}: ${status}" >> "${RESULTS_DIR}/summary.txt"
}

# Initialize summary file
echo "Battle Test Results - $(date)" > "${RESULTS_DIR}/summary.txt"
echo "======================================" >> "${RESULTS_DIR}/summary.txt"
echo "" >> "${RESULTS_DIR}/summary.txt"

# Run benchmark test (quick, establishes baseline)
run_test "benchmark" \
    "${SCRIPT_DIR}/benchmark_test.py" \
    --samples "${BENCHMARK_SAMPLES}"

# Run stress test
run_test "stress" \
    "${SCRIPT_DIR}/stress_test.py" \
    --devices "${STRESS_DEVICES}" \
    --duration "${STRESS_DURATION}"

# Run load test
run_test "load" \
    "${SCRIPT_DIR}/load_test.py" \
    --concurrent-users "${LOAD_USERS}" \
    --duration "${LOAD_DURATION}"

# Run chaos test (only if requested via environment variable)
if [ "${RUN_CHAOS_TESTS}" = "true" ]; then
    echo -e "${YELLOW}⚠️  Chaos tests will disrupt services${NC}"
    echo "These tests will stop and restart Docker containers."
    echo "Press Ctrl+C to skip, or wait 5 seconds to continue..."
    sleep 5
    
    run_test "chaos" \
        "${SCRIPT_DIR}/chaos_test.py" \
        --scenarios all
else
    echo -e "${YELLOW}⚠️  Skipping chaos tests (set RUN_CHAOS_TESTS=true to enable)${NC}"
    echo "  Chaos tests disrupt services and should be run in isolated environments."
    echo ""
fi

# Print final summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Battle Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cat "${RESULTS_DIR}/summary.txt"
echo ""

echo "Detailed results saved to: ${RESULTS_DIR}/"
echo "Report file: ${REPORT_FILE}"
echo ""

if [ "${ALL_PASSED}" = true ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✅ ALL BATTLE TESTS PASSED${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}❌ SOME BATTLE TESTS FAILED${NC}"
    echo -e "${RED}========================================${NC}"
    exit 1
fi
