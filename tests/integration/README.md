# Integration Tests Architecture

This directory contains comprehensive integration tests for the Raqeem IoT monitoring platform, covering all major communication patterns and system interactions.

## Overview

The integration tests validate real-world scenarios where different components of the system interact with each other, databases, and external services (MinIO/S3). Each test file focuses on a specific communication pattern and is designed to be run independently or as part of a comprehensive test suite.

## Test Files

### 1. test_devices_backend_db_s3.py

**Purpose**: Validates Devices Backend interactions with PostgreSQL and MinIO (S3)

**What it tests**:
- Device registration and storage in PostgreSQL
- Metrics data persistence
- Activity logging
- Alert storage
- Screenshot upload to MinIO (S3)

**Duration**: ~15-20 seconds

**Key validations**:
- Data is correctly stored in the database
- Screenshots are successfully uploaded to S3
- API responses are correct and complete

### 2. test_mentor_backend_db_s3.py

**Purpose**: Validates Mentor Backend interactions with PostgreSQL and MinIO (S3)

**What it tests**:
- Device listing from database
- Alert submission and retrieval
- Metrics retrieval
- Screenshot presigned URL generation

**Duration**: ~15-20 seconds

**Key validations**:
- Database queries return correct data
- Alert CRUD operations work correctly
- Presigned URLs are generated for S3 objects

### 3. test_backend_communication.py

**Purpose**: Validates communication between Devices Backend and Mentor Backend

**What it tests**:
- Device registration in devices backend
- Alert submission to devices backend
- Automatic alert forwarding from devices → mentor
- Alert retrieval from mentor backend
- Data consistency across backends

**Duration**: ~20-25 seconds

**Key validations**:
- Alerts are automatically forwarded
- Data remains consistent across services
- No data loss during forwarding

### 4. test_e2e_system_flow.py

**Purpose**: Validates complete end-to-end system workflows with multiple devices

**What it tests**:
- Multiple device registration and operation
- Normal device scenario (moderate load)
- Critical device scenario (high alerts)
- Screenshot uploads
- Cross-device data verification
- Complete data flow pipeline

**Duration**: ~30-40 seconds

**Key validations**:
- System handles multiple concurrent devices
- All data flows work together correctly
- Data from different devices remains isolated
- System performs under realistic load

### 5. test_alert_flow.py (Legacy)

**Purpose**: Original E2E alert pipeline test (maintained for backward compatibility)

**What it tests**:
- Complete alert submission and forwarding pipeline

**Duration**: ~15-20 seconds

## Test Infrastructure

### Docker Compose Setup

The tests use `docker-compose.test.yml` which defines:

- **PostgreSQL 14**: Shared database for both backends
- **MinIO**: S3-compatible object storage
- **Devices Backend**: FastAPI service (Python)
- **Mentor Backend**: Go/Gin service

All services include health checks to ensure they're ready before tests run.

### Test Runner Scripts

#### run_all_integration_tests.sh

Comprehensive test runner that:
1. Validates prerequisites (Docker, Python)
2. Cleans up any existing containers
3. Starts all services with Docker Compose
4. Waits for services to be healthy
5. Runs all integration tests in sequence
6. Provides detailed summary of results
7. Shows logs on failure
8. Offers option to keep services running

**Usage**:
```bash
./tests/integration/run_all_integration_tests.sh
```

#### run_integration_tests.sh (Legacy)

Runs only the original alert flow test. Maintained for backward compatibility.

**Usage**:
```bash
./tests/integration/run_integration_tests.sh
```

## Running Tests

### Run All Tests (Recommended)

```bash
cd /path/to/raqeem
./tests/integration/run_all_integration_tests.sh
```

Expected output:
```
==============================================================================
Raqeem Comprehensive Integration Test Suite
==============================================================================

✓ Prerequisites satisfied
✓ Services started
✓ All services are healthy

Running: Devices Backend ↔ DB & S3
✓ All Devices Backend ↔ DB/S3 tests passed!

Running: Mentor Backend ↔ DB & S3
✓ All Mentor Backend ↔ DB/S3 tests passed!

Running: Backend-to-Backend Communication
✓ All Backend-to-Backend communication tests passed!

Running: Alert Flow (Original)
✓ All integration tests passed!

Running: End-to-End System Flow
✓ All End-to-End System tests passed!

Total: 5 | Passed: 5 | Failed: 0
```

### Run Individual Tests

Start services first:
```bash
cd /path/to/raqeem
docker compose -f .github/docker-compose.test.yml up -d
```

Wait for services to be healthy (~30-60 seconds), then run any test:
```bash
python3 tests/integration/test_devices_backend_db_s3.py
python3 tests/integration/test_mentor_backend_db_s3.py
python3 tests/integration/test_backend_communication.py
python3 tests/integration/test_e2e_system_flow.py
```

Stop services:
```bash
docker compose -f .github/docker-compose.test.yml down -v
```

## Test Design Principles

### 1. Isolation

Each test creates unique device IDs using timestamps to avoid conflicts:
```python
TEST_DEVICE_ID = f"test-device-{int(time.time())}"
```

