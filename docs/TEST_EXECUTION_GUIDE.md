# Test Execution Guide

Complete guide for running all tests in the Raqeem IoT monitoring platform.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Test Types](#test-types)
4. [Execution Procedures](#execution-procedures)
5. [CI/CD Integration](#cicd-integration)
6. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Software Requirements

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| Docker | 20.10+ | Container runtime | [docker.com](https://docker.com) |
| Docker Compose | v2.0+ | Multi-container orchestration | Included with Docker Desktop |
| Python | 3.11+ | Test execution | [python.org](https://python.org) |
| requests | 2.31+ | Python HTTP library | `pip install requests` |

### Database Setup (for Unit Tests)

Unit tests require PostgreSQL:

```bash
# Using Docker (recommended)
docker run --name raqeem-test-postgres \
  -e POSTGRES_USER=monitor \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=monitoring_db \
  -p 5432:5432 \
  -d docker.io/library/postgres:16

# Or create locally
createuser -P monitor  # Password: password
createdb -O monitor monitoring_db
```

---

## Quick Start

### For Developers (Fast Validation)

```bash
# 1. Smoke test (10 seconds)
./scripts/start.sh
python3 tests/smoke_test.py

# 2. Integration tests (3-5 minutes)
./tests/integration/run_all_integration_tests.sh
```

### For Release Validation (Complete)

```bash
# 1. Install dependencies
pip install -r tests/battle/requirements.txt

# 2. Run all tests
./tests/integration/run_all_integration_tests.sh  # Integration
./tests/battle/run_battle_tests.sh                # Battle tests
```

---

## Test Types

### Test Pyramid

```
         /\
        /  \    Battle Tests (Stress, Load, Chaos, Benchmark)
       /----\
      /      \  E2E Integration Tests (System Workflows)
     /--------\
    /          \ Component Integration Tests (DB, S3)
   /------------\
  /______________\ Unit Tests (Individual Components)
```

### Test Categories

| Category | Duration | Frequency | Purpose |
|----------|----------|-----------|---------|
| **Unit Tests** | Seconds | Every save | Validate individual functions |
| **Component Tests** | Minutes | Every commit | Validate service interactions |
| **Integration Tests** | 3-5 min | Every PR | Validate end-to-end flows |
| **Smoke Tests** | 10 sec | Post-deployment | Validate deployment health |
| **Battle Tests** | 30-120 min | Pre-release | Validate production readiness |

---

## Execution Procedures

### 1. Unit Tests

Run individual component tests during development.

#### Devices Backend (Python/FastAPI)

```bash
cd devices/backend/src

# Run all tests
pytest -v

# Run with coverage
pytest --cov=app --cov-report=html -v

# Run specific test
pytest tests/api/test_alerts_forwarding.py -v

# Run marker-specific tests
pytest -m unit -v
```

#### Mentor Backend (Go/Gin)

```bash
cd mentor/backend/src

# Set environment variables
export POSTGRES_USER=monitor
export POSTGRES_PASSWORD=password
export POSTGRES_DB=monitoring_db
export POSTGRES_HOST=127.0.0.1
export POSTGRES_PORT=5432

# Run all tests
go test ./... -v

# Run with coverage
go test ./... -v -race -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

#### Frontends (React/Vitest)

```bash
# Devices frontend
cd devices/frontend
npm run test

# Mentor frontend
cd mentor/frontend
npm run test

# With coverage
npm run test:coverage
```

---

### 2. Smoke Tests

**Purpose**: Quick validation of deployed services  
**Duration**: ~10 seconds  
**When**: Post-deployment, pre-release, health checks

#### Prerequisites
Services must be running:

```bash
# Option 1: Local development
./scripts/start.sh

# Option 2: Docker Compose
docker compose -f .github/docker-compose.test.yml up -d
```

#### Execution

```bash
# With default URLs (localhost)
python3 tests/smoke_test.py

# With custom URLs
python3 tests/smoke_test.py http://devices:8081 http://mentor:8080
```

#### Expected Output

```
==============================================================
[13:45:23] ℹ️  Raqeem Smoke Test
==============================================================
[13:45:23] ℹ️  Devices Backend: http://localhost:8081
[13:45:23] ℹ️  Mentor Backend:  http://localhost:8080

[13:45:23] ℹ️  Checking service health...
[13:45:23] ✓ Devices Backend is healthy
[13:45:23] ✓ Mentor Backend is healthy

[13:45:23] ℹ️  Testing alert flow...
[13:45:26] ✓ Alert found in mentor backend

==============================================================
[13:45:26] ✓ All smoke tests passed!
==============================================================
```

---

### 3. Integration Tests

**Purpose**: Validate end-to-end system workflows  
**Duration**: 3-5 minutes  
**When**: Before every commit/PR, pre-release

#### Automated Execution (Recommended)

```bash
# Run all integration tests
./tests/integration/run_all_integration_tests.sh
```

This will:
1. Validate prerequisites (Docker, Python, requests)
2. Clean up any existing containers
3. Start services with Docker Compose
4. Wait for services to be healthy
5. Run all 6 integration test suites
6. Provide detailed summary
7. Show logs on failure

#### Manual Execution

```bash
# 1. Start services
docker compose -f .github/docker-compose.test.yml up -d

# 2. Wait for health (30-60 seconds)
# Check with: docker compose -f .github/docker-compose.test.yml ps

# 3. Run individual tests
python3 tests/integration/test_devices_backend_db_s3.py
python3 tests/integration/test_mentor_backend_db_s3.py
python3 tests/integration/test_backend_communication.py
python3 tests/integration/test_e2e_system_flow.py
python3 tests/integration/test_observability_features.py
python3 tests/integration/test_alert_flow.py

# 4. Cleanup
docker compose -f .github/docker-compose.test.yml down -v
```

#### Test Coverage

| Test | Duration | Coverage |
|------|----------|----------|
| **test_devices_backend_db_s3.py** | 15-20s | Device registration, metrics, activities, alerts, screenshots |
| **test_mentor_backend_db_s3.py** | 15-20s | Device listing, alert CRUD, metrics, presigned URLs |
| **test_backend_communication.py** | 20-25s | Inter-service communication, alert forwarding |
| **test_e2e_system_flow.py** | 30-40s | Complete workflows, multiple devices |
| **test_observability_features.py** | 10-15s | Health checks, request tracing |
| **test_alert_flow.py** | 15-20s | Alert pipeline (legacy) |

#### Expected Output

```
==============================================================================
Raqeem Comprehensive Integration Test Suite
==============================================================================

[14:23:45] ℹ️  Checking prerequisites...
[14:23:45] ✓ Docker is available
[14:23:45] ✓ Python requests library is installed

[14:23:46] ℹ️  Starting services...
[14:23:50] ✓ Services started successfully

[14:24:20] ✓ All services are healthy

Running Test 1/6: Devices Backend ↔ DB & S3
[14:24:21] ✓ All Devices Backend ↔ DB/S3 tests passed!

Running Test 2/6: Mentor Backend ↔ DB & S3
[14:24:36] ✓ All Mentor Backend ↔ DB/S3 tests passed!

Running Test 3/6: Backend-to-Backend Communication
[14:24:56] ✓ All Backend-to-Backend communication tests passed!

Running Test 4/6: Alert Flow (Original)
[14:25:11] ✓ All integration tests passed!

Running Test 5/6: End-to-End System Flow
[14:25:51] ✓ All End-to-End System tests passed!

Running Test 6/6: Observability Features
[14:26:06] ✓ All observability feature tests passed!

==============================================================================
Test Summary
==============================================================================
Total: 6 | Passed: 6 | Failed: 0
Duration: 3 minutes 21 seconds
```

---

### 4. Battle Tests

**Purpose**: Production readiness validation  
**Duration**: 30-120 minutes (configurable)  
**When**: Pre-release, nightly, performance validation

#### Prerequisites

```bash
# Install battle test dependencies
pip install -r tests/battle/requirements.txt

# Start services
docker compose -f .github/docker-compose.test.yml up -d
```

#### Quick Execution (All Tests)

```bash
# Run all battle tests with default settings
./tests/battle/run_battle_tests.sh
```

#### Individual Test Execution

##### Stress Test
High-volume testing to find breaking points:

```bash
# Default (100 devices, 60 seconds)
python3 tests/battle/stress_test.py

# Full production stress (1000 devices, 5 minutes)
python3 tests/battle/stress_test.py --devices 1000 --duration 300

# Quick test (10 devices, 30 seconds)
python3 tests/battle/stress_test.py --devices 10 --duration 30 --verbose
```

**Tests**:
- 1000+ device registration
- Continuous telemetry ingestion
- Concurrent alert generation
- Bulk screenshot uploads
- Database performance under load

##### Load Test
Sustained load to validate normal operations:

```bash
# Default (10 concurrent users, 60 seconds)
python3 tests/battle/load_test.py

# Production load (100 users, 5 minutes)
python3 tests/battle/load_test.py --concurrent-users 100 --duration 300

# With verbose output
python3 tests/battle/load_test.py --concurrent-users 50 --duration 120 --verbose
```

**Tests**:
- Continuous device operation simulation
- Frontend API concurrent access
- Alert forwarding pipeline under load
- Resource utilization monitoring

##### Benchmark Test
Performance baselines and regression detection:

```bash
# Default (100 samples)
python3 tests/battle/benchmark_test.py

# Comprehensive (1000 samples)
python3 tests/battle/benchmark_test.py --samples 1000

# Quick baseline (10 samples)
python3 tests/battle/benchmark_test.py --samples 10 --verbose
```

**Metrics**:
- Device registration latency (p50, p95, p99)
- Telemetry ingestion throughput
- Alert forwarding latency
- Database query response times
- Screenshot upload/download times

##### Chaos Test
Failure scenario validation (disruptive):

```bash
# Run all chaos scenarios (restarts services!)
RUN_CHAOS_TESTS=true python3 tests/battle/chaos_test.py --scenarios all

# Run specific scenario
RUN_CHAOS_TESTS=true python3 tests/battle/chaos_test.py --scenarios service_restart

# Available scenarios:
# - service_restart: Service crash and recovery
# - database_disruption: Database connection failures
# - storage_failure: MinIO/S3 unavailability
# - all: Run all scenarios
```

**Tests**:
- Service crash and recovery
- Database connection failures
- Storage (MinIO/S3) failures
- Partial service degradation

#### Performance Targets

| Metric | Target | Acceptable |
|--------|--------|------------|
| Device registration | <100ms p95 | <200ms p95 |
| Telemetry ingestion | >1000 msg/sec | >500 msg/sec |
| Alert forwarding | <500ms p95 | <1s p95 |
| Database queries | <50ms p95 | <100ms p95 |
| Screenshot upload | <2s for 1MB | <5s for 1MB |
| API response time | <200ms p95 | <500ms p95 |

---

## CI/CD Integration

### GitHub Actions Workflows

#### Integration Tests (Automatic)
Runs on every push/PR:

```yaml
# Implied workflow in .github/workflows/ci.yml
jobs:
  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Integration Tests
        run: |
          pip install requests
          ./tests/integration/run_all_integration_tests.sh
```

#### Battle Tests (Manual/Scheduled)
**Workflow**: `.github/workflows/battle-test.yml`

Triggers:
- Manual workflow dispatch
- Scheduled (nightly at 2 AM UTC)
- Push to release branches

```bash
# Trigger manually with GitHub CLI
gh workflow run battle-test.yml \
  -f stress_devices=1000 \
  -f duration=300 \
  -f run_chaos=false
```

### Local CI Testing

```bash
# Install act (GitHub Actions local runner)
brew install act  # macOS
# or: curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | bash

# Run CI locally
act -j build-and-test
```

---

## Troubleshooting

### Common Issues

#### Services Not Starting

**Symptom**: Docker Compose fails to start services

**Solutions**:
```bash
# Check Docker is running
docker ps

# Clean up and retry
docker compose -f .github/docker-compose.test.yml down -v
docker system prune -f
./tests/integration/run_all_integration_tests.sh
```

#### Services Not Healthy

**Symptom**: Tests timeout waiting for services

**Solutions**:
```bash
# Check service logs
docker compose -f .github/docker-compose.test.yml logs devices-backend
docker compose -f .github/docker-compose.test.yml logs mentor-backend

# Check if ports are in use
lsof -i :8080  # Mentor backend
lsof -i :8081  # Devices backend
lsof -i :5432  # PostgreSQL
lsof -i :9000  # MinIO

# Increase wait time
# Services can take 30-60 seconds to become healthy
```

#### PostgreSQL Not Available (Unit Tests)

**Symptom**: Unit tests fail with connection errors

**Solutions**:
```bash
# Start PostgreSQL container
docker run --name raqeem-test-postgres \
  -e POSTGRES_USER=monitor \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=monitoring_db \
  -p 5432:5432 \
  -d docker.io/library/postgres:16

# Verify connection
psql -h localhost -U monitor -d monitoring_db -c "SELECT 1;"
```

#### Python Dependencies Missing

**Symptom**: `ModuleNotFoundError: No module named 'requests'`

**Solutions**:
```bash
# Install test dependencies
pip install requests

# For battle tests
pip install -r tests/battle/requirements.txt

# Verify installation
python3 -c "import requests; print('requests:', requests.__version__)"
```

#### Tests Fail Intermittently

**Symptom**: Tests pass sometimes, fail other times

**Causes & Solutions**:
1. **Timing Issues**: Increase wait times in tests
2. **Resource Constraints**: Close other applications
3. **Port Conflicts**: Kill processes using required ports
4. **Docker Issues**: Restart Docker daemon

```bash
# Find and kill processes on ports
lsof -ti :8080 | xargs kill -9
lsof -ti :8081 | xargs kill -9

# Restart Docker (macOS)
osascript -e 'quit app "Docker"'
open -a Docker

# Clean Docker state
docker system prune -a -f --volumes
```

---

## Testing Checklist (Before Release)

Use this checklist before any release:

### Development Phase
- [ ] All unit tests pass
  - [ ] Devices backend: `cd devices/backend/src && pytest -v`
  - [ ] Mentor backend: `cd mentor/backend/src && go test ./... -v`
  - [ ] Devices frontend: `cd devices/frontend && npm run test`
  - [ ] Mentor frontend: `cd mentor/frontend && npm run test`

### Pre-Commit
- [ ] Smoke test passes: `python3 tests/smoke_test.py`
- [ ] Code formatted/linted
- [ ] No hardcoded credentials

### Pre-PR
- [ ] All integration tests pass: `./tests/integration/run_all_integration_tests.sh`
- [ ] No new warnings in logs
- [ ] Documentation updated

### Pre-Release
- [ ] Battle tests pass: `./tests/battle/run_battle_tests.sh`
- [ ] Performance targets met
- [ ] Manual verification complete
- [ ] CHANGELOG.md updated

---

## Additional Resources

- [tests/README.md](../tests/README.md) - Quick start guide
- [tests/integration/README.md](../tests/integration/README.md) - Integration test details
- [tests/battle/README.md](../tests/battle/README.md) - Battle test details
- [TESTING.md](./TESTING.md) - Complete testing guide

---

**Last Updated**: 2025-11-18  
**Version**: v0.2.0
