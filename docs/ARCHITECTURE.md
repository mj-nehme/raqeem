# 🏗️ Architecture Documentation

## Overview

Raqeem is a full-stack IoT device monitoring platform built with microservices architecture, designed for scalability, maintainability, and cloud-native deployment. The system collects real-time telemetry from distributed devices, processes alerts, and provides a unified monitoring dashboard.

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                       Web Browser                               │
├──────────────────┬──────────────────────────────────────────────┤
│   Mentor         │          Devices                             │
│   Frontend       │          Frontend                            │
│   (React/Vite)   │          (React/Vite)                       │
│                  │                                              │
│ - Dashboard      │       - Device Simulator                     │
│ - Device List    │       - Auto-simulation                      │
│ - Metrics Charts │       - Manual Controls                      │
│ - Screenshots    │       - Test Data Gen                        │
│ - Activities     │                                              │
│ - Alerts View    │                                              │
└────────┬─────────┴──────────────┬───────────────────────────────┘
         │ HTTP/REST              │ HTTP/REST
         │                        │
    ┌────▼────────────┐      ┌────▼──────────────┐
    │   Mentor        │      │    Devices        │
    │   Backend       │◄─────│    Backend        │
    │   (Go + Gin)    │ Fwd  │  (Python/FastAPI) │
    │                 │      │                   │
    │ Port: 30081     │      │  Port: 30080      │
    └────────┬────────┘      └──────────┬────────┘
             │                          │
             │  Kubernetes DNS          │
             │  Service Discovery       │
             │                          │
             └────────────┬─────────────┘
                          │
         ┌────────────────▼────────────────┐
         │         PostgreSQL              │
         │    (Persistent Storage)         │
         │                                 │
         │  - devices table                │
         │  - device_metrics table         │
         │  - device_processes table       │
         │  - device_activity table        │
         │  - device_alerts table          │
         │  - remote_commands table        │
         │  - screenshots table            │
         │                                 │
         │  Port: 30432                    │
         └────────────────┬────────────────┘
                          │
         ┌────────────────▼────────────────┐
         │       MinIO                     │
         │    (S3-Compatible Storage)      │
         │                                 │
         │  - Screenshot Storage           │
         │  - Presigned URLs               │
         │                                 │
         │  API Port: 9000                 │
         │  Console Port: 30001            │
         └─────────────────────────────────┘
```

## Core Components

### 1. Devices Backend (Python/FastAPI)

**Purpose**: High-throughput telemetry ingestion and device data collection

**Technology Stack**:
- Python 3.11+
- FastAPI web framework
- SQLAlchemy ORM
- PostgreSQL driver
- S3/MinIO client

**Responsibilities**:
- Device registration and management
- Real-time metrics ingestion (CPU, memory, disk, network)
- Activity logging (app usage, file access)
- Alert collection and forwarding
- Screenshot upload to MinIO
- Process tracking
- Keystroke/location logging (legacy features)

**Key Endpoints**:
- `POST /api/v1/devices/register` - Register new device
- `POST /api/v1/metrics` - Submit device metrics
- `POST /api/v1/activities` - Log device activity
- `POST /api/v1/alerts` - Report alerts (auto-forwards to mentor)
- `POST /api/v1/screenshots` - Upload screenshots to MinIO
- `GET /docs` - Interactive API documentation (OpenAPI/Swagger)

**Port**: 30080 (NodePort on Kubernetes)

### 2. Mentor Backend (Go/Gin)

**Purpose**: Device management, analytics, and monitoring dashboard API

**Technology Stack**:
- Go 1.23+
- Gin web framework
- GORM (ORM)
- PostgreSQL driver

**Responsibilities**:
- Device listing and status
- Metrics retrieval and aggregation
- Alert management and display
- Remote command execution
- Screenshot retrieval with presigned URLs
- Activity log access
- Process information

**Key Endpoints**:
- `GET /devices` - List all devices
- `GET /devices/:id/metrics` - Get device metrics
- `GET /devices/:id/alerts` - Get device alerts
- `GET /devices/:id/screenshots` - Get screenshot URLs
- `POST /devices/:id/alerts` - Report alert (from devices backend)
- `POST /devices/commands` - Create remote command
- `GET /health` - Health check

**Port**: 30081 (NodePort on Kubernetes)

### 3. Mentor Frontend (React/Vite)

**Purpose**: Web-based monitoring dashboard for viewing device status and metrics

**Technology Stack**:
- React 18
- Vite (build tool)
- TypeScript/JavaScript
- Chart libraries for metrics visualization

**Features**:
- Device list with online/offline status
- Real-time metrics charts (CPU, memory, disk, network)
- Alert viewing with severity levels
- Screenshot gallery with thumbnails
- Activity timeline
- Device detail views
- Responsive UI

**Port**: Auto-detected (typically 5001+)

### 4. Devices Frontend (React/Vite)

**Purpose**: Device simulator for testing and development

**Technology Stack**:
- React 18
- Vite (build tool)
- JavaScript

**Features**:
- Device registration interface
- Auto-simulation mode with periodic metric generation
- Manual control for sending specific metrics
- Screenshot capture simulation
- Activity generation
- Alert triggering

**Port**: Auto-detected (typically 4000+)

### 5. PostgreSQL Database

**Purpose**: Persistent data storage for all device data

**Technology**: PostgreSQL 16

**Schema**: See [Database Schema](#database-schema) section

**Port**: 30432 (NodePort), 5432 (internal)

**Storage**: Persistent Volume Claim (survives pod restarts)

### 6. MinIO Object Storage

**Purpose**: S3-compatible storage for screenshots and binary data

**Technology**: MinIO (S3-compatible)

**Features**:
- Bucket: `screenshots`
- Presigned URL generation for secure access
- Web console for management

**Ports**: 
- API: 9000 (internal)
- Console: 30001 (NodePort)

## Data Flow Patterns

### 1. Device Registration Flow

```
Device Simulator → POST /devices/register → Devices Backend
                                           ↓
                                      Insert into DB
                                           ↓
                                   Return device_id
