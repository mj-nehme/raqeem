# Raqeem Testing Guide

Complete guide to testing the Raqeem monitoring system for reliability and correctness.

## Test Pyramid

```
            /\
           /  \    Battle Tests (Production Stress/Load/Chaos)
          /----\
         /      \  E2E Integration Tests (Comprehensive Workflows)
        /--------\
       /          \  Integration/API Tests (Component Communication)
      /------------\
     /______________\ Unit Tests (Fastest, Most Focused)
```

## Quick Start

### Prerequisites

**All tests require PostgreSQL**. Tests use transaction-based isolation to ensure no data persists between test runs.

#### PostgreSQL Setup

Tests expect PostgreSQL to be available with these credentials:
- **User**: `monitor`
- **Password**: `password`
- **Database**: `monitoring_db`
- **Host**: `127.0.0.1` (or `localhost`)
- **Port**: `5432`

**Using Docker:**
```bash
docker run --name raqeem-test-postgres \
  -e POSTGRES_USER=monitor \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=monitoring_db \
  -p 5432:5432 \
  -d postgres:16
```

**Local PostgreSQL:**
```sql
CREATE USER monitor WITH PASSWORD 'password';
CREATE DATABASE monitoring_db OWNER monitor;
```

### 1. Unit Tests (Development)

Run individual component tests during development:

**Python (Devices Backend)**
```bash
cd devices/backend/src
pytest tests/api/test_alerts_forwarding.py -v
```

**Go (Mentor Backend)**
```bash
cd mentor/backend/src

# Set environment variables
export POSTGRES_USER=monitor
export POSTGRES_PASSWORD=password
export POSTGRES_DB=monitoring_db
export POSTGRES_HOST=127.0.0.1
export POSTGRES_PORT=5432

# Run tests
go test ./... -v

# Run tests with coverage
go test ./... -v -race -coverprofile=coverage.out
```

**Note**: Go tests use transaction-based isolation. All database changes are automatically rolled back after each test completes.

**Frontend (React)**
```bash
# Devices frontend
cd devices/frontend
npm run test

# Mentor frontend
cd mentor/frontend
npm run test
```

### 2. Smoke Test (Quick Validation)

Test running services without docker-compose:

```bash
# Start services first
./scripts/start.sh

# In another terminal
python3 tests/smoke_test.py
```

This validates:
- ✓ Services are reachable
- ✓ Health checks pass
- ✓ Alert flow works end-to-end

### 3. Integration Tests (Pre-Deployment)

Full end-to-end tests with docker-compose covering all system communication patterns:

```bash
# Run comprehensive integration test suite (recommended)
./tests/integration/run_all_integration_tests.sh

# Or run legacy single test
./tests/integration/run_integration_tests.sh
```

The comprehensive test suite includes:
- **Devices Backend ↔ DB & S3**: Device registration, metrics, activities, alerts, screenshot uploads
- **Mentor Backend ↔ DB & S3**: Device listing, alert storage/retrieval, presigned URLs
- **Backend-to-Backend Communication**: Alert forwarding pipeline and data consistency
- **End-to-End System Flow**: Complete workflows with multiple devices and scenarios

This:
- Starts Postgres, MinIO, and both backends in Docker
- Runs all integration tests systematically
- Validates data persistence across components
- Shows detailed logs on failure
- Provides summary of test results

### 4. Battle Tests (Production Readiness)

Comprehensive stress, load, and chaos testing for production confidence:

```bash
# Install battle test dependencies
pip install -r tests/battle/requirements.txt

# Start services
docker compose -f .github/docker-compose.test.yml up -d

# Run all battle tests (takes 30-120 minutes)
./tests/battle/run_battle_tests.sh

# Or run individual tests
python3 tests/battle/stress_test.py --devices 1000 --duration 300
python3 tests/battle/load_test.py --concurrent-users 100 --duration 300
python3 tests/battle/benchmark_test.py --samples 1000

# Chaos tests (disruptive - restarts services)
RUN_CHAOS_TESTS=true python3 tests/battle/chaos_test.py --scenarios all
```

**What Battle Tests Cover:**
- 🔥 **Stress Testing**: 1000+ concurrent devices, high-volume ingestion, database stress
- 📊 **Load Testing**: Sustained operations, frontend concurrency, alert pipeline
- 💥 **Chaos Engineering**: Service failures, database disruptions, storage failures
- ⚡ **Performance Benchmarking**: Latency baselines, throughput metrics, regression detection

**Performance Targets:**
- Device registration: <200ms p95
- Telemetry ingestion: >500 msg/sec
- Alert forwarding: <1s p95
- Database queries: <100ms p95
- Screenshot upload: <5s for 1MB

