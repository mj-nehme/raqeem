# Raqeem — IoT Device Monitoring Platform

[![CI](https://github.com/mj-nehme/raqeem/actions/workflows/ci.yml/badge.svg)](https://github.com/mj-nehme/raqeem/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/mj-nehme/raqeem/branch/master/graph/badge.svg)](https://codecov.io/gh/mj-nehme/raqeem)
[![Release](https://img.shields.io/github/v/release/mj-nehme/raqeem)](https://github.com/mj-nehme/raqeem/releases)

> Full-stack IoT device monitoring platform with real-time telemetry, alerts, and analytics

## Quick Start

```bash
# Start everything (one command)
./start.sh

# Stop everything
./stop.sh
```

**Prerequisites**: Docker Desktop with Kubernetes, kubectl, helm, node/npm

First time? → [Setup Guide](docs/FIRST_TIME_SETUP.md)

## What It Does

- **Monitor IoT Devices** — Track CPU, memory, disk, network metrics in real-time
- **Capture Screenshots** — Upload and view device screens via S3 storage
- **Send Alerts** — Configurable severity levels with auto-forwarding
- **Simulate Data** — Built-in device simulator for testing

## Tech Stack

`Python/FastAPI` • `Go/Gin` • `React/Vite` • `PostgreSQL` • `MinIO` • `Kubernetes`

## Access Points

After `./start.sh`, services auto-discover ports:
- **Monitor Dashboard** — View all devices and metrics
- **Device Simulator** — Generate test data
- **API Docs** — http://localhost:30080/docs (FastAPI)
- **Service URLs** — `./scripts/discover.sh list`

## Documentation

### Getting Started
- [First Time Setup](docs/FIRST_TIME_SETUP.md) — Complete installation guide
- [Architecture](docs/ARCHITECTURE.md) — System design and data flow
- [API Documentation](docs/API.md) — REST endpoints and examples

### Development
- [Development Guide](docs/DEVELOPMENT.md) — Local setup and coding standards
- [Testing Guide](docs/TESTING.md) — Unit, integration, and E2E testing
- [Troubleshooting](docs/TROUBLESHOOTING.md) — Common issues and fixes

### Operations
- [Deployment Guide](docs/DEPLOYMENT.md) — Production Kubernetes deployment
- [Version Management](docs/VERSION_MANAGEMENT.md) — Release workflow
- [Local CI](docs/LOCAL_CI.md) — Run GitHub Actions locally

## Glossary

| Term | Definition |
|------|------------|
| **Device** | IoT endpoint being monitored (CPU, memory, etc.) |
| **Telemetry** | Real-time metrics data (performance, usage stats) |
| **Activity** | User actions logged on devices (file access, apps) |
| **Alert** | Notification triggered by metric thresholds |
| **Screenshot** | Visual capture of device screen stored in S3 |
| **Mentor** | Management interface for viewing all devices |
| **Simulator** | Tool for generating realistic test data |
| **Service Discovery** | Auto-detection of service URLs and ports |

## Architecture

```
Device Simulator → Devices Backend (FastAPI) → PostgreSQL
                     ↓ alerts
Monitor Dashboard ← Mentor Backend (Go) → Screenshots (MinIO)
```

## Contributing

1. Run tests: `pytest`, `go test ./...`, `npm test`
2. Follow existing code style
3. Update documentation for new features
4. Create atomic, well-described commits

## License

MIT License
