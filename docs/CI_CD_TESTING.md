# CI/CD Automated Testing Guide

## Overview

Raqeem has a comprehensive CI/CD pipeline that automatically runs tests, linting, and builds on every push and pull request. This document describes the automated testing infrastructure and how to work with it effectively.

## Pipeline Architecture

### Current Setup

The CI/CD pipeline is defined in `.github/workflows/ci.yml` and includes:

1. **Check Services** - Verify PostgreSQL availability
2. **Linting** - Code style and quality checks
3. **Type Checking** - Static type validation
4. **Unit Tests** - Component-level tests
5. **Coverage Reporting** - Upload to Codecov
6. **Build** - Docker image builds
7. **Publish** - Push images to GitHub Container Registry (main branch only)

### Pipeline Flow

```
┌─────────────────────┐
│   Push/PR Event     │
└─────────┬───────────┘
          │
          ├──────────────────────────────────────┐
          │                                      │
          v                                      v
┌─────────────────────┐              ┌─────────────────────┐
│  Check Services     │              │  Linting & Checks   │
│  • PostgreSQL       │              │  • Python (ruff)    │
│    Health Check     │              │  • Python (mypy)    │
└─────────┬───────────┘              │  • Go (golangci)    │
          │                          │  • Frontend (ESLint)│
          │                          └─────────┬───────────┘
          │                                    │
          └────────────┬───────────────────────┘
                       │
                       v
          ┌────────────────────────────┐
          │   Build Docker Images      │
          │   • Devices Backend        │
          │   • Mentor Backend         │
          │   (build only, no push)    │
          └────────────┬───────────────┘
                       │
          ┌────────────┴───────────────┐
          │                            │
          v                            v
┌─────────────────────┐    ┌─────────────────────┐
│  Backend Tests      │    │  Frontend Tests     │
│  • Devices (Python) │    │  • Mentor (React)   │
│  • Mentor (Go)      │    │  • Devices (React)  │
│  • With Coverage    │    │  • With Coverage    │
└─────────┬───────────┘    └─────────┬───────────┘
          │                           │
          └────────────┬──────────────┘
                       │
                       v
          ┌────────────────────────────┐
          │   Upload Coverage          │
          │   • Codecov Integration    │
          │   • Per-component flags    │
          └────────────┬───────────────┘
                       │
                       v (main branch only)
          ┌────────────────────────────┐
          │   Publish to GHCR          │
          │   • Tagged images          │
          │   • Latest tags            │
          └────────────────────────────┘
```

## Job Breakdown

### 1. Check Services Job

**Purpose**: Ensure required services (PostgreSQL) are healthy before running tests

**What it does**:
- Starts PostgreSQL container with test credentials
- Waits for PostgreSQL to be ready (health checks)
- Verifies database connection
- Initializes test database

**Runs**: Always (prerequisite for test jobs)

### 2. Linting Jobs

#### Python Linting (ruff)
- **Tool**: `ruff`
- **Target**: `devices/backend/src`
- **Checks**: Code style, imports, complexity, best practices
- **Configuration**: `pyproject.toml`

#### Python Type Checking (mypy)
- **Tool**: `mypy`
- **Target**: `devices/backend/src`
- **Checks**: Type annotations, type safety
- **Configuration**: `pyproject.toml` (relaxed mode)

#### Go Linting (golangci-lint)
- **Tool**: `golangci-lint`
- **Target**: `mentor/backend/src`
- **Checks**: Go best practices, code quality, potential bugs
- **Timeout**: 5 minutes

#### Frontend Linting (ESLint)
- **Tool**: `ESLint`
- **Target**: `devices/frontend`
- **Checks**: JavaScript/React best practices, code style

### 3. Build Artifacts Job

**Purpose**: Verify Docker images can be built successfully

**What it does**:
- Sets up Docker Buildx for multi-platform builds
- Builds devices backend image (no push)
- Builds mentor backend image (no push)
- Validates Dockerfile syntax and build process

**Runs**: After all linting jobs pass

### 4. Test Jobs

#### Test Devices Backend (Python)

```yaml
Language: Python 3.11
Framework: FastAPI
Test Tool: pytest
Coverage: pytest-cov
Database: PostgreSQL 16
```

**What it tests**:
- API endpoints
- Service layer
- Database operations (with real PostgreSQL)
- Schema validation
- Error handling

