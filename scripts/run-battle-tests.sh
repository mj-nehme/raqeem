#!/bin/bash
# Battle Test Runner Script (relocated to scripts/)
# Runs all battle test suites and generates comprehensive report

set -e

# Usage/help
usage() {
    cat <<'USAGE'
Raqeem Battle Test Suite Runner

Usage:
    bash scripts/run-battle-tests.sh [options]

Options:
    --devices-url URL       Devices backend base URL (default: http://localhost:30080)
    --mentor-url URL        Mentor backend base URL (default: http://localhost:30090)
    --tests LIST            Comma-separated list of tests to run: benchmark,stress,load,chaos
                                                    Default: benchmark,stress,load (chaos disabled unless RUN_CHAOS_TESTS=true)
    --benchmark             Run only benchmark test
    --stress                Run only stress test
    --load                  Run only load test
    --include-chaos         Include chaos tests (or set RUN_CHAOS_TESTS=true)
    --quick                 Run a shorter, faster version of tests (reduced users/devices and duration)
    --scale FLOAT           Scale factor (0.1, 0.25, 0.5, 1.0) applied to users/devices and duration
    --keep-logs             Do not delete previous logs (by default old logs are cleaned)
    -h, --help              Show this help message and exit

Examples:
    bash scripts/run-battle-tests.sh                           # run all standard tests
    bash scripts/run-battle-tests.sh --tests benchmark,load    # select tests
    bash scripts/run-battle-tests.sh --devices-url http://localhost:30080 --mentor-url http://localhost:30090
    bash scripts/run-battle-tests.sh --quick                   # fast run with smaller load/duration
    bash scripts/run-battle-tests.sh --scale 0.25              # quarter load/duration

Notes:
    - Output streams live and saves detailed JSON logs under logs/battle.
    - On macOS, Python runs unbuffered for live streaming when stdbuf is unavailable.
USAGE
}

# Defaults and argument parsing
DEVICES_URL_DEFAULT="http://localhost:30080"
MENTOR_URL_DEFAULT="http://localhost:30090"
SELECTED_TESTS=()
INCLUDE_CHAOS=false
QUICK=false
SCALE=""
KEEP_LOGS=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --devices-url)
            DEVICES_URL=${2:-}
            shift 2
            ;;
        --mentor-url)
            MENTOR_URL=${2:-}
            shift 2
            ;;
        --tests)
            IFS=',' read -r -a SELECTED_TESTS <<< "${2:-}"
            shift 2
            ;;
        --benchmark)
            SELECTED_TESTS+=("benchmark")
            shift 1
            ;;
        --stress)
            SELECTED_TESTS+=("stress")
            shift 1
            ;;
        --load)
            SELECTED_TESTS+=("load")
            shift 1
            ;;
        --include-chaos)
            INCLUDE_CHAOS=true
            shift 1
            ;;
        --quick)
            QUICK=true
            shift 1
            ;;
        --scale)
            SCALE=${2:-}
            shift 2
            ;;
        --keep-logs)
            KEEP_LOGS=true
            shift 1
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            usage
            exit 2
            ;;
    esac
done

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BATTLE_DIR="${SCRIPT_DIR}/../tests/battle"
LOGS_DIR="${SCRIPT_DIR}/../logs/battle"
RESULTS_DIR="${LOGS_DIR}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="${RESULTS_DIR}/battle_test_report_${TIMESTAMP}.json"

