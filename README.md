# Raqeem — IoT Device Monitoring Platform

[![CI](https://github.com/mj-nehme/raqeem/actions/workflows/ci.yml/badge.svg)](https://github.com/mj-nehme/raqeem/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/mj-nehme/raqeem/branch/master/graph/badge.svg)](https://codecov.io/gh/mj-nehme/raqeem)
[![Release](https://img.shields.io/github/v/release/mj-nehme/raqeem)](https://github.com/mj-nehme/raqeem/releases)
[![Go Version](https://img.shields.io/badge/go-1.23%2B-blue)](https://golang.org)
[![Python Version](https://img.shields.io/badge/python-3.11%2B-blue)](https://python.org)

> 🔍 Full-stack IoT device monitoring platform with real-time telemetry, alerts, and analytics — built with FastAPI, Go, React, and Kubernetes

A production-ready monitoring platform for tracking IoT devices, collecting telemetry, and analyzing device behavior in real-time. Built with modern cloud-native technologies and Kubernetes-first architecture.

## 🎯 What It Does

- **Device Management** — Register, track, and monitor devices with customizable properties
- **Real-Time Telemetry** — Collect metrics (CPU, memory, disk, network) from distributed devices
- **Activity Logging** — Track application usage, file access, and network activity
- **Alert System** — Configurable alerts with severity levels (low/medium/high/critical)
- **Screenshot Capture** — Upload and view device screenshots via S3-compatible storage
- **Interactive Dashboard** — Monitor all devices from a unified web interface
- **Device Simulator** — Built-in testing tool for generating realistic telemetry data

## ✨ Key Features

- ⚡ **One-command deployment** — `./start.sh` spins up the entire stack in minutes
- 🔄 **Auto-forwarding alerts** — Devices backend forwards critical alerts to mentor backend
- 📦 **Semantic versioning** — Automated tagging and release workflow with Docker image management
- 🧪 **Comprehensive testing** — Unit tests (pytest, go test, vitest), integration tests, smoke tests
- 🔐 **Security-first** — Environment-based configuration, no hardcoded credentials
- 📊 **Health checks** — Built-in endpoints for reliability monitoring
- 🌐 **Service discovery** — Kubernetes DNS for zero-configuration service communication

## 🛠️ Tech Stack

`Python` • `FastAPI` • `Go` • `Gin` • `React` • `Vite` • `PostgreSQL` • `MinIO` • `Kubernetes` • `Helm` • `Docker` • `GitHub Actions`

---

## 🚀 Quick Start

### Prerequisites
- Docker Desktop with Kubernetes enabled ([Download](https://www.docker.com/products/docker-desktop))
- `kubectl` - Kubernetes CLI ([Install](https://kubernetes.io/docs/tasks/tools/))
- `helm` 3.x - Kubernetes package manager ([Install](https://helm.sh/docs/intro/install/))
- `node` & `npm` - JavaScript runtime ([Install](https://nodejs.org/))

**First time?** Check out the [📚 First Time Setup Guide](docs/FIRST_TIME_SETUP.md) for detailed instructions.

### Start Everything
```bash
# Start the platform with smart service discovery
./start.sh
```

That's it! **No configuration needed.** The smart discovery system will:
- ✅ Deploy PostgreSQL and MinIO
- ✅ Deploy both backends with stable NodePort services
- ✅ Auto-detect available ports for frontends
- ✅ Register all services in a discovery registry
- ✅ Launch frontend applications with proper backend URLs

### Advanced Configuration (Optional)
```bash
# Customize frontend starting ports
DEVICES_FRONTEND_START_PORT=6000 MENTOR_FRONTEND_START_PORT=7000 ./start.sh

# Use different Kubernetes namespace
NAMESPACE=dev ./start.sh
```

### Stop Everything
```bash
./stop.sh
```

### Access Services

After starting, services are auto-discovered and available at:
- **Mentor Dashboard** — `http://localhost:<auto-detected>` (monitoring interface)
- **Device Simulator** — `http://localhost:<auto-detected>` (test data generator)  
- **Devices API** — `http://localhost:30080/docs` (FastAPI interactive docs)
- **Mentor API** — `http://localhost:30081` (Go backend)

Use `./scripts/discover.sh list` to see actual discovered URLs.

*Frontend ports are auto-detected to avoid conflicts*

---

## � Service Discovery

The platform includes smart service discovery that automatically handles port conflicts and service registration.

### Discovery Commands
```bash
# List all registered services
./scripts/discover.sh list

# Get specific service URL
./scripts/discover.sh get devices-backend

# Check service health
./scripts/discover.sh health

# Wait for a service to be ready
./scripts/discover.sh wait mentor-backend 60
```

### How It Works
- **Zero Configuration**: No .env files needed, uses smart defaults
- **Auto-Port Detection**: Finds available ports automatically, no conflicts
- **Service Registry**: All services register their URLs in `.deploy/registry/`
- **Kubernetes NodePort**: Backends use stable NodePort services (no port-forwarding)
- **Health Monitoring**: Built-in health checks and service verification

---

## �📚 Documentation

- **[First Time Setup](docs/FIRST_TIME_SETUP.md)** - Complete beginner's guide
- **[Testing Guide](docs/TESTING.md)** - Unit, integration, and E2E tests
- **[Local CI](docs/LOCAL_CI.md)** - Run GitHub Actions locally with `act`
- **[Version Management](docs/VERSION_MANAGEMENT.md)** - Release workflow
- **[Release Workflow](docs/RELEASE_WORKFLOW.md)** - Pre-release testing
- **[API Documentation](docs/)** - OpenAPI specs for both backends

---

## Quick Start (Automated)

### Start Everything
Run the automated script to deploy all services in proper order:

```bash
./start.sh
# or
./scripts/start.sh
```

This will automatically:
1. Load configuration from `.env` file
2. Deploy PostgreSQL and wait for it to be ready
3. Deploy MinIO and wait for it to be ready  
4. Deploy Devices Backend and wait for it to be ready
5. Deploy Mentor Backend and wait for it to be ready
6. Start port-forwards for backends (ports configurable via .env)
7. Start both frontends (ports configurable via .env)

**Deployment Order & Error Handling:**
- Services are deployed sequentially in proper dependency order
- Each service must be ready before proceeding to the next
- If any step fails, the script exits with a clear error message showing:
  - What went wrong
  - Which services were successfully started

---

## 📦 Version Management

Create tagged releases linked to specific commits:

```bash
# Create a versioned release (e.g., v1.0.0)
./scripts/tag-release.sh v1.0.0
```

This automatically:
1. ✅ Validates code (Go compilation, Python syntax, Helm charts)
2. ✅ Builds Docker images with version tags
3. ✅ Tags images as: `v1.0.0`, `v1.0.0-<git-sha>`, and `latest`
4. ✅ Pushes images to Docker Hub
5. ✅ Updates Helm charts to use the new version
6. ✅ Creates a git tag linking the version to the commit
7. ✅ Generates release notes

**Deploy a specific version:**
```bash
echo "IMAGE_TAG=v1.0.0" > .deploy/tag.env
./start.sh
```

See [Version Management docs](docs/VERSION_MANAGEMENT.md) for details.

---

## 🧪 Testing

### Run All Tests
```bash
# Python backend tests
cd devices/backend && pytest -v

# Go backend tests
cd mentor/backend/src && go test ./...

# Frontend tests
cd devices/frontend && npm test
cd mentor/frontend && npm test
```

### Integration Tests
```bash
# End-to-end test with docker-compose
./tests/integration/run_integration_tests.sh

# Smoke test (requires running services)
python3 tests/smoke_test.py
```

### CI/CD
Tests run automatically on every push via GitHub Actions. Run CI locally:
```bash
brew install act  # macOS
act -j build-and-test
```

See [Testing Guide](docs/TESTING.md) and [Local CI](docs/LOCAL_CI.md) for details.

---

## 🔧 Development

### Local Development Options

**Option 1: Full Stack in Kubernetes (Recommended)**
```bash
./start.sh  # Everything runs in K8s with port-forwards
```

**Option 2: Infrastructure in K8s, Backends Local**
```bash
# Deploy infrastructure
helm upgrade --install postgres ./charts/postgres -n default
helm upgrade --install minio ./charts/minio -n default

# Port-forward
kubectl port-forward svc/postgres-service 5432:5432 -n default &
kubectl port-forward svc/minio-service 9000:9000 9001:9001 -n default &

# Run backends locally
cd mentor/backend/src && go run main.go
cd devices/backend/src && uvicorn app.main:app --reload

# Run frontends
cd mentor/frontend && npm run dev
cd devices/frontend && npm run dev
```

### Useful Commands

```bash
# Check service discovery status
./scripts/discover.sh list
./scripts/discover.sh health

# Comprehensive health check
./scripts/health-check.sh

# View service registry
ls -la .deploy/registry/

# Check Kubernetes services
kubectl get svc -n default
kubectl get pods -n default

# View logs
kubectl logs -f deployment/devices-backend -n default
kubectl logs -f deployment/mentor-backend -n default

# Restart services
kubectl rollout restart deployment/devices-backend -n default

# Delete all data (fresh start)
kubectl delete pvc --all -n default
```

---

## ❓ Troubleshooting

### Common Issues

**Pods not starting:**
```bash
kubectl describe pod <pod-name> -n default
kubectl logs <pod-name> -n default
```

**Port conflicts resolved:**
```bash
# The system auto-detects available ports, but if you see issues:
./scripts/discover.sh list  # Check registered services
lsof -i :4000 :5000        # Check what's using common ports

# Or restart with fresh discovery
./stop.sh && ./start.sh
```

**Database connection issues:**
- Verify postgres pod is running: `kubectl get pods -n default`
- Check service exists: `kubectl get svc postgres-service -n default`
- Ensure port-forward is active
- Verify credentials in `.env` match Helm values

**Reset everything:**
```bash
./stop.sh
kubectl delete pvc --all -n default  # Deletes all data
./start.sh
```

See [First Time Setup Guide](docs/FIRST_TIME_SETUP.md) for more troubleshooting tips.

---

## 🤝 Contributing

Contributions are welcome! Please ensure:
- All tests pass (`pytest`, `go test`, `npm test`)
- Code follows existing style conventions
- Documentation is updated for new features
- Commits are atomic and well-described

## 📄 License

This project is licensed under the MIT License.

## 🙏 Acknowledgments

Built with modern cloud-native technologies and best practices for production-ready IoT monitoring.

---

## 🏗️ Architecture

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

### Service Discovery

Services communicate via **Kubernetes DNS** (no hardcoded IPs):
- `postgres-service.default.svc.cluster.local`
- `minio-service.default.svc.cluster.local`
- `devices-backend.default.svc.cluster.local`
- `mentor-backend.default.svc.cluster.local`

---

## Notes

- All services run in the `default` namespace by default
- Persistent data is stored in PersistentVolumeClaims (survives pod restarts)
- NodePorts are configured for direct access without port-forwarding:
  - Postgres: 30432
  - MinIO Console: 30001
  - Mentor Backend: 30081
  - Devices Backend: 30080
- For production/AWS deployment, update Helm values to use RDS, S3, and LoadBalancer services

---

## Next Steps

- Set up CI/CD pipeline for automated deployments
- Add Ingress controller for external access
- Configure monitoring and logging (Prometheus, Grafana)
- Implement auto-scaling for backends
- Add database migrations/initialization scripts
- Set up automated testing in CI/CD