```

### 2. Metrics Collection Flow

```
Device Simulator → POST /metrics → Devices Backend
                                   ↓
                              Validate & store
                                   ↓
                        Insert into device_metrics
```

### 3. Alert Flow (Bidirectional Communication)

```
Device Simulator → POST /alerts → Devices Backend
                                   ↓
                          1. Store in devices DB
                                   ↓
                          2. Forward to Mentor Backend
                                   ↓
                          POST /devices/:id/alerts
                                   ↓
                          Mentor Backend stores alert
                                   ↓
                          Dashboard displays alert
```

**Key Design Decision**: Devices backend forwards alerts to mentor backend to enable centralized monitoring. If forwarding fails, the alert is still stored locally (fire-and-forget pattern).

### 4. Screenshot Upload Flow

```
Device Simulator → POST /screenshots → Devices Backend
                                       ↓
                                 Upload to MinIO
                                       ↓
                            Store metadata in DB
                                       ↓
           Dashboard → GET /screenshots → Mentor Backend
                                          ↓
                                  Generate presigned URL
                                          ↓
                                  Return URL to frontend
                                          ↓
                            Frontend fetches from MinIO
```

### 5. Remote Command Flow

```
Dashboard → POST /devices/commands → Mentor Backend
                                     ↓
                            Store as "pending"
                                     ↓
Device Polls → GET /devices/:id/commands/pending
                                     ↓
                            Execute command locally
                                     ↓
                POST /commands/status → Update status
