# 🛠️ Development Guide

## Overview

This guide covers local development workflows, coding standards, testing practices, and contribution guidelines for the Raqeem IoT monitoring platform.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Structure](#code-structure)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Git Workflow](#git-workflow)
- [Release Process](#release-process)
- [Troubleshooting](#troubleshooting)

## Getting Started

### Prerequisites

**Required**:
- Docker Desktop with Kubernetes enabled
- `kubectl` - Kubernetes CLI
- `helm` 3.x - Kubernetes package manager
- Node.js 18+ and npm
- Git

**Optional (for backend development)**:
- Python 3.11+ with `pip`
- Go 1.25+ 
- `pytest` - Python testing framework
- `pre-commit` - Git hooks (recommended)

### Initial Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/mj-nehme/raqeem.git
   cd raqeem
   ```

2. **Start the platform**:
   ```bash
   ./start.sh
   ```

3. **Verify all services are running**:
   ```bash
   ./scripts/discover.sh list
   kubectl get pods -n default
   ```

See [First Time Setup Guide](FIRST_TIME_SETUP.md) for detailed instructions.

## Development Workflow

### Option 1: Full Stack in Kubernetes (Recommended)

Best for:
- Testing the complete system
- Working on frontend
- Integration testing
- Production-like environment

```bash
# Start everything
./start.sh

# Make changes to frontend code
# Changes are auto-reloaded by Vite dev server

# View logs
kubectl logs -f deployment/devices-backend -n default
kubectl logs -f deployment/mentor-backend -n default

# Restart backend after code changes
kubectl rollout restart deployment/devices-backend -n default
kubectl rollout restart deployment/mentor-backend -n default
```

### Option 2: Hybrid Development (Backends Local)

Best for:
- Backend development with fast iteration
- Debugging backend code
- Database schema changes

```bash
# 1. Deploy only infrastructure
helm upgrade --install postgres ./charts/postgres -n default
helm upgrade --install minio ./charts/minio -n default

# 2. Port-forward infrastructure
kubectl port-forward svc/postgres-service 5432:5432 -n default &
kubectl port-forward svc/minio-service 9000:9000 9001:9001 -n default &

# 3. Set up environment variables
export DATABASE_URL="postgresql://monitor:password@localhost:5432/monitoring_db"
export MINIO_ENDPOINT="localhost:9000"
export MINIO_ACCESS_KEY="minioadmin"
export MINIO_SECRET_KEY="minioadmin"

# 4. Run backends locally
# Terminal 1: Devices Backend
cd devices/backend/src
pip install -r ../requirements.txt
uvicorn app.main:app --reload --port 8081

# Terminal 2: Mentor Backend
cd mentor/backend/src
go run main.go

# 5. Run frontends locally
# Terminal 3: Mentor Frontend
cd mentor/frontend
npm install
npm run dev

# Terminal 4: Devices Frontend
cd devices/frontend
npm install
npm run dev
```

### Option 3: Frontend-Only Development

Best for:
- UI/UX work
- Frontend feature development

```bash
# 1. Deploy backends and infrastructure
./start.sh

# 2. Stop frontend dev servers (Ctrl+C on start.sh output)

# 3. Run frontends manually with custom ports
cd mentor/frontend
VITE_API_URL=http://localhost:30081 npm run dev -- --port 5173

cd devices/frontend
VITE_DEVICES_API_URL=http://localhost:30080 npm run dev -- --port 5174
```

## Code Structure

### Repository Layout

```
raqeem/
├── devices/
│   ├── backend/           # FastAPI backend
│   │   ├── src/
│   │   │   ├── app/
│   │   │   │   ├── api/          # API routes
│   │   │   │   ├── models/       # SQLAlchemy models
│   │   │   │   ├── db/           # Database connection
│   │   │   │   └── main.py       # Application entry point
│   │   │   └── tests/            # Unit tests
│   │   ├── requirements.txt      # Python dependencies
│   │   └── Dockerfile
│   └── frontend/          # React frontend (simulator)
│       ├── src/
│       │   ├── components/       # React components
│       │   ├── App.jsx           # Main app component
│       │   └── main.jsx          # Entry point
│       ├── package.json
│       └── vite.config.js
├── mentor/
│   ├── backend/           # Go backend
│   │   ├── src/
│   │   │   ├── controllers/      # HTTP handlers
│   │   │   ├── models/           # GORM models
│   │   │   ├── database/         # DB connection
│   │   │   └── main.go           # Application entry point
│   │   ├── go.mod
│   │   └── Dockerfile
│   └── frontend/          # React frontend (dashboard)
│       ├── src/
│       │   ├── components/       # React components
│       │   ├── App.jsx
│       │   └── main.jsx
│       ├── package.json
│       └── vite.config.js
├── charts/                # Helm charts
│   ├── postgres/
│   ├── minio/
│   ├── devices-backend/
│   └── mentor-backend/
├── docs/                  # Documentation
├── scripts/               # Helper scripts
├── tests/                 # Integration tests
│   └── integration/
└── README.md
```

### Component Responsibilities

#### Devices Backend (`devices/backend/src`)
- `app/main.py` - FastAPI application setup, CORS, routes
- `app/api/routes.py` - API route definitions
- `app/api/v1/endpoints/` - Endpoint implementations
- `app/models/` - SQLAlchemy database models
- `app/db/` - Database connection and session management
- `tests/` - pytest test files

#### Mentor Backend (`mentor/backend/src`)
- `main.go` - Gin application setup, routes, CORS
- `controllers/` - HTTP request handlers
- `models/` - GORM database models
- `database/` - Database connection setup

#### Frontends
- `src/components/` - Reusable React components
- `src/App.jsx` - Main application component
- `src/main.jsx` - React DOM rendering entry point

## Coding Standards

### Python (Devices Backend)

**Style Guide**: Follow PEP 8

**Key Conventions**:
- Use 4 spaces for indentation
- Max line length: 88 characters (Black formatter)
- Use type hints for function signatures
- Docstrings for all public functions/classes

**Example**:
```python
from typing import Optional
from pydantic import BaseModel

class DeviceMetric(BaseModel):
    """Device metrics data model."""
    device_id: str
    cpu_usage: Optional[float] = None
    memory_used: Optional[int] = None

async def get_device_metrics(device_id: str) -> list[DeviceMetric]:
    """
    Retrieve metrics for a specific device.
    
    Args:
        device_id: Unique device identifier
        
    Returns:
        List of device metrics
    """
    # Implementation
    pass
```

**Tools**:
```bash
# Format code
black devices/backend/src

# Lint code
flake8 devices/backend/src

# Type checking
mypy devices/backend/src
```

### Go (Mentor Backend)

**Style Guide**: Follow Effective Go

**Key Conventions**:
- Use `gofmt` for formatting (built-in)
- Use tabs for indentation
- Exported names start with capital letter
- Short variable names in limited scope
- Comment exported functions

**Example**:
```go
// DeviceMetric represents system performance metrics
type DeviceMetric struct {
    DeviceID  string    `json:"device_id"`
    CPUUsage  float64   `json:"cpu_usage"`
    Timestamp time.Time `json:"timestamp"`
}

// GetDeviceMetric retrieves metrics for a device
func GetDeviceMetric(c *gin.Context) {
    deviceID := c.Param("id")
    // Implementation
}
```

**Tools**:
```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Vet code
go vet ./...
```

### JavaScript/React (Frontends)

**Style Guide**: Airbnb JavaScript Style Guide (adapted)

**Key Conventions**:
- Use 2 spaces for indentation
- Prefer functional components with hooks
- Use meaningful component and variable names
- PropTypes or TypeScript for type checking

**Example**:
```jsx
import React, { useState, useEffect } from 'react';

function DeviceMetric({ deviceId }) {
  const [metrics, setMetrics] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        const response = await fetch(`/api/devices/${deviceId}/metrics`);
        const data = await response.json();
        setMetrics(data);
      } catch (error) {
        console.error('Failed to fetch metrics:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchMetrics();
  }, [deviceId]);

  if (loading) return <div>Loading...</div>;

  return (
    <div className="metrics">
      {metrics.map(m => (
        <MetricCard key={m.id} metric={m} />
      ))}
    </div>
  );
}

export default DeviceMetric;
```

**Tools**:
```bash
# Format code
npm run format  # or prettier --write src/

# Lint code
npm run lint    # or eslint src/
```

### Database Migrations

**Python (SQLAlchemy)**:
```bash
# Generate migration
alembic revision --autogenerate -m "Add new column"

# Apply migration
alembic upgrade head

# Rollback
alembic downgrade -1
```

**Go (GORM)**:
GORM uses AutoMigrate. Add new models in `main.go`:
```go
database.DB.AutoMigrate(&models.NewModel{})
```

## Testing Guidelines

### Test Pyramid

Focus on:
1. **Unit Tests** (most) - Fast, isolated, test individual functions
2. **Integration Tests** (moderate) - Test service interactions
3. **E2E Tests** (few) - Test complete user workflows

### Python Testing (pytest)

**Location**: `devices/backend/src/tests/`

**Running Tests**:
```bash
cd devices/backend/src

# Run all tests
pytest

# Run with coverage
pytest --cov=app --cov-report=html

# Run specific test file
pytest tests/api/test_alerts_forwarding.py

# Run with verbose output
pytest -v

# Run specific test
pytest tests/api/test_alerts_forwarding.py::test_post_alerts_is_saved_and_forwarded
```

**Writing Tests**:
```python
import pytest
from httpx import AsyncClient, ASGITransport
from app.main import app

@pytest.mark.asyncio
async def test_register_device():
    """Test device registration endpoint."""
    device_data = {
        "id": "test-device-001",
        "name": "Test Device",
        "type": "laptop"
    }
    
    async with AsyncClient(
        transport=ASGITransport(app=app),
        base_url="http://test"
    ) as client:
        response = await client.post(
            "/api/v1/devices/register",
            json=device_data
        )
        
    assert response.status_code == 200
    data = response.json()
    assert data["id"] == device_data["id"]
```

### Go Testing

**Location**: `mentor/backend/src/controllers/*_test.go`

**Running Tests**:
```bash
cd mentor/backend/src

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./controllers

# Verbose output
go test -v ./...
```

**Writing Tests**:
```go
func TestGetDeviceMetric(t *testing.T) {
    // Setup
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Params = gin.Params{
        {Key: "id", Value: "test-device-001"},
    }

    // Execute
    GetDeviceMetric(c)

    // Assert
    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }
}
```

### Frontend Testing (Vitest)

**Running Tests**:
```bash
cd mentor/frontend  # or devices/frontend

# Run tests
npm test

# Run with UI
npm run test:ui

# Run with coverage
npm run test:coverage
```

**Writing Tests**:
```jsx
import { test, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import DeviceList from './DeviceList';

test('renders device list', () => {
  const devices = [
    { id: '1', name: 'Device 1', is_online: true }
  ];
  
  render(<DeviceList devices={devices} />);
  
  expect(screen.getByText('Device 1')).toBeInTheDocument();
});
```

### Integration Testing

**Location**: `tests/integration/`

**Running Integration Tests**:
```bash
# Full integration test with docker-compose
./tests/integration/run_integration_tests.sh

# Manual smoke test (requires running services)
python3 tests/smoke_test.py
```

### Test Checklist Before Commit

- [ ] All unit tests pass
- [ ] Added tests for new features
- [ ] Code coverage maintained (>70%)
- [ ] Integration tests pass
- [ ] No linting errors

## Git Workflow

### Branch Strategy

**Main Branches**:
- `master` (or `main`) - Production-ready code
- Feature branches - `feature/<feature-name>`
- Bug fixes - `bugfix/<bug-name>`
- Releases - `release/<version>`

### Workflow

1. **Create feature branch**:
   ```bash
   git checkout -b feature/add-device-filtering
   ```

2. **Make changes and commit**:
   ```bash
   git add .
   git commit -m "feat: add device filtering by status"
   ```

3. **Keep branch updated**:
   ```bash
   git fetch origin
   git rebase origin/master
   ```

4. **Push to remote**:
   ```bash
   git push origin feature/add-device-filtering
   ```

5. **Create Pull Request** on GitHub

6. **Code Review** - Address feedback

7. **Merge** - Squash and merge or merge commit

### Commit Message Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types**:
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting)
- `refactor` - Code refactoring
- `test` - Adding/updating tests
- `chore` - Maintenance tasks

**Examples**:
```bash
feat(devices): add device filtering endpoint
fix(mentor): correct alert severity sorting
docs(readme): update installation instructions
test(devices): add tests for metrics endpoint
chore(deps): update FastAPI to 0.104.0
```

### Pull Request Guidelines

**PR Title**: Follow commit message convention

**PR Description** should include:
- Summary of changes
- Related issue number (`Fixes #123`)
- Testing steps
- Screenshots (for UI changes)
- Breaking changes (if any)

**Template**:
```markdown
## Description
Brief description of changes

## Related Issue
Fixes #123

## Changes Made
- Added device filtering
- Updated API documentation
- Added tests

## Testing Steps
1. Start the platform
2. Navigate to devices page
3. Apply filter
4. Verify results

## Screenshots
[If applicable]

## Checklist
- [ ] Tests pass
- [ ] Documentation updated
- [ ] No breaking changes
```

## Release Process

### Semantic Versioning

Format: `MAJOR.MINOR.PATCH` (e.g., `v1.2.3`)

- **MAJOR** - Breaking changes
- **MINOR** - New features (backward compatible)
- **PATCH** - Bug fixes

### Creating a Release

1. **Update version**:
   ```bash
   # Decide version number
   export NEW_VERSION="v1.2.0"
   ```

2. **Run release script**:
   ```bash
   ./scripts/tag-release.sh $NEW_VERSION
   ```

   This automatically:
   - Validates code (compiles, checks syntax)
   - Builds Docker images
   - Tags images with version and commit SHA
   - Pushes to Docker Hub
   - Updates Helm charts
   - Creates git tag
   - Generates release notes

3. **Verify release**:
   ```bash
   git tag -l
   docker images | grep raqeem
   ```

4. **Deploy specific version**:
   ```bash
   echo "IMAGE_TAG=v1.2.0" > .deploy/tag.env
   ./start.sh
   ```

See [Version Management](VERSION_MANAGEMENT.md) and [Release Workflow](RELEASE_WORKFLOW.md) for details.

### Pre-Release Checklist

- [ ] All tests pass (unit, integration, E2E)
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] Version bumped in appropriate files
- [ ] Docker images build successfully
- [ ] Helm charts deploy successfully
- [ ] No security vulnerabilities (run `npm audit`, `pip check`)

## Troubleshooting

### Backend Development Issues

**Import errors in Python**:
```bash
# Ensure you're in the right directory
cd devices/backend/src

# Reinstall dependencies
pip install -r ../requirements.txt
```

**Go module issues**:
```bash
cd mentor/backend/src
go mod tidy
go mod download
```

### Frontend Development Issues

**Module not found**:
```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

**Port conflicts**:
```bash
# Check what's using the port
lsof -i :5173

# Kill the process
kill -9 <PID>

# Or use a different port
npm run dev -- --port 5174
```

### Database Issues

**Connection refused**:
```bash
# Check if postgres is running
kubectl get pods | grep postgres

# Check service
kubectl get svc postgres-service

# Port-forward manually
kubectl port-forward svc/postgres-service 5432:5432 -n default
```

**Schema mismatch**:
```bash
# Python: Run migrations
cd devices/backend/src
alembic upgrade head

# Go: Restart app (AutoMigrate runs on startup)
kubectl rollout restart deployment/mentor-backend
```

### Kubernetes Issues

**Pod not starting**:
```bash
# Check pod status
kubectl describe pod <pod-name>

# Check logs
kubectl logs <pod-name>

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp
```

**Service not accessible**:
```bash
# Check service endpoints
kubectl get endpoints

# Test from inside cluster
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://devices-backend.default.svc.cluster.local:8080/health
```

## Best Practices

### Code Quality

1. **Write tests first** (TDD when possible)
2. **Keep functions small** (single responsibility)
3. **Use meaningful names** for variables and functions
4. **Comment complex logic** but prefer self-documenting code
5. **Handle errors gracefully** with proper logging

### Performance

1. **Use connection pooling** for databases
2. **Implement pagination** for large datasets
3. **Add indexes** for frequently queried columns
4. **Cache expensive computations**
5. **Use async/await** for I/O operations (Python)

### Security

1. **Never commit secrets** to git
2. **Validate all inputs** (use Pydantic, struct tags)
3. **Use parameterized queries** (ORMs handle this)
4. **Implement rate limiting** for APIs
5. **Keep dependencies updated**

### Documentation

1. **Update docs with code changes**
2. **Include code examples** in documentation
3. **Document breaking changes** prominently
4. **Keep README up to date**
5. **Use OpenAPI/Swagger** for API documentation

## Resources

### Official Documentation
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [Gin Documentation](https://gin-gonic.com/docs/)
- [React Documentation](https://react.dev/)
- [Vite Documentation](https://vitejs.dev/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)

### Learning Resources
- [Python Testing with pytest](https://docs.pytest.org/)
- [Go Testing Guide](https://golang.org/doc/tutorial/add-a-test)
- [React Testing Library](https://testing-library.com/react)
- [Git Workflow](https://www.atlassian.com/git/tutorials/comparing-workflows)

### Tools
- [Black (Python formatter)](https://black.readthedocs.io/)
- [golangci-lint (Go linter)](https://golangci-lint.run/)
- [Prettier (JS formatter)](https://prettier.io/)
- [act (Run GitHub Actions locally)](https://github.com/nektos/act)

## Getting Help

- Check existing documentation in `docs/`
- Review [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
- Search GitHub issues
- Ask in team chat/discussions
- Create a new issue with detailed information

## Contributing

We welcome contributions! Please:
1. Read this development guide
2. Follow coding standards
3. Write tests for new features
4. Update documentation
5. Submit a well-described PR

Thank you for contributing to Raqeem! 🎉
