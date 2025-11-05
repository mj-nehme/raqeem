# Coverage Reporting Setup

This document explains the automated coverage reporting infrastructure for the Raqeem IoT Platform.

## Overview

Automated coverage reporting is **fully configured and operational** in the CI/CD pipeline. Coverage reports are collected for all four components and uploaded to Codecov automatically.

## Current Configuration

### 1. CI/CD Workflow

**File**: `.github/workflows/ci.yml`

The CI workflow includes comprehensive test jobs for:
- ✅ **Python Backend** (Devices) - pytest with coverage
- ✅ **Go Backend** (Mentor) - go test with coverage
- ✅ **React Frontend** (Devices) - vitest with coverage
- ✅ **React Frontend** (Mentor) - vitest with coverage

Each test job:
1. Sets up the required runtime (Python 3.11 / Go 1.25 / Node 20)
2. Installs dependencies with caching for faster builds
3. Runs tests with coverage collection
4. Uploads coverage reports to Codecov with appropriate flags
5. Uses `continue-on-error: true` to allow workflow to complete even if individual tests fail

### 2. Coverage Collection Script

**File**: `scripts/collect-comprehensive-coverage.sh`

A local script for collecting coverage across all components. Useful for:
- Pre-commit coverage validation
- Local development testing
- Coverage baseline verification

**Usage**:
```bash
./scripts/collect-comprehensive-coverage.sh
```

**Note**: Local execution requires all services (PostgreSQL, etc.) to be running. Tests will skip if dependencies are not available.

### 3. Test Configurations

#### Python Backend (Devices)
- **Config**: `devices/backend/src/pytest.ini`
- **Test command**: `pytest --cov=app --cov-report=xml`
- **Coverage output**: `devices/backend/src/coverage.xml`
- **Requirements**: `devices/backend/requirements-test.txt`

#### Go Backend (Mentor)
- **Test command**: `go test ./... -race -coverprofile=coverage.out`
- **Coverage output**: `mentor/backend/src/coverage.out`

#### React Frontends (Both)
- **Config**: `vite.config.js` in each frontend directory
- **Test command**: `npm run test:coverage -- --run`
- **Coverage output**: `*/frontend/coverage/lcov.info`
- **Coverage provider**: v8 (via vitest)

### 4. Codecov Configuration

**File**: `codecov.yml`

Configured with:
- **Target coverage**: 90% minimum (per MVP requirements)
- **Threshold**: 5% (allow up to 5% decrease)
- **Flags**: Separate tracking for each component
  - `mentor-backend`
  - `devices-backend`
  - `mentor-frontend`
  - `devices-frontend`
- **Comments**: Enabled on PRs with coverage diff and flag details
- **Carryforward**: Enabled to handle missing coverage uploads

### Key Features

**Coverage on Test Failure**: Both frontend applications are configured with `reportOnFailure: true` in their Vitest configurations, ensuring coverage reports are generated even when some tests fail. This provides comprehensive coverage data regardless of test status.

## Codecov Integration

### Viewing Coverage Reports

1. **Repository Badge**: README.md includes a Codecov badge showing overall coverage
2. **Codecov Dashboard**: https://codecov.io/gh/mj-nehme/raqeem
3. **PR Comments**: Automatic comments on pull requests showing coverage changes

### Coverage Flags

Each component uploads with a specific flag, allowing you to:
- Track coverage per component
- See trends for individual services
- Identify which component needs improved test coverage

### Setup Requirements

The Codecov GitHub Action requires a `CODECOV_TOKEN` secret to be configured in the repository settings. This token authenticates the coverage uploads.

**To configure**:
1. Go to https://codecov.io/gh/mj-nehme/raqeem/settings
2. Copy the upload token
3. Add as `CODECOV_TOKEN` in GitHub repository secrets

## Running Tests Locally

### Prerequisites

Each component has specific requirements:

#### Python Backend
```bash
cd devices/backend
pip install -r requirements.txt -r requirements-test.txt
cd src
pytest --cov=app --cov-report=html
```