# Test configuration (use smaller values for quicker runs)
STRESS_DEVICES=${STRESS_DEVICES:-100}
STRESS_DURATION=${STRESS_DURATION:-60}
LOAD_USERS=${LOAD_USERS:-50}
LOAD_DURATION=${LOAD_DURATION:-60}
BENCHMARK_SAMPLES=${BENCHMARK_SAMPLES:-20}

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Raqeem Battle Test Suite Runner${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Timestamp: $(date)"
echo "Results will be saved to: ${REPORT_FILE}"
echo ""

# Create logs/results directory (top-level logs/)
mkdir -p "${RESULTS_DIR}"

# By default, clean previous logs unless --keep-logs provided
if [[ "${KEEP_LOGS}" != "true" ]]; then
    if [[ -d "${RESULTS_DIR}" ]]; then
        find "${RESULTS_DIR}" -type f -name '*.json' -delete || true
        rm -f "${RESULTS_DIR}/summary.txt" || true
    fi
fi

DEVICES_URL="${DEVICES_URL:-${DEVICES_BACKEND_URL:-}}"
MENTOR_URL="${MENTOR_URL:-${MENTOR_BACKEND_URL:-}}"

# Prefer discovery registry if available
if [ -f .deploy/registry/devices-backend ]; then
    # shellcheck disable=SC1091
    source .deploy/registry/devices-backend
    DEVICES_URL="${URL}"
fi
if [ -f .deploy/registry/mentor-backend ]; then
    # shellcheck disable=SC1091
    source .deploy/registry/mentor-backend
    MENTOR_URL="${URL}"
fi

# Fallbacks to NodePorts
DEVICES_URL="${DEVICES_URL:-$DEVICES_URL_DEFAULT}"
MENTOR_URL="${MENTOR_URL:-$MENTOR_URL_DEFAULT}"

# Check if services are running
echo -e "${YELLOW}Checking service health...${NC}"
if ! curl -sf "${DEVICES_URL}/health" > /dev/null 2>&1; then
        echo -e "${RED}✗ Devices backend is not running at ${DEVICES_URL}${NC}"
        echo "Please start services first:"
        echo "  docker compose -f .github/docker-compose.test.yml up -d"
        echo "  OR"
        echo "  ./start.sh"
        exit 1
fi

if ! curl -sf "${MENTOR_URL}/health" > /dev/null 2>&1; then
        echo -e "${RED}✗ Mentor backend is not running at ${MENTOR_URL}${NC}"
        echo "Please start services first:"
        echo "  docker compose -f .github/docker-compose.test.yml up -d"
        echo "  OR"
        echo "  ./start.sh"
        exit 1
fi

echo -e "${GREEN}✓ Services are healthy${NC}"
echo ""

# Ensure results directory writable and pre-create report file
mkdir -p "${RESULTS_DIR}" && touch "${REPORT_FILE}" || {
    echo -e "${RED}✗ Unable to create report file at ${REPORT_FILE}${NC}"
}

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
        echo -e "${YELLOW}→ Streaming ${test_name} output (also saving to ${test_result_file})${NC}"

        # Detect stdbuf availability (GNU coreutils). On macOS it's often absent.
        if command -v stdbuf >/dev/null 2>&1; then
            STREAM_PREFIX="stdbuf -oL -eL"
            PY_UNBUFFER=""
        else
            STREAM_PREFIX=""
            # Force unbuffered python output to stream live
            PY_UNBUFFER="-u"
        fi

        ${STREAM_PREFIX} python3 ${PY_UNBUFFER} "${test_script}" "${test_args[@]}" 2>&1 | tee "${test_result_file}"
    local py_exit=${PIPESTATUS[0]}
    if [ "$py_exit" -eq 0 ]; then
        echo -e "${GREEN}✓ ${test_name} PASSED${NC}"
        local status="PASS"
    else
        echo -e "${RED}✗ ${test_name} FAILED (exit ${py_exit})${NC}"
        local status="FAIL"
        ALL_PASSED=false
    fi
    
    echo "" 
    echo "  ${test_name}: ${status}" >> "${RESULTS_DIR}/summary.txt"
}

# Initialize summary file
echo "Battle Test Results - $(date)" > "${RESULTS_DIR}/summary.txt"
echo "======================================" >> "${RESULTS_DIR}/summary.txt"
echo "" >> "${RESULTS_DIR}/summary.txt"

# Apply quick/scale settings to reduce load and duration
if [[ "$QUICK" == "true" ]]; then
    # Fast profile: ~10% load, shorter durations
    STRESS_DEVICES=$(( STRESS_DEVICES / 10 ))
    STRESS_DEVICES=$(( STRESS_DEVICES < 10 ? 10 : STRESS_DEVICES ))
    STRESS_DURATION=$(( STRESS_DURATION / 3 ))
    STRESS_DURATION=$(( STRESS_DURATION < 20 ? 20 : STRESS_DURATION ))

    LOAD_USERS=$(( LOAD_USERS / 10 ))
    LOAD_USERS=$(( LOAD_USERS < 10 ? 10 : LOAD_USERS ))
    LOAD_DURATION=$(( LOAD_DURATION / 3 ))
    LOAD_DURATION=$(( LOAD_DURATION < 20 ? 20 : LOAD_DURATION ))

    BENCHMARK_SAMPLES=$(( BENCHMARK_SAMPLES / 2 ))
    BENCHMARK_SAMPLES=$(( BENCHMARK_SAMPLES < 10 ? 10 : BENCHMARK_SAMPLES ))
    # In quick mode, reduce mentor query load and optionally skip it to avoid timeouts
    export BATTLE_QUERY_LIMIT=50
    export BATTLE_SKIP_MENTOR_QUERIES=true