See [tests/battle/README.md](../tests/battle/README.md) for detailed documentation.

### 5. CI Tests (GitHub Actions)

Automatically runs on every push/PR. To run locally:

```bash
# Install act (GitHub Actions runner)
brew install act  # macOS

# Run CI locally
act -j build-and-test
```

## Test Coverage

### Devices Backend (Python/FastAPI)

| Test | What It Validates | File |
|------|------------------|------|
| `test_post_alerts_is_saved_and_forwarded` | Alert storage + forwarding to mentor | `devices/backend/src/tests/api/test_alerts_forwarding.py` |
| `test_db_connection` | Database connectivity | `devices/backend/src/tests/test_config.py` |

### Mentor Backend (Go/Gin)

| Test | What It Validates | File |
|------|------------------|------|
| `TestReportAndGetAlerts` | Alert creation and retrieval | `mentor/backend/src/controllers/device_test.go` |

### Frontends (React/Vitest)

| Test | What It Validates | File |
|------|------------------|------|
| `renders Device Simulator` | Device simulator UI | `devices/frontend/src/components/DeviceSimulator.test.jsx` |
| `renders Devices list` | Mentor dashboard UI | `mentor/frontend/src/components/DeviceDashboard.test.jsx` |

### Integration Tests

| Test | What It Validates | File |
|------|------------------|------|
| `test_devices_backend_db_s3` | **Devices Backend ↔ DB & S3**:<br>1. Device registration (DB)<br>2. Metrics storage (DB)<br>3. Activity logging (DB)<br>4. Alert storage (DB)<br>5. Screenshot upload (S3) | `tests/integration/test_devices_backend_db_s3.py` |
| `test_mentor_backend_db_s3` | **Mentor Backend ↔ DB & S3**:<br>1. Device listing (DB)<br>2. Alert submission and retrieval (DB)<br>3. Metrics retrieval (DB)<br>4. Screenshot presigned URLs (S3) | `tests/integration/test_mentor_backend_db_s3.py` |
| `test_backend_communication` | **Backend-to-Backend Communication**:<br>1. Device registration<br>2. Alert submission to devices backend<br>3. Automatic forwarding to mentor<br>4. Data consistency verification | `tests/integration/test_backend_communication.py` |
| `test_e2e_system_flow` | **End-to-End System Flow**:<br>1. Multiple device scenarios<br>2. Normal and critical operations<br>3. Complete data flow pipeline<br>4. Cross-device verification | `tests/integration/test_e2e_system_flow.py` |
| `test_alert_flow` | **Alert Pipeline (Legacy)**:<br>1. Device registration<br>2. Alert submission<br>3. Storage in devices DB<br>4. Forwarding to mentor<br>5. Storage in mentor DB<br>6. Retrieval from mentor API | `tests/integration/test_alert_flow.py` |

### Battle Tests

| Test | What It Validates | File |
|------|------------------|------|
| `stress_test` | **High-Volume Stress Testing**:<br>1. 1000+ device registration<br>2. Continuous telemetry ingestion<br>3. Concurrent alert generation<br>4. Bulk screenshot uploads<br>5. Database query performance | `tests/battle/stress_test.py` |
| `load_test` | **Sustained Load Testing**:<br>1. Device lifecycle simulation<br>2. Frontend API concurrent access<br>3. Alert pipeline under load<br>4. Resource utilization | `tests/battle/load_test.py` |
| `chaos_test` | **Chaos Engineering**:<br>1. Service restart and recovery<br>2. Database disruption<br>3. Storage failures<br>4. Concurrent disruptions | `tests/battle/chaos_test.py` |
| `benchmark_test` | **Performance Benchmarking**:<br>1. Latency percentiles (p50, p95, p99)<br>2. Throughput metrics<br>3. Concurrent operations<br>4. Regression detection | `tests/battle/benchmark_test.py` |

## Testing Checklist (Before Release)

```bash
# 1. Run all unit tests
cd devices/backend/src && pytest -v
cd ../../mentor/backend/src && go test ./... -v
cd ../../devices/frontend && npm run test -- --run
cd ../mentor/frontend && npm run test -- --run

# 2. Run integration tests
cd ../..
./tests/integration/run_all_integration_tests.sh

# 3. Run smoke test on running system
./scripts/start.sh  # In one terminal
python3 tests/smoke_test.py  # In another terminal

# 4. Run battle tests (production readiness)
pip install -r tests/battle/requirements.txt
docker compose -f .github/docker-compose.test.yml up -d
./tests/battle/run_battle_tests.sh

# 5. Manual verification (optional)
# - Open device simulator: http://localhost:14000
# - Register a device
# - Send an alert
# - Open mentor dashboard: http://localhost:15000
# - Verify alert appears in the Alerts tab
```

