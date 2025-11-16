# Raqeem — IoT Device Monitoring Platform

[![CI](https://github.com/mj-nehme/raqeem/actions/workflows/ci.yml/badge.svg)](https://github.com/mj-nehme/raqeem/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/mj-nehme/raqeem/branch/master/graph/badge.svg)](https://codecov.io/gh/mj-nehme/raqeem)
[![Release](https://img.shields.io/github/v/release/mj-nehme/raqeem?sort=semver&display_name=tag)](https://github.com/mj-nehme/raqeem/releases)
[![Go Version](https://img.shields.io/badge/go-1.25%2B-blue)](https://golang.org)
[![Python Version](https://img.shields.io/badge/python-3.11%2B-blue)](https://python.org)

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
- **Devices API Docs** — http://localhost:30080/docs (FastAPI Swagger UI)
- **Mentor API Docs** — http://localhost:30081/docs (Swagger UI)
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
- [GHCR Migration Guide](docs/GHCR_MIGRATION.md) — Migrating from DockerHub to GHCR
- [Version Management](docs/VERSION_MANAGEMENT.md) — Release workflow
- [Local CI](docs/LOCAL_CI.md) — Run GitHub Actions locally
- [Branch Cleanup](docs/BRANCH_CLEANUP.md) — Managing and cleaning up branches

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
┌──────────────────────┐     ┌──────────────────────┐
│  Mentor Frontend     │     │  Devices Frontend    │
│  (React/Vite)        │     │  (React/Vite)        │
│  - Device Dashboard  │     │  - Device Simulator  │
│  - Metrics & Charts  │     │  - Auto-simulation   │
│  - Screenshots       │     │  - Manual Controls   │
│  - Activities/Alerts │     │  - Test Data Gen     │
└──────────┬───────────┘     └──────────┬───────────┘
           │ HTTP                       │ HTTP
           │                            │
      ┌────▼────────┐              ┌────▼──────────┐
      │   Mentor    │              │   Devices     │
      │  Backend    │◄─────────────│   Backend     │
      │    (Go)     │  Alert Fwd   │  (FastAPI)    │
      └──────┬──────┘              └───────┬───────┘
             │                             │
             └─────────────┬───────────────┘
                           │
        ┌──────────────────▼──────────────────┐
        │         PostgreSQL                  │
        │  - Devices, Metrics, Activities     │
        │  - Users, Alerts, Commands          │
        └──────────────────┬──────────────────┘
                           │
        ┌──────────────────▼──────────────────┐
        │       MinIO (S3-compatible)         │
        │  - Screenshot Storage               │
        │  - Presigned URLs                   │
        └─────────────────────────────────────┘
```

### Components

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Devices Backend** | FastAPI (Python) | High-throughput telemetry ingestion |
| **Mentor Backend** | Go + Gin | Device management & analytics |
| **Mentor Frontend** | React + Vite | Monitoring dashboard |
| **Devices Frontend** | React + Vite | Interactive device simulator |
| **PostgreSQL** | Database | Persistent storage for all data |
| **MinIO** | Object Storage | S3-compatible screenshot storage |

### Data Flow

1. **Device Registration** → Device Simulator → Devices Backend → PostgreSQL
2. **Telemetry Ingestion** → Device Simulator → Devices Backend → PostgreSQL
3. **Alert Forwarding** → Devices Backend → Mentor Backend → PostgreSQL
4. **Screenshot Upload** → Device Simulator → Devices Backend → MinIO
5. **Dashboard Display** → Mentor Frontend → Mentor Backend → PostgreSQL + MinIO

## Contributing

1. Run tests: `pytest`, `go test ./...`, `npm test`
2. Follow existing code style
3. Update documentation for new features
4. Create atomic, well-described commits

## License

MIT License