**Environment**:
- PostgreSQL connection
- MinIO configuration
- Mentor API URL
- Test credentials

**Coverage**:
- Reports to Codecov with `devices-backend` flag
- Target: 90% coverage
- XML report: `coverage.xml`

#### Test Mentor Backend (Go)

```yaml
Language: Go 1.25
Framework: Gin
Test Tool: go test
Coverage: -coverprofile
Database: PostgreSQL 16
```

**What it tests**:
- Controllers
- Models
- Database operations (with real PostgreSQL)
- API endpoints
- Transaction isolation

**Features**:
- Race condition detection (`-race` flag)
- Transaction-based test isolation
- Automatic rollback after each test

**Coverage**:
- Reports to Codecov with `mentor-backend` flag
- Target: 90% coverage
- Coverage file: `coverage.out`

#### Test Mentor Frontend (React)

```yaml
Language: Node.js 20
Framework: React + Vite
Test Tool: Vitest
Coverage: c8
```

**What it tests**:
- Component rendering
- User interactions
- State management
- API integration (mocked)

**Coverage**:
- Reports to Codecov with `mentor-frontend` flag
- Target: 90% coverage
- Format: lcov

#### Test Devices Frontend (React)

```yaml
Language: Node.js 20
Framework: React + Vite
Test Tool: Vitest
Coverage: c8
```

**What it tests**:
- Device simulator components
- Form handling
- API calls (mocked)
- UI interactions

**Coverage**:
- Reports to Codecov with `devices-frontend` flag
- Target: 90% coverage
- Format: lcov

### 5. Publish Images Job

**Purpose**: Push Docker images to GitHub Container Registry

**When**: Only on pushes to `main` branch (not PRs)

**Requirements**:
- All previous jobs must pass
- Automatic authentication with `GITHUB_TOKEN`

**Images**:
- `ghcr.io/mj-nehme/raqeem/devices-backend`
- `ghcr.io/mj-nehme/raqeem/mentor-backend`

**Tags**:
- `latest` - Most recent main branch build
- `main` - Main branch tag
- `<sha>` - Commit SHA tag

## Coverage Reporting

### Codecov Integration

**Configuration**: `codecov.yml`

**Flags**:
- `devices-backend` - Python backend coverage
- `mentor-backend` - Go backend coverage
- `mentor-frontend` - React frontend coverage
- `devices-frontend` - React frontend coverage

**Targets**:
- Project coverage: 90%
- Patch coverage: 90%
- Threshold: 5%

**Features**:
- Comment on PRs with coverage changes
- Status checks (can block merge if coverage drops)
- Coverage badges in README

### Coverage Reports

Each test job uploads coverage separately:

```yaml
- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v5
  with:
    files: <coverage-file>
    flags: <component-flag>
    name: <component-name>
    token: ${{ secrets.CODECOV_TOKEN }}
```

## Test Database Setup

### PostgreSQL Service

All test jobs that need database access use a PostgreSQL service container:

```yaml
services:
  postgres:
    image: docker.io/library/postgres:16
    env:
      POSTGRES_USER: monitor
      POSTGRES_PASSWORD: password
      POSTGRES_DB: monitoring_db
    ports:
      - 5432:5432
    options: >-
      --health-cmd "pg_isready -U monitor"
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5
```

**Credentials** (test only):
- User: `monitor`
- Password: `password`
- Database: `monitoring_db`
- Host: `127.0.0.1`
- Port: `5432`

### Test Isolation

- **Go tests**: Use transaction-based isolation
  - Each test runs in a transaction
  - Automatic rollback after test completes
  - No test data persists between tests

- **Python tests**: Use test database fixtures
  - Async SQLAlchemy sessions
  - NullPool for connection management
  - Cleanup after test runs

## Caching Strategy

### Python Dependencies

```yaml
- name: Cache pip packages
  uses: actions/cache@v4
  with:
    path: ~/.cache/pip
    key: ${{ runner.os }}-pip-<component>-${{ hashFiles('requirements*.txt') }}
```

### Go Modules

```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: "1.25.x"
    cache: true
    cache-dependency-path: go.sum
```

### Node Modules

```yaml
- name: Cache npm dependencies
  uses: actions/cache@v4
  with:
    path: ~/.npm
    key: ${{ runner.os }}-npm-<component>-${{ hashFiles('package-lock.json') }}
```

**Benefits**:
- Faster pipeline execution
- Reduced bandwidth usage
- Consistent dependency versions