## Reliability Features

### Health Checks

Both backends expose `/health` endpoints:

```bash
# Check devices backend
curl http://localhost:8081/health

# Check mentor backend
curl http://localhost:8080/health
```

Returns:
```json
{"status": "ok", "service": "devices-backend"}
```

### Service Dependencies

Services use health checks in docker-compose:

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 10s
  timeout: 5s
  retries: 3
```

### Error Handling

- **Devices Backend**: Swallows mentor forwarding errors (doesn't block ingestion)
- **Mentor Backend**: Validates all inputs, returns proper HTTP status codes
- **Frontends**: Gracefully handle API failures with user feedback

## Troubleshooting

### Unit Tests Fail

**Postgres not running:**
```bash
docker run -d --name test-postgres \
  -e POSTGRES_USER=monitor \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=monitoring_db \
  -p 5432:5432 \
  postgres:16
```

**Connection refused (Go tests):**
- Ensure PostgreSQL is running on 127.0.0.1:5432
- Check environment variables are set correctly
- Verify database `monitoring_db` exists

**Missing dependencies:**
```bash
# Python
pip install -r devices/backend/requirements-test.txt

# Node
npm install  # in frontend directories
```

**Go: "relation does not exist":**
- This shouldn't happen as tests run AutoMigrate automatically
- If it does, ensure `database.SetupTestDB(t)` is called in your test

### Integration Tests Fail

**Docker issues:**
```bash
# Clean up containers
docker compose -f .github/docker-compose.test.yml down -v

# Remove old images
docker system prune -a
```

**Services not healthy:**
```bash
# Check logs
docker compose -f .github/docker-compose.test.yml logs devices-backend
docker compose -f .github/docker-compose.test.yml logs mentor-backend
```

### Smoke Test Fails

**Services not running:**
```bash
# Make sure services are started
./scripts/start.sh

# Check they're accessible
curl http://localhost:8081/health
curl http://localhost:8080/health
```

**Alert not forwarding:**
- Check `MENTOR_API_URL` in devices backend config
- Verify mentor backend is reachable from devices backend
- Check logs for forwarding errors

## Performance Benchmarks

Expected test execution times:

| Test Type | Duration | Notes |
|-----------|----------|-------|
| Unit tests (Python) | ~1-2s | With DB setup |
| Unit tests (Go) | ~0.5s | Fast compiled tests |
| Unit tests (Frontend) | ~1s each | Vitest is quick |
| Smoke test | ~10s | With running services |
| Integration test | ~30-60s | Includes docker compose up |
| Battle tests (quick) | ~5-10min | Small load (100 devices) |
| Battle tests (full) | ~30-120min | Full load (1000+ devices) |
| Full CI pipeline | ~3-5min | Includes builds |

### Battle Test Performance Targets

| Metric | Target | Acceptable | Current |
|--------|--------|------------|---------|
| Device registration | <100ms p95 | <200ms p95 | Measured by benchmark |
| Telemetry ingestion | >1000 msg/sec | >500 msg/sec | Measured by stress test |
| Alert forwarding | <500ms p95 | <1s p95 | Measured by benchmark |
| Database queries | <50ms p95 | <100ms p95 | Measured by benchmark |
| Screenshot upload | <2s for 1MB | <5s for 1MB | Measured by benchmark |
| API response time | <200ms p95 | <500ms p95 | Measured by benchmark |

These targets are validated by the battle test suite and reported in test outputs.

## Test Coverage

### Overview

The project has comprehensive test coverage tracking for all components (backend and frontend). Coverage is automatically collected in the CI/CD pipeline and uploaded to [Codecov](https://codecov.io/gh/mj-nehme/raqeem).

**Coverage Target**: 90% for all components as per MVP requirements

### Running Coverage Locally

#### Python Backend (Devices)
```bash
cd devices/backend/src
pytest --cov=app --cov-report=html --cov-report=term
# Open htmlcov/index.html to view detailed report
```

#### Go Backend (Mentor)
```bash
cd mentor/backend/src
go test ./... -race -coverprofile=coverage.out
go tool cover -html=coverage.out  # Opens in browser
```

#### Devices Frontend (React)
```bash
cd devices/frontend
npm run test:coverage -- --run
# Open coverage/index.html to view detailed report
```

#### Mentor Frontend (React)
```bash
cd mentor/frontend
npm run test:coverage -- --run
# Open coverage/index.html to view detailed report
```

### Coverage Files Generated

- **Python**: `devices/backend/src/coverage.xml` (XML format for Codecov)
- **Go**: `mentor/backend/src/coverage.out` (Go coverage format)
- **Devices Frontend**: `devices/frontend/coverage/lcov.info` (LCOV format)
- **Mentor Frontend**: `mentor/frontend/coverage/lcov.info` (LCOV format)

### Viewing Combined Coverage

1. Visit the [Codecov Dashboard](https://codecov.io/gh/mj-nehme/raqeem)
2. View overall project coverage and trends
3. Select individual components using the "Flags" filter:
   - `devices-backend` - Python backend coverage
   - `mentor-backend` - Go backend coverage
   - `devices-frontend` - Devices React frontend coverage
   - `mentor-frontend` - Mentor React frontend coverage

### Coverage in CI/CD

The GitHub Actions workflow (`.github/workflows/ci.yml`) automatically:
1. Runs tests for all four components in parallel
2. Generates coverage reports for each component
3. Uploads coverage to Codecov with component-specific flags
4. Comments on PRs with coverage changes

**Note**: Coverage is generated even when tests fail, ensuring comprehensive reporting.

### Coverage Configuration

Coverage is configured in the following files:
- **Python**: `devices/backend/src/pytest.ini` and `devices/backend/src/pyproject.toml`
- **Go**: Command-line flags in CI workflow
- **Frontends**: `vite.config.js` in each frontend directory
- **Codecov**: `codecov.yml` at repository root

### Improving Coverage

To improve coverage for a specific component:

1. **Identify uncovered code**: Check Codecov dashboard or local HTML reports
2. **Add tests**: Write tests for uncovered functions/lines
3. **Run coverage locally**: Verify improvements before committing
4. **Check CI results**: Ensure coverage increases after PR merge

Example workflow:
```bash
# Check current coverage
cd devices/frontend
npm run test:coverage -- --run

