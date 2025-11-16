# Raqeem Tests

This directory contains integration and end-to-end tests for the Raqeem monitoring system.

## Quick Start

### Smoke Test (Fastest)

Test running services without docker-compose:

```bash
# Start services
./scripts/start.sh

# Run smoke test (in another terminal)
python3 tests/smoke_test.py
```

### Full Integration Test

Complete end-to-end test with docker-compose:

```bash
./tests/integration/run_integration_tests.sh
```

## Directory Structure

```
tests/
├── smoke_test.py                 # Quick health + alert flow test
├── integration/                  # Integration tests
│   ├── run_integration_tests.sh     # Integration test runner
│   ├── run_all_integration_tests.sh # Comprehensive test runner
│   ├── test_alert_flow.py           # E2E alert pipeline test
│   ├── test_backend_communication.py
│   ├── test_devices_backend_db_s3.py
│   ├── test_mentor_backend_db_s3.py
│   └── test_e2e_system_flow.py
├── battle/                       # Battle tests (stress, load, chaos)
│   ├── README.md                    # Battle test documentation
│   ├── run_battle_tests.sh          # Battle test runner
│   ├── stress_test.py               # High-volume stress testing
│   ├── load_test.py                 # Sustained load testing
│   ├── chaos_test.py                # Chaos engineering tests
│   ├── benchmark_test.py            # Performance benchmarking
│   └── requirements.txt             # Battle test dependencies
└── README.md                     # This file
```

## What Each Test Does

### smoke_test.py

**Purpose**: Quick validation of running services  
**Duration**: ~10 seconds  
**Requirements**: Services must be running

Tests:
- ✓ Health checks for both backends
- ✓ Device registration
- ✓ Alert submission
- ✓ Alert forwarding to mentor backend
- ✓ Alert retrieval from mentor backend

Usage:
```bash
# With default URLs (localhost:8081, localhost:8080)
python3 tests/smoke_test.py

# With custom URLs
python3 tests/smoke_test.py http://devices:8081 http://mentor:8080
```

### integration/test_alert_flow.py

**Purpose**: Comprehensive end-to-end test  
**Duration**: ~30-60 seconds (includes service startup)  
**Requirements**: Docker and docker-compose

Tests:
- ✓ Service health and readiness
- ✓ Device registration with full payload
- ✓ Alert creation with all fields
- ✓ Alert storage in devices backend
- ✓ Alert forwarding to mentor backend
- ✓ Alert storage in mentor backend
- ✓ Alert retrieval with correct data
- ✓ Field validation (device_id, level, type, value, threshold)

Usage:
```bash
# Automated (via runner script)
./tests/integration/run_integration_tests.sh

# Manual
docker-compose -f .github/docker-compose.test.yml up -d
python3 tests/integration/test_alert_flow.py
docker-compose -f .github/docker-compose.test.yml down -v
```

### Battle Tests (Production Readiness)

**Purpose**: Comprehensive stress, load, and chaos testing for production confidence  
**Duration**: ~30-120 minutes (configurable)  
**Requirements**: Docker and docker-compose, battle test dependencies

Tests:
- 🔥 **Stress Testing**: 1000+ concurrent devices, high-volume ingestion
- 📊 **Load Testing**: Sustained operations, frontend concurrency
- 💥 **Chaos Engineering**: Service failures, database disruptions, network issues
- ⚡ **Performance Benchmarking**: Baseline metrics, regression detection

Usage:
```bash
# Install dependencies
pip install -r tests/battle/requirements.txt

# Run all battle tests (quick version)
./tests/battle/run_battle_tests.sh

# Run individual tests
python3 tests/battle/stress_test.py --devices 1000 --duration 300
python3 tests/battle/load_test.py --concurrent-users 100 --duration 300
python3 tests/battle/benchmark_test.py --samples 1000

# Run chaos tests (disruptive - will restart services)
RUN_CHAOS_TESTS=true python3 tests/battle/chaos_test.py --scenarios all
```

See [battle/README.md](battle/README.md) for detailed documentation.

### Comprehensive Integration Test Suite

**Purpose**: Complete test coverage for all system communication patterns  
**Duration**: ~3-5 minutes (includes all tests)  
**Requirements**: Docker and docker-compose

The comprehensive test suite includes:

#### 1. test_devices_backend_db_s3.py
Tests Devices Backend communication with PostgreSQL and MinIO:
- ✓ Device registration (DB write)
- ✓ Metrics storage (DB write)
- ✓ Activity logging (DB write)
- ✓ Alert storage (DB write)
- ✓ Screenshot upload to S3 (MinIO)