#### Go Backend
```bash
cd mentor/backend/src
go test ./... -v -race -coverprofile=coverage.out
go tool cover -html=coverage.out  # View HTML report
```

#### React Frontends
```bash
cd devices/frontend  # or mentor/frontend
npm ci
npm run test:coverage
# Coverage report in ./coverage/
```

### Database Requirements

Tests that interact with the database require PostgreSQL:
- The CI workflow provides PostgreSQL via Docker services
- Locally, you can use Docker Compose or the platform's start.sh script
- Tests gracefully skip if database is unavailable (for isolated unit tests)

## CI/CD Pipeline Flow

```
┌─────────────────────────────────────┐
│  Push / Pull Request                │
└──────────────┬──────────────────────┘
               │
               v
┌─────────────────────────────────────┐
│  1. Lint & Type Check (parallel)    │
│     - Python (ruff, mypy)           │
│     - Go (golangci-lint)            │
│     - Frontend (ESLint)             │
└──────────────┬──────────────────────┘
               │
               v
┌─────────────────────────────────────┐
│  2. Build Docker Images             │
│     (verify build succeeds)         │
└──────────────┬──────────────────────┘
               │
               v
┌─────────────────────────────────────┐
│  3. Test with Coverage (parallel)   │
│     - Python Backend + PostgreSQL   │
│     - Go Backend + PostgreSQL       │
│     - Devices Frontend              │
│     - Mentor Frontend               │
└──────────────┬──────────────────────┘
               │
               v
┌─────────────────────────────────────┐
│  4. Upload Coverage to Codecov      │
│     (each component with flags)     │
└──────────────┬──────────────────────┘
               │
               v (if master branch)
┌─────────────────────────────────────┐
│  5. Build & Push Docker Images      │
│     to Docker Hub                   │
└─────────────────────────────────────┘
```

## Coverage Artifacts

Coverage reports are excluded from version control (`.gitignore`):
- `coverage.xml` - Python XML format
- `coverage.out` - Go coverage data
- `.coverage` - Python coverage database
- `coverage/` - HTML reports directory
- `*.lcov` - LCOV format (JavaScript)

## Troubleshooting

### Tests Pass Locally but Fail in CI

Check:
1. **Environment variables**: CI uses test-specific env vars
2. **Database connection**: CI provides PostgreSQL on `127.0.0.1:5432`
3. **Dependencies**: Ensure all test dependencies are in requirements-test.txt or package.json

### Coverage Not Uploading

Check:
1. **CODECOV_TOKEN**: Verify the secret is configured in GitHub
2. **File paths**: Ensure coverage files are generated at expected paths
3. **Workflow logs**: Check for upload errors in the Actions tab

### Low Coverage Warnings

The Codecov configuration sets a 90% target (per MVP requirements). To improve:
1. Add more test cases for untested functions
2. Check the Codecov dashboard for specific uncovered lines
3. Focus on critical paths and error handling
4. Use local coverage reports to identify gaps quickly

## Best Practices

1. **Run tests before committing**: Use `./scripts/collect-comprehensive-coverage.sh`
2. **Check coverage trends**: Monitor the Codecov dashboard after PRs merge
3. **Write meaningful tests**: Focus on behavior, not just line coverage
4. **Test error cases**: Don't just test the happy path
5. **Keep tests fast**: Use mocking for external services where appropriate

## Success Metrics

Current setup achieves:
- ✅ **Automated coverage** collection on every push/PR
- ✅ **Multi-language support**: Python, Go, JavaScript
- ✅ **Parallel execution**: Tests run concurrently for faster feedback
- ✅ **Comprehensive reporting**: Per-component and aggregate coverage
- ✅ **PR integration**: Coverage changes visible in PR discussions
- ✅ **Badge visibility**: README shows current coverage status

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Codecov GitHub Action](https://github.com/codecov/codecov-action)
- [pytest-cov Documentation](https://pytest-cov.readthedocs.io/)
- [Go Coverage Testing](https://go.dev/blog/cover)
- [Vitest Coverage](https://vitest.dev/guide/coverage.html)
