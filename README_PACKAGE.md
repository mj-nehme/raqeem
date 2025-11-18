# Raqeem v0.2.0 - Package Release

## Quick Summary

Raqeem is a full-stack IoT device monitoring platform with real-time telemetry, alerts, and analytics. This release includes enhanced reliability, comprehensive test coverage, and consistent Swagger/OpenAPI documentation across all backends.

## What's Included

### Core Components
- **Devices Backend** (FastAPI/Python): High-throughput telemetry ingestion
- **Mentor Backend** (Go/Gin): Device management & analytics  
- **Devices Frontend** (React/Vite): Interactive device simulator
- **Mentor Frontend** (React/Vite): Monitoring dashboard
- **Database** (PostgreSQL): Persistent storage
- **Object Storage** (MinIO): S3-compatible screenshot storage

### Key Features
- Real-time device metrics (CPU, memory, disk, network)
- Screenshot capture and viewing
- Alert system with configurable severity
- Remote command execution
- Device activity tracking
- Built-in device simulator

### Test Coverage
- Comprehensive test suite with 310+ tests
- Backend unit tests (Python and Go)
- Frontend tests (Devices and Mentor)
- Integration and end-to-end tests
- Battle-tested validation
- **90% coverage target maintained**

### Code Quality
- Clean linting (ruff, golangci-lint, ESLint)
- Type checking with mypy
- Comprehensive Swagger/OpenAPI documentation for all APIs
- No critical security vulnerabilities
- Detailed documentation

## Installation

### Docker/Kubernetes (Recommended)
```bash
git clone https://github.com/mj-nehme/raqeem.git
cd raqeem
./start.sh
```

### Python Package
```bash
pip install raqeem
```

### From Source
```bash
git clone https://github.com/mj-nehme/raqeem.git
cd raqeem
pip install -e .
```

## Quick Start

1. **Start Services**
   ```bash
   ./start.sh
   ```

2. **Access Dashboards**
   - Monitor Dashboard: Auto-discovered port
   - Device Simulator: Auto-discovered port
   - Devices API: http://localhost:30080/docs
   - Mentor API: http://localhost:30090/docs

3. **View Service URLs**
   ```bash
   ./scripts/discover.sh list
   ```

## Documentation

- [Architecture Guide](docs/ARCHITECTURE.md)
- [API Documentation](docs/API.md)
- [Development Guide](docs/DEVELOPMENT.md)
- [Testing Guide](docs/TESTING.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)

## Requirements

- Docker Desktop with Kubernetes enabled
- kubectl, helm (for Kubernetes deployment)
- Node.js 20+ (for frontend development)
- Python 3.11+ (for backend development)
- Go 1.25+ (for mentor backend development)

## Support

- [Report Issues](https://github.com/mj-nehme/raqeem/issues)
- [Contributing Guide](CONTRIBUTING.md)
- [Changelog](CHANGELOG.md)

## License

MIT License - See [LICENSE](LICENSE) file

---

**Version**: 0.2.0  
**Release Date**: 2025-11-16  
**Status**: Stable