```

## Database Schema

### Tables and Relationships

#### 1. `devices` (Devices & Mentor)
Primary table for device registration and status

| Column | Type | Description |
|--------|------|-------------|
| id | String (PK) | Device unique identifier |
| name | Text | Human-readable device name |
| type | Text | Device type (laptop, desktop, mobile) |
| os | Text | Operating system |
| last_seen | Timestamp | Last heartbeat/update time |
| is_online | Boolean | Current online status |
| location | Text | Physical location |
| ip_address | Text | Network IP address |
| mac_address | Text | MAC address |
| current_user | Text | Currently logged-in user |

#### 2. `device_metrics` (Devices & Mentor)
Time-series metrics data

| Column | Type | Description |
|--------|------|-------------|
| id | UUID (PK) | Metric record ID |
| device_id | String (FK) | Reference to devices.id |
| timestamp | Timestamp | When metrics were collected |
| cpu_usage | Float | CPU usage percentage |
| cpu_temp | Float | CPU temperature (Celsius) |
| memory_total | BigInt | Total memory (bytes) |
| memory_used | BigInt | Used memory (bytes) |
| swap_used | BigInt | Swap usage (bytes) |
| disk_total | BigInt | Total disk space (bytes) |
| disk_used | BigInt | Used disk space (bytes) |
| net_bytes_in | BigInt | Network bytes received/sec |
| net_bytes_out | BigInt | Network bytes sent/sec |

**Indexing**: Index on `device_id` and `timestamp` for efficient queries

#### 3. `device_processes` (Devices & Mentor)
Running processes snapshot

| Column | Type | Description |
|--------|------|-------------|
| id | UUID (PK) | Process record ID |
| device_id | String (FK) | Reference to devices.id |
| timestamp | Timestamp | When snapshot was taken |
| pid | Integer | Process ID |
| name | Text | Process name |
| cpu | Float | CPU usage percentage |
| memory | BigInt | Memory usage (bytes) |
| command | Text | Full command line |

#### 4. `device_activity` (Devices & Mentor)
User activity logs

| Column | Type | Description |
|--------|------|-------------|
| id | UUID (PK) | Activity record ID |
| device_id | String (FK) | Reference to devices.id |
| timestamp | Timestamp | When activity occurred |
| type | Text | Activity type (app_launch, file_access, etc.) |
| description | Text | Activity description |
| app | Text | Application name |
| duration | Integer | Activity duration (seconds) |

#### 5. `device_alerts` (Devices & Mentor)
System alerts and warnings

| Column | Type | Description |
|--------|------|-------------|
| id | UUID (PK) | Alert record ID |
| device_id | String (FK) | Reference to devices.id |
| timestamp | Timestamp | When alert was triggered |
| level | Text | Severity (low, medium, high, critical) |
| type | Text | Alert type (cpu, memory, disk, network, security) |
| message | Text | Human-readable alert message |
| value | Float | Measured value that triggered alert |
| threshold | Float | Threshold that was exceeded |

#### 6. `remote_commands` (Mentor only)
Remote command execution

| Column | Type | Description |
|--------|------|-------------|
| id | Integer (PK) | Command ID |
| device_id | String (FK) | Reference to devices.id |
| command | Text | Command to execute |
| status | Text | Status (pending, running, completed, failed) |
| created_at | Timestamp | When command was created |
| completed_at | Timestamp | When command finished |
| result | Text | Command output |
| exit_code | Integer | Exit code |

#### 7. `screenshots` (Mentor only)
Screenshot metadata

| Column | Type | Description |
|--------|------|-------------|
| id | Integer (PK) | Screenshot ID |
| device_id | String (FK) | Reference to devices.id |
| timestamp | Timestamp | When screenshot was taken |
| path | Text | Path in MinIO (bucket/object key) |
| resolution | Text | Image resolution (e.g., "1920x1080") |
| size | BigInt | File size (bytes) |

### Database Design Considerations

**Data Duplication**: Both databases (devices and mentor) have the same schema for core tables (devices, metrics, alerts) to enable:
- Independent operation of each service
- Reduced inter-service dependencies
- Faster local queries
- Resilience to network issues

**Time-Series Optimization**: Metrics and activity tables can grow large. Consider:
- Partitioning by time (monthly/weekly)
- Archiving old data
- Aggregation for long-term storage

**Indexing Strategy**:
- Primary keys on all tables
- Foreign keys on device_id columns
- Composite indexes on (device_id, timestamp) for time-series queries
- Index on alert.level for filtering

## Service Communication

### Kubernetes Service Discovery

Services communicate via Kubernetes DNS, eliminating hardcoded IPs:

```
postgres-service.default.svc.cluster.local:5432
minio-service.default.svc.cluster.local:9000
devices-backend.default.svc.cluster.local:8080
mentor-backend.default.svc.cluster.local:8080
```

**Format**: `<service-name>.<namespace>.svc.cluster.local`

### CORS Configuration

Both backends support CORS for browser access:
- Mentor Backend: Configurable via `FRONTEND_ORIGIN` environment variable
- Devices Backend: Configured for local development

### API Communication Patterns

1. **Synchronous HTTP**: Frontend ↔ Backend (REST APIs)
2. **Fire-and-forget**: Devices Backend → Mentor Backend (alert forwarding)
3. **Polling**: Device Simulator → Mentor Backend (pending commands)

## Security Considerations

### 1. Environment-Based Configuration

**No hardcoded credentials** - All sensitive data in environment variables:
- Database passwords
- MinIO access keys
- API URLs
- CORS origins

### 2. Network Security

- **Internal communication**: Services communicate within Kubernetes cluster
- **NodePort services**: Exposed ports for external access (development)
- **Production**: Use LoadBalancer or Ingress with TLS

### 3. Authentication (Future Enhancement)

Current MVP has no authentication. Recommended additions:
- JWT-based authentication
- API key authentication for devices
- Role-based access control (RBAC)
- OAuth2 for frontend

### 4. Data Security

- **PostgreSQL**: Network encryption available
- **MinIO**: Presigned URLs expire (time-limited access)
- **Secrets**: Kubernetes Secrets for sensitive data

### 5. Input Validation

- FastAPI: Pydantic models for request validation
- Go: Manual validation with error handling
- SQL: Parameterized queries (ORM protection against SQL injection)

## Deployment Architecture

### Kubernetes Resources

#### Namespaces
- `default`: All services (can be customized)

#### Deployments
- `postgres`: 1 replica
- `minio`: 1 replica
- `devices-backend`: 1+ replicas (scalable)
- `mentor-backend`: 1+ replicas (scalable)

#### Services
- `postgres-service`: ClusterIP + NodePort (30432)
- `minio-service`: ClusterIP (9000, 9001) + NodePort (30001)
- `devices-backend`: ClusterIP + NodePort (30080)
- `mentor-backend`: ClusterIP + NodePort (30081)

#### Persistent Volumes
- `postgres-pvc`: 1Gi (database data)
- `minio-pvc`: 1Gi (screenshot storage)

### Helm Chart Structure

```
charts/
├── postgres/
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
│       ├── deployment.yaml
│       ├── service.yaml
│       └── pvc.yaml
├── minio/
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
├── devices-backend/
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
└── mentor-backend/
    ├── Chart.yaml
    ├── values.yaml
    └── templates/
