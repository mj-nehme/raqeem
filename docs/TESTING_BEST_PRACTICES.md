# Testing Best Practices

## Overview

This document outlines the testing approach and best practices for the Raqeem IoT monitoring platform.

## Current Test Coverage Status

### Backend - Go (Mentor)
- **Controllers**: 69.9% coverage
- **Database**: 40.3% coverage (test utilities)
- **S3 Client**: 11.8% coverage
- **Overall**: 52.5% coverage

**Key Coverage Areas:**
- ✅ Device registration and management endpoints
- ✅ Metrics, processes, activities, alerts endpoints
- ✅ Command management (GetDeviceCommands, UpdateCommandStatus)
- ✅ Comprehensive edge case testing
- ❌ Main.go (0%) - Entry point, hard to unit test
- ❌ database.Connect() (0%) - Integration point

### Backend - Python (Devices)
- **Without Database**: 72% coverage
- **With Full Test Suite**: ~20-30% (many tests need PostgreSQL)

**Key Coverage Areas:**
- ✅ Models (100% coverage)
- ✅ Schemas (100% coverage)
- ✅ Main app initialization (100% coverage)
- ✅ API routes structure (100% coverage)
- ✅ CORS configuration (100% coverage)
- ⚠️ Endpoints (20-77%) - Many require database
- ⚠️ Database session management (55%)

### Frontend - React (Mentor)
- **Status**: Tests run but 8/14 failing due to async/timeout issues
- **Coverage**: Not generated due to test failures

**Key Test Areas:**
- ✅ Basic component rendering
- ✅ Material-UI integration
- ⚠️ Device list display (timeout issues)
- ⚠️ User interactions (waitFor timeouts)
- ⚠️ API mocking needs refinement

### Frontend - React (Devices)
- **Status**: 26/38 tests passing
- **Coverage**: Partial

**Key Test Areas:**
- ✅ Component rendering
- ✅ Form interactions
- ⚠️ Async API calls (12 tests failing)
- ⚠️ State management after API responses

## Testing Strategy

### Unit Tests
**Purpose**: Test individual functions and components in isolation

**Go Backend:**
```bash
cd mentor/backend/src
go test ./... -v -coverprofile=coverage.out
go tool cover -func=coverage.out
```

**Python Backend:**
```bash
cd devices/backend/src
pytest tests/schemas/ tests/models/ -v --cov=app
```

**Frontend:**
```bash
cd mentor/frontend  # or devices/frontend
npm run test:coverage -- --run
```

### Integration Tests
**Purpose**: Test interactions between components with real or mocked dependencies

**With Database (CI):**
```bash
# Python with PostgreSQL service
pytest -v --cov=app --cov-report=xml

# Go with PostgreSQL service
go test ./... -v -race -coverprofile=coverage.out
```

### Test Organization

#### Go Backend Structure
```
mentor/backend/src/
├── controllers/
│   ├── device.go
│   ├── device_test.go          # Unit tests
│   ├── integration_test.go     # Integration tests
│   ├── comprehensive_test.go   # Comprehensive scenarios
│   ├── coverage_test.go        # Coverage improvement tests
│   └── edge_case_test.go       # Edge cases
├── database/
│   ├── db.go
│   ├── test_db.go              # Test utilities
│   └── database_test.go        # Database tests
└── s3/
    ├── client.go
    └── client_test.go          # S3 client tests
```

#### Python Backend Structure
```
devices/backend/src/tests/
├── api/
│   ├── test_alerts_forwarding.py
│   ├── test_comprehensive_endpoints.py
│   └── v1/endpoints/
│       ├── test_devices.py
│       ├── test_app_activity.py
│       └── ...
├── models/
│   └── test_devices.py         # Model tests (no DB needed)
├── schemas/
│   └── test_schemas.py         # Schema validation tests
└── conftest.py                 # Test configuration
```

## Best Practices

### 1. Test Isolation
- Each test should be independent
- Use `beforeEach`/`afterEach` (JS) or `setup`/`teardown` (Python) for cleanup
- Don't rely on test execution order

### 2. Database Testing
**Go:**
```go
// Use test_db.go utilities
db := database.SetupTestDB(t)
defer database.CleanupTestDB(t, db)
```