## Running Tests Locally

### Quick Test Run

```bash
# Python backend
cd devices/backend/src
pytest tests/ -v

# Go backend
cd mentor/backend/src
go test ./... -v

# Frontend (Mentor)
cd mentor/frontend
npm test

# Frontend (Devices)
cd devices/frontend
npm test
```

### With Coverage

```bash
# Python
pytest --cov=app --cov-report=html

# Go
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Frontend
npm run test:coverage
```

### With Database

```bash
# Start PostgreSQL
docker run --name test-postgres \
  -e POSTGRES_USER=monitor \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=monitoring_db \
  -p 5432:5432 \
  -d postgres:16

# Run tests (will use localhost:5432)
pytest tests/ -v
go test ./... -v

# Cleanup
docker stop test-postgres
docker rm test-postgres
```

## Troubleshooting CI Failures

### Test Failures

**Symptom**: Tests fail in CI but pass locally

**Common causes**:
1. **Missing database**: Ensure PostgreSQL service is available
2. **Timing issues**: Tests may be flaky with timeouts
3. **Environment differences**: Check environment variables

**Solution**:
```bash
# Run tests with same setup as CI
DATABASE_URL="postgresql+asyncpg://monitor:password@127.0.0.1:5432/monitoring_db" \
  pytest tests/ -v
```

### Linting Failures

**Symptom**: Linting fails in CI but passes locally

**Common causes**:
1. **Different tool versions**: CI uses pinned versions
2. **Cached results**: Local cache may be stale

**Solution**:
```bash
# Python - use exact CI version
pip install ruff==0.9.0
ruff check .

# Go - use CI version
golangci-lint run --timeout=5m

# Frontend
npm run lint
```

### Coverage Failures

**Symptom**: Coverage below threshold

**Common causes**:
1. **New code without tests**: Added code not covered
2. **Deleted tests**: Removed tests without removing code

**Solution**:
```bash
# Identify uncovered lines
pytest --cov=app --cov-report=term-missing

# View detailed coverage
go tool cover -html=coverage.out
```

### Build Failures

**Symptom**: Docker build fails

**Common causes**:
1. **Missing dependencies**: Requirements file not updated
2. **Syntax errors**: Invalid Dockerfile

**Solution**:
```bash
# Test build locally
docker build -t test-image -f devices/backend/Dockerfile devices/backend

# Check build logs
docker build --progress=plain -t test-image .
```

## Best Practices

### Writing CI-Friendly Tests

1. **Make tests independent**: Don't rely on execution order
2. **Clean up resources**: Close connections, delete test data
3. **Use fixtures**: Reuse common setup/teardown logic
4. **Mock external services**: Don't call real external APIs
5. **Set timeouts**: Prevent tests from hanging

### Maintaining Fast CI

1. **Parallelize tests**: Run independent tests concurrently
2. **Use caching**: Cache dependencies between runs
3. **Skip slow tests**: Mark slow tests, run separately
4. **Optimize Docker builds**: Use layer caching, multi-stage builds

### Handling Flaky Tests

1. **Identify flaky tests**: Monitor test failures
2. **Add retries**: Use `pytest-rerunfailures` or similar
3. **Increase timeouts**: For timing-sensitive tests
4. **Fix root cause**: Don't just disable flaky tests

## Monitoring and Alerts

### GitHub Actions Status

- **Badge**: Shows current status
- **Notifications**: Email/Slack on failure
- **Required checks**: Block PRs if tests fail

### Codecov Status

- **Coverage badge**: Shows current coverage
- **PR comments**: Coverage changes highlighted
- **Status checks**: Fail if coverage drops below threshold

## Future Enhancements

### Planned Improvements

1. **Parallel test execution**: Run tests faster
2. **Matrix builds**: Test on multiple Python/Go versions
3. **Performance tests**: Add benchmark tests to CI
4. **E2E tests**: Run full system tests in CI
5. **Security scanning**: Add dependency vulnerability checks

## See Also

- [Testing Guide](TESTING.md) - Comprehensive testing documentation
- [Error Handling](ERROR_HANDLING.md) - Error handling patterns
- [Coverage Report](TEST_COVERAGE_REPORT.md) - Current coverage status
- [Local CI](LOCAL_CI.md) - Running GitHub Actions locally

## Version History

- **2025-11-18**: Initial CI/CD documentation for v0.2.0