### 2. Clear Logging

All tests provide detailed, timestamped logging:
```
[14:23:45.123] [INFO] Waiting for Devices Backend at http://localhost:8081...
[14:23:46.456] [SUCCESS] ✓ Devices Backend is ready
[14:23:47.789] [INFO] Testing device registration...
[14:23:48.012] [SUCCESS] ✓ Device registered successfully
```

### 3. Comprehensive Validation

Tests validate not just success, but also:
- Correct HTTP status codes
- Response data structure
- Field values and types
- Data consistency across services

### 4. Realistic Data

Tests use realistic device data:
- Valid device IDs, names, types
- Realistic metrics (CPU, memory, disk)
- Meaningful alert messages
- Proper severity levels

### 5. Error Handling

All tests properly handle:
- Network timeouts
- Service unavailability
- Invalid responses
- Database errors

## Prerequisites

### Software Requirements

- Docker (with Docker Compose v2)
- Python 3.10+
- Python `requests` library

### Installation

```bash
# Install requests library
pip install requests

# Verify Docker
docker --version
docker compose version
```

## Troubleshooting

### Services Not Starting

```bash
# Check Docker is running
docker ps

# Clean up and retry
docker compose -f .github/docker-compose.test.yml down -v
docker system prune -f
./tests/integration/run_all_integration_tests.sh
```

### Services Not Healthy

Services can take 30-60 seconds to become healthy. If they still don't:

```bash
# Check service logs
docker compose -f .github/docker-compose.test.yml logs devices-backend
docker compose -f .github/docker-compose.test.yml logs mentor-backend

# Check if ports are already in use
lsof -i :8080  # Mentor backend
lsof -i :8081  # Devices backend
lsof -i :5432  # PostgreSQL
lsof -i :9000  # MinIO
```

### Tests Failing

```bash
# View detailed test output
python3 tests/integration/test_devices_backend_db_s3.py

# Check service health manually
curl http://localhost:8081/health  # Devices backend
curl http://localhost:8080/health  # Mentor backend

# Verify database connectivity
docker compose -f .github/docker-compose.test.yml exec postgres \
  psql -U monitor -d monitoring_db -c "SELECT 1;"
```

### Python Dependencies

```bash
# Install test dependencies
pip install requests

# Verify installation
python3 -c "import requests; print('requests:', requests.__version__)"
```

## CI/CD Integration

These tests are designed to run in CI/CD pipelines. Key considerations:

1. **Docker Availability**: CI environment must have Docker
2. **Port Availability**: Ports 8080, 8081, 5432, 9000 must be free
3. **Timeouts**: Allow 5-10 minutes for complete test suite
4. **Cleanup**: Tests clean up containers on completion

### GitHub Actions Example

```yaml
jobs:
  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Integration Tests
        run: |
          pip install requests
          ./tests/integration/run_all_integration_tests.sh
```

## Performance Expectations

| Test | Duration | Services Used |
|------|----------|---------------|
| test_devices_backend_db_s3 | 15-20s | Devices Backend, PostgreSQL, MinIO |
| test_mentor_backend_db_s3 | 15-20s | Mentor Backend, PostgreSQL, MinIO |
| test_backend_communication | 20-25s | Both Backends, PostgreSQL |
| test_e2e_system_flow | 30-40s | All services |
| test_alert_flow | 15-20s | Both Backends, PostgreSQL |
| **Total Suite** | **3-5 minutes** | All services |

## Extending the Tests

### Adding a New Test

1. Create a new test file in `tests/integration/`:
```python
#!/usr/bin/env python3
"""
Integration test: [Description]
"""

import sys
import time
import requests
from datetime import datetime

# Configuration
BACKEND_URL = "http://localhost:8080"
TEST_ID = f"test-{int(time.time())}"

def log(message, level="INFO"):
    timestamp = datetime.now().strftime("%H:%M:%S.%f")[:-3]
    print(f"[{timestamp}] [{level}] {message}")

# ... test functions ...

if __name__ == "__main__":
    success = run_test()
    sys.exit(0 if success else 1)
```

2. Add to `run_all_integration_tests.sh`:
```bash
run_test "tests/integration/test_my_feature.py" "My Feature Test"
```

3. Update documentation in:
   - `tests/README.md`
   - `docs/TESTING.md`
   - This file

### Best Practices

- Use unique IDs with timestamps
- Provide detailed logging
- Validate all response fields
- Handle errors gracefully
- Clean up test data
- Document expected behavior

## Related Documentation

- [tests/README.md](../README.md) - Quick start guide
- [docs/TESTING.md](../../docs/TESTING.md) - Complete testing guide
- [docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md) - System architecture
- [.github/docker-compose.test.yml](../../.github/docker-compose.test.yml) - Docker Compose configuration

## Contributing

When adding new integration tests:

1. Follow the existing test structure
2. Use descriptive test names
3. Provide comprehensive validation
4. Update all documentation
5. Test locally before committing
6. Ensure tests pass in CI

## License

MIT License - See repository root for details