fi

if [[ -n "$SCALE" ]]; then
    # Apply scale factor to numeric settings
    # Use awk for safe float multiplication
    STRESS_DEVICES=$(awk -v v="$STRESS_DEVICES" -v s="$SCALE" 'BEGIN{printf "%d", (v*s)<1?1:(v*s)}')
    STRESS_DURATION=$(awk -v v="$STRESS_DURATION" -v s="$SCALE" 'BEGIN{printf "%d", (v*s)<1?1:(v*s)}')
    LOAD_USERS=$(awk -v v="$LOAD_USERS" -v s="$SCALE" 'BEGIN{printf "%d", (v*s)<1?1:(v*s)}')
    LOAD_DURATION=$(awk -v v="$LOAD_DURATION" -v s="$SCALE" 'BEGIN{printf "%d", (v*s)<1?1:(v*s)}')
    BENCHMARK_SAMPLES=$(awk -v v="$BENCHMARK_SAMPLES" -v s="$SCALE" 'BEGIN{printf "%d", (v*s)<1?1:(v*s)}')
fi

# Determine tests to run: defaults to all if none selected
if [[ ${#SELECTED_TESTS[@]} -eq 0 ]]; then
    SELECTED_TESTS=("benchmark" "stress" "load")
fi

for test in "${SELECTED_TESTS[@]}"; do
    case "$test" in
        benchmark)
            run_test "benchmark" \
                "${BATTLE_DIR}/benchmark_test.py" \
                --samples "${BENCHMARK_SAMPLES}" \
                --devices-url "${DEVICES_URL}" \
                --mentor-url "${MENTOR_URL}"
            ;;
        stress)
            run_test "stress" \
                "${BATTLE_DIR}/stress_test.py" \
                --devices "${STRESS_DEVICES}" \
                --duration "${STRESS_DURATION}" \
                --devices-url "${DEVICES_URL}" \
                --mentor-url "${MENTOR_URL}"
            ;;
        load)
            run_test "load" \
                "${BATTLE_DIR}/load_test.py" \
                --concurrent-users "${LOAD_USERS}" \
                --duration "${LOAD_DURATION}" \
                --devices-url "${DEVICES_URL}" \
                --mentor-url "${MENTOR_URL}"
            ;;
        chaos)
            INCLUDE_CHAOS=true
            ;;
        *)
            echo -e "${YELLOW}Unknown test: ${test}${NC}"
            ;;
    esac
done

# Chaos optional
if [ "${RUN_CHAOS_TESTS}" = "true" ] || [ "${INCLUDE_CHAOS}" = "true" ]; then
    echo -e "${YELLOW}⚠️  Chaos tests will disrupt services${NC}"
    echo "These tests will stop and restart Docker containers."
    echo "Press Ctrl+C to skip, or wait 5 seconds to continue..."
    sleep 5
    run_test "chaos" \
        "${BATTLE_DIR}/chaos_test.py" \
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

# Brief report: show per-test status and overall pass rate
if [[ -f "${REPORT_FILE}" ]]; then
    BRIEF=$(python3 - "${REPORT_FILE}" <<'PY'
import json,sys
fp = sys.argv[1] if len(sys.argv) > 1 else None
if not fp:
        print("Brief Report\n============\n(no report file provided)")
else:
        try:
                with open(fp) as f:
                        data = json.load(f)
                tests = data.get('tests', {})
                total_ops = 0
                total_fails = 0
                lines = []
                for name, res in tests.items():
                        failed = res.get('failed', 0)
                        total = res.get('total', res.get('samples', res.get('total_sent', 0)))
                        total_ops += total
                        total_fails += failed
                        rate = (0 if total == 0 else (total - failed) * 100 / total)
                        lines.append(f"- {name}: {rate:.1f}% ({total - failed}/{total})")
                print("Brief Report")
                print("============")
                for l in lines:
                        print(l)
                overall = (0 if total_ops == 0 else (total_ops - total_fails) * 100 / total_ops)
                print(f"Overall: {overall:.1f}% ({total_ops - total_fails}/{total_ops})")
        except Exception as e:
                print("Brief Report\n============\n(error generating brief report:", e, ")")
PY
    )
    echo "$BRIEF"
fi
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