# View detailed report
open coverage/index.html  # macOS
xdg-open coverage/index.html  # Linux

# Add tests for uncovered code
vim src/components/MyComponent.test.jsx

# Verify improvement
npm run test:coverage -- --run
```

## CI/CD Pipeline

GitHub Actions workflow (`.github/workflows/ci.yml`):

1. **Setup** - Provision Postgres service
2. **Linting** - Run linters (ruff, mypy, golangci-lint, ESLint)
3. **Build** - Verify Docker images build successfully
4. **Tests** - Run tests with coverage for all components:
   - Python Backend (pytest + coverage)
   - Go Backend (go test + coverage)
   - Devices Frontend (vitest + coverage)
   - Mentor Frontend (vitest + coverage)
5. **Coverage Upload** - Upload all coverage reports to Codecov
6. **Docker Push** - Push images to Docker Hub (on master branch only)

Runs on:
- Every push to `main`/`master`
- Every pull request

## Adding New Tests

### Python Test Template

```python
import pytest
from httpx import AsyncClient, ASGITransport
from app.main import app

@pytest.mark.asyncio
async def test_my_feature():
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
        response = await client.get("/api/v1/my-endpoint")
        assert response.status_code == 200
```

### Go Test Template

```go
func TestMyFeature(t *testing.T) {
    // SetupTestDB begins a transaction - all changes are rolled back automatically
    db, err := database.SetupTestDB(t)
    require.NoError(t, err)
    
    // Set global DB for controllers
    database.DB = db
    
    // Create test data - will be rolled back
    device := models.Device{
        DeviceID: uuid.New(),
        DeviceName: "Test Device",
    }
    db.Create(&device)
    
    // Test your feature
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request, _ = http.NewRequest("GET", "/my-endpoint", nil)
    
    MyController(c)
    
    assert.Equal(t, http.StatusOK, w.Code)
    
    // Transaction is automatically rolled back when test ends
}
```

### Frontend Test Template

```jsx
import { test, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import MyComponent from './MyComponent'

test('renders correctly', () => {
    render(<MyComponent />)
    expect(screen.getByText('Expected Text')).toBeInTheDocument()
})
```

## Best Practices

1. **Test Isolation**: Each test should be independent
   - Go: Tests use transaction-based isolation (automatic rollback)
   - Python: Tests use unique IDs to avoid conflicts
2. **Mock External Services**: Use respx/nock for HTTP mocking
3. **Clear Test Data**: 
   - Go: Data is automatically cleaned up via transaction rollback
   - Python: Use unique IDs (timestamps, UUIDs) for test entities
4. **Meaningful Assertions**: Test behavior, not implementation
5. **Fast Feedback**: Unit tests should run in <5s
6. **Descriptive Names**: Test names should explain what's being tested
7. **PostgreSQL Only**: All tests require PostgreSQL - no SQLite fallback

## Resources

- [pytest documentation](https://docs.pytest.org/)
- [Go testing package](https://pkg.go.dev/testing)
- [Vitest documentation](https://vitest.dev/)
- [act (run GitHub Actions locally)](https://github.com/nektos/act)