```

### Service Discovery System

**Smart Discovery** (`.deploy/registry/`):
- Auto-detects available ports for frontends
- Registers service URLs in local registry
- Provides CLI for service lookup (`./scripts/discover.sh`)
- Eliminates need for .env files

## Scalability Considerations

### Horizontal Scaling

**Backends**:
```bash
kubectl scale deployment devices-backend --replicas=3
kubectl scale deployment mentor-backend --replicas=3
```

**Database**: Single instance (consider PostgreSQL replication for HA)

**MinIO**: Single instance (consider distributed MinIO for HA)

### Performance Optimization

1. **Database Connection Pooling**: Both backends use connection pools
2. **Caching**: Add Redis for frequently accessed data
3. **CDN**: Serve static frontend assets via CDN
4. **Database Indexes**: Optimize query performance
5. **Metrics Aggregation**: Pre-aggregate metrics for dashboard queries

### Monitoring and Observability

**Current**:
- Health check endpoints on all services
- Application logs via `kubectl logs`

**Recommended**:
- Prometheus for metrics collection
- Grafana for visualization
- Loki for log aggregation
- Jaeger for distributed tracing
- Alertmanager for incident management

## Technology Choices Rationale

### Why FastAPI (Devices Backend)?
- **High throughput**: Async/await for handling many concurrent requests
- **Auto-documentation**: OpenAPI specs generated automatically
- **Type safety**: Pydantic for data validation
- **Python ecosystem**: Rich libraries for data processing

### Why Go + Gin (Mentor Backend)?
- **Performance**: Compiled language, low memory footprint
- **Concurrency**: Goroutines for handling multiple requests
- **Simple deployment**: Single binary
- **Type safety**: Strong typing

### Why React + Vite?
- **Fast development**: Hot module replacement
- **Modern tooling**: ES modules, fast builds
- **Component reusability**: React ecosystem
- **TypeScript support**: Type safety in frontend

### Why Kubernetes?
- **Consistency**: Same setup across all environments
- **Service discovery**: Built-in DNS
- **Scalability**: Easy horizontal scaling
- **Portability**: Works on any Kubernetes cluster

### Why PostgreSQL?
- **Reliability**: ACID compliance
- **Rich features**: JSON support, full-text search
- **Performance**: Excellent for read-heavy workloads
- **Open source**: No licensing costs

### Why MinIO?
- **S3 compatibility**: Standard API
- **Self-hosted**: No cloud dependencies
- **Lightweight**: Easy to run locally
- **Production-ready**: Can scale to petabytes

## Future Architecture Enhancements

1. **Message Queue**: Add RabbitMQ/Kafka for async processing
2. **Caching Layer**: Redis for session storage and caching
3. **API Gateway**: Kong/Traefik for routing and rate limiting
4. **Service Mesh**: Istio for advanced traffic management
5. **Distributed Tracing**: OpenTelemetry for request tracking
6. **Event Sourcing**: Store all state changes as events
7. **GraphQL**: Unified API layer for flexible queries
8. **WebSockets**: Real-time updates without polling

## References

- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [Gin Web Framework](https://gin-gonic.com/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Helm Documentation](https://helm.sh/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [MinIO Documentation](https://min.io/docs/)
- [React Documentation](https://react.dev/)
- [Vite Documentation](https://vitejs.dev/)