**Python:**
```python
# Use pytest fixtures
@pytest.fixture
async def db_session():
    # Create test database session
    yield session
    # Cleanup
```

### 3. Mocking External Dependencies

**Go:**
```go
// Mock HTTP client
mockClient := &MockHTTPClient{
    DoFunc: func(req *http.Request) (*http.Response, error) {
        return &http.Response{StatusCode: 200}, nil
    },
}
```

**Python:**
```python
# Mock database or external APIs
@pytest.fixture
def mock_db():
    with patch('app.db.session.get_db') as mock:
        yield mock
```

**Frontend:**
```javascript
// Mock fetch API
global.fetch = vi.fn(() =>
    Promise.resolve({
        ok: true,
        json: () => Promise.resolve([])
    })
)
```

### 4. Testing Async Code

**Frontend (Vitest):**
```javascript
test('loads data', async () => {
    render(<Component />)
    
    await waitFor(() => {
        expect(screen.getByText('Data')).toBeInTheDocument()
    }, { timeout: 10000 })
})
```

**Python (pytest-asyncio):**
```python
@pytest.mark.asyncio
async def test_async_endpoint():
    async with AsyncClient(app=app, base_url="http://test") as ac:
        response = await ac.get("/devices")
        assert response.status_code == 200
```

### 5. Coverage Goals

**Target Coverage by Component:**
- Models/Schemas: 100% (easy to test, no external dependencies)
- Controllers/Endpoints: 90% (core business logic)
- Integration Points: 70% (harder to test, often entry points)
- Frontend Components: 80% (user-facing, critical paths)

**Excluded from Coverage:**
- Main entry points (main.go, main.py)
- Configuration files
- Migration scripts
- Test utilities

### 6. CI/CD Integration

**GitHub Actions Workflow:**
```yaml
- name: Run tests with coverage
  run: pytest -v --cov=app --cov-report=xml
  
- name: Upload to Codecov
  uses: codecov/codecov-action@v5
  with:
    file: coverage.xml
    flags: component-name
```

**Codecov Configuration** (`codecov.yml`):
```yaml
coverage:
  target: 70%
  threshold: 5%

flags:
  mentor-backend:
    paths:
      - mentor/backend/src/
  devices-backend:
    paths:
      - devices/backend/src/
```

## Common Issues and Solutions

### Issue: Tests timeout in frontend
**Solution**: Increase timeout in vite.config.js:
```javascript
test: {
  testTimeout: 10000,
}
```

### Issue: PostgreSQL connection errors in Python tests
**Solution**: Run tests with database service or skip DB tests:
```bash
pytest tests/schemas/ tests/models/  # Skip DB tests
```

### Issue: Go main.go has 0% coverage
**Solution**: This is expected. Main functions are entry points and are tested through integration tests.

### Issue: Frontend tests fail with MUI import errors
**Solution**: Ensure compatible versions:
```json
{
  "@mui/material": "^5.14.0",
  "@mui/icons-material": "^5.14.0"  // Not 7.x
}
```

## Running Tests Locally

### Prerequisites
```bash
# Go
go version  # 1.25+

# Python
python3 --version  # 3.11+
pip install -r requirements.txt -r requirements-test.txt

# Node
node --version  # 20+
npm install
```

### Quick Test Commands

**All Go tests:**
```bash
cd mentor/backend/src && go test ./... -v
```

**Python tests (without DB):**
```bash
cd devices/backend/src
pytest tests/schemas/ tests/models/ tests/test_main.py -v
```

**Frontend tests:**
```bash
cd mentor/frontend && npm test
```

**With coverage:**
```bash
# Go
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Python
pytest --cov=app --cov-report=html
open htmlcov/index.html

# Frontend
npm run test:coverage
open coverage/index.html
```

## Future Improvements

1. **Increase integration test coverage** for main.go and database.Connect()
2. **Fix frontend async test timeouts** by improving mock consistency
3. **Add E2E tests** using Playwright or Cypress
4. **Implement visual regression testing** for UI components
5. **Add performance testing** for high-throughput endpoints
6. **Set up mutation testing** to verify test quality

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Pytest Documentation](https://docs.pytest.org/)
- [Vitest Documentation](https://vitest.dev/)
- [Testing Library](https://testing-library.com/)
- [Codecov Documentation](https://docs.codecov.com/)