#### 2. test_mentor_backend_db_s3.py
Tests Mentor Backend communication with PostgreSQL and MinIO:
- ✓ Device listing (DB read)
- ✓ Alert submission and storage (DB write/read)
- ✓ Metrics retrieval (DB read)
- ✓ Screenshot presigned URL generation (S3)

#### 3. test_backend_communication.py
Tests Backend-to-Backend communication (alert forwarding):
- ✓ Device registration in devices backend
- ✓ Alert submission to devices backend
- ✓ Automatic forwarding to mentor backend
- ✓ Alert retrieval from mentor backend
- ✓ Data consistency verification

#### 4. test_e2e_system_flow.py
Tests complete end-to-end system workflows:
- ✓ Multiple device scenarios
- ✓ Normal device operation with metrics/activities/alerts
- ✓ Critical device with multiple alerts
- ✓ Screenshot uploads
- ✓ Cross-device data verification

Usage:
```bash
# Run all tests (recommended)
./tests/integration/run_all_integration_tests.sh

# Run individual tests
docker-compose -f .github/docker-compose.test.yml up -d
python3 tests/integration/test_devices_backend_db_s3.py
python3 tests/integration/test_mentor_backend_db_s3.py
python3 tests/integration/test_backend_communication.py
python3 tests/integration/test_e2e_system_flow.py
```


## Prerequisites

### All Tests

```bash
pip install requests
```

### Integration Tests

```bash
# macOS
brew install docker docker-compose

# Linux
apt-get install docker.io docker-compose

# Verify
docker --version
docker-compose --version
```

## Expected Output

### Successful Smoke Test

```
[13:45:23] ℹ️  Raqeem Smoke Test
[13:45:23] ℹ️  Devices Backend: http://localhost:8081
[13:45:23] ℹ️  Mentor Backend:  http://localhost:8080

[13:45:23] ℹ️  Checking service health...
[13:45:23] ✓ Devices Backend is healthy: {'status': 'ok', 'service': 'devices-backend'}
[13:45:23] ✓ Mentor Backend is healthy: {'status': 'ok', 'service': 'mentor-backend'}

[13:45:23] ℹ️  Testing alert flow...
[13:45:23] ℹ️  Registering test device...
[13:45:23] ✓ Device registered
[13:45:24] ℹ️  Sending test alert...
[13:45:24] ✓ Alert sent
[13:45:24] ℹ️  Checking alert in mentor backend...
[13:45:26] ✓ Alert found in mentor backend

[13:45:26] ✓ All smoke tests passed!
```

### Successful Integration Test

```
[13:30:24] [INFO] Starting E2E Integration Test
[13:30:24] [INFO] Waiting for Devices Backend at http://localhost:8081...
[13:30:26] [INFO] ✓ Devices Backend is ready
[13:30:26] [INFO] Waiting for Mentor Backend at http://localhost:8080...
[13:30:28] [INFO] ✓ Mentor Backend is ready
[13:30:28] [INFO] Registering device test-device-1730471428...
[13:30:28] [INFO] ✓ Device registered: {'device_id': 'test-device-1730471428', 'created': True}
[13:30:28] [INFO] Sending alert...
[13:30:28] [INFO] ✓ Alert sent: {'inserted': 1}
[13:30:28] [INFO] Verifying alert in mentor backend...
[13:30:30] [INFO] ✓ Alert found in mentor backend
[13:30:30] [INFO]   ✓ device_id matches
[13:30:30] [INFO]   ✓ level is warning
[13:30:30] [INFO]   ✓ type is cpu_high
[13:30:30] [INFO]   ✓ value is 95.5
[13:30:30] [INFO]   ✓ threshold is 80.0
[13:30:30] [SUCCESS] ✓ All integration tests passed!
```

## Troubleshooting

### "Services are not reachable"

Make sure services are running:
```bash
./scripts/start.sh

# Or with docker-compose
docker-compose -f .github/docker-compose.test.yml up -d
```

### "Alert was not found in mentor backend"

1. Check mentor backend logs:
```bash
docker-compose -f .github/docker-compose.test.yml logs mentor-backend
```

2. Verify `MENTOR_API_URL` is set in devices backend
3. Check network connectivity between services

### "Failed to start services"

```bash
# Clean up and retry
docker-compose -f .github/docker-compose.test.yml down -v
docker system prune -a  # Warning: removes all unused containers/images
./tests/integration/run_integration_tests.sh
```

## CI Integration

These tests run automatically in GitHub Actions on every push/PR.

See `.github/workflows/ci.yml` for the CI configuration.

To run CI locally:
```bash
brew install act  # macOS
act -j build-and-test
```

## See Also

- [TESTING.md](../docs/TESTING.md) - Complete testing guide
- [LOCAL_CI.md](../docs/LOCAL_CI.md) - Running GitHub Actions locally
- [README.md](../README.md) - Project overview
