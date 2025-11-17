# 📡 API Documentation

## ⚠️ Breaking Change Notice

**All API field names now use canonical lowercase underscore naming.**

Legacy field names are no longer supported and will result in a `400 Bad Request` error with a clear migration message.

### Field Name Changes

| Legacy Field | Canonical Field | Context |
|-------------|----------------|---------|
| `id` | `deviceid` | Device registration |
| `name` | `device_name` | Device info |
| `name` | `process_name` | Process info |
| `location` | `device_location` | Device info |
| `type` | `activity_type` | Activities |
| `type` | `alert_type` | Alerts |
| `command` | `command_text` | Processes and commands |

**Example Error Response:**
```json
{
  "detail": "unsupported legacy field: name; use device_name"
}
```

## Overview

Raqeem provides two REST APIs for device monitoring and management. This guide covers authentication, common patterns, and example requests for both APIs.

## Table of Contents

- [API Overview](#api-overview)
- [Authentication](#authentication)
- [Common Patterns](#common-patterns)
- [Devices Backend API](#devices-backend-api)
- [Mentor Backend API](#mentor-backend-api)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Examples](#examples)

## API Overview

### Devices Backend API

**Base URL**: `http://localhost:30080/api/v1` (local) or `https://api.example.com/api/v1` (production)

**Purpose**: High-throughput telemetry ingestion from devices

**Technology**: FastAPI (Python)

**Documentation**: Interactive docs at `/docs` (Swagger UI) and `/redoc` (ReDoc)

**Key Features**:
- Device registration
- Metrics ingestion
- Activity logging
- Alert submission with auto-forwarding
- Screenshot upload to MinIO

### Mentor Backend API

**Base URL**: `http://localhost:30090` (local) or `https://dashboard-api.example.com` (production)

**Purpose**: Device management and monitoring dashboard API

**Technology**: Gin (Go)

**Documentation**: Interactive docs at `/docs` (Swagger UI) and `/swagger/index.html`

**Key Features**:
- Device listing and details
- Metrics retrieval
- Alert management
- Remote command execution
- Screenshot presigned URLs

## Authentication

### Current Status (MVP)

**No authentication required** for MVP release. All endpoints are publicly accessible.

### Future Implementation (Recommended)

For production use, implement:

#### 1. JWT Authentication

```http
POST /auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "secure_password"
}

Response:
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "bearer",
  "expires_in": 3600
}
```

**Use token in requests**:
```http
GET /api/v1/devices
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

#### 2. API Key Authentication

For device-to-backend communication:

```http
POST /api/v1/metrics
X-API-Key: device_12345_api_key_here
Content-Type: application/json

{
  "deviceid": "device-001",
  "cpu_usage": 45.5
}
```

## OpenAPI Specifications and Client Generation

### Accessing OpenAPI Schemas

Both backends provide OpenAPI 3.0 specifications that can be used for documentation, testing, and client generation.

#### Devices Backend (FastAPI)

- **Swagger UI**: http://localhost:30080/docs
- **ReDoc**: http://localhost:30080/redoc
- **OpenAPI JSON**: http://localhost:30080/openapi.json

```bash
# Download OpenAPI spec
curl http://localhost:30080/openapi.json > devices-openapi.json
```

#### Mentor Backend (Go)

- **Swagger UI**: http://localhost:30090/swagger/index.html
- **Docs Redirect**: http://localhost:30090/docs
- **Swagger JSON**: Available in `mentor/backend/src/docs/swagger.json`

```bash
# Access Swagger JSON
cat mentor/backend/src/docs/swagger.json
```

### Generating API Clients

You can generate client libraries in various languages from the OpenAPI specifications.

#### Python Client (using openapi-generator)

```bash
# Install openapi-generator
npm install -g @openapitools/openapi-generator-cli

# Generate Python client for Devices Backend
openapi-generator-cli generate \
  -i http://localhost:30080/openapi.json \
  -g python \
  -o ./clients/python-devices \
  --additional-properties=packageName=raqeem_devices

# Use the generated client
cd ./clients/python-devices
pip install -e .
```

**Example usage**:
```python
from raqeem_devices import ApiClient, Configuration, DevicesApi

config = Configuration(host="http://localhost:30080/api/v1")
client = ApiClient(configuration=config)
api = DevicesApi(client)

# Register a device
device_data = {
    "deviceid": "550e8400-e29b-41d4-a716-446655440000",
    "device_name": "laptop-001",
    "device_type": "laptop"
}
response = api.register_device(device_data)
print(response)
```

#### JavaScript/TypeScript Client

```bash
# Generate TypeScript/Axios client
openapi-generator-cli generate \
  -i http://localhost:30080/openapi.json \
  -g typescript-axios \
  -o ./clients/typescript-devices

# Or using fetch
openapi-generator-cli generate \
  -i http://localhost:30080/openapi.json \
  -g typescript-fetch \
  -o ./clients/typescript-devices-fetch
```

**Example usage**:
```typescript
import { Configuration, DevicesApi } from './clients/typescript-devices';

const config = new Configuration({
  basePath: 'http://localhost:30080/api/v1'
});

const api = new DevicesApi(config);

// Register device
const deviceData = {
  deviceid: '550e8400-e29b-41d4-a716-446655440000',
  device_name: 'laptop-001',
  device_type: 'laptop'
};

const response = await api.registerDevice(deviceData);
console.log(response.data);
```

#### Go Client

```bash
# Generate Go client for Mentor Backend
openapi-generator-cli generate \
  -i mentor/backend/src/docs/swagger.json \
  -g go \
  -o ./clients/go-mentor \
  --additional-properties=packageName=raqeem_mentor

cd ./clients/go-mentor
go mod init github.com/yourorg/raqeem-mentor-client
go mod tidy
```

**Example usage**:
```go
package main

import (
    "context"
    "fmt"
    raqeem "github.com/yourorg/raqeem-mentor-client"
)

func main() {
    cfg := raqeem.NewConfiguration()
    cfg.Servers = raqeem.ServerConfigurations{
        {URL: "http://localhost:30090"},
    }
    client := raqeem.NewAPIClient(cfg)
    
    // List devices
    devices, _, err := client.DevicesAPI.ListDevices(context.Background()).Execute()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Found %d devices\n", len(devices))
}
```

### Alternative Client Generation Tools

#### Using Swagger Codegen

```bash
# Install swagger-codegen
brew install swagger-codegen  # macOS
# or download from https://swagger.io/tools/swagger-codegen/

# Generate Python client
swagger-codegen generate \
  -i http://localhost:30080/openapi.json \
  -l python \
  -o ./clients/python-swagger
```

#### Using oapi-codegen (Go)

For the Mentor Backend, you can also use oapi-codegen:

```bash
# Install oapi-codegen
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

# Generate Go client types and server stubs
oapi-codegen -package mentor -generate types,client \
  mentor/backend/src/docs/swagger.json > mentor_client.go
```

### Testing with OpenAPI Specifications

#### Using Postman

1. Open Postman
2. Click "Import" → "Link"
3. Enter: `http://localhost:30080/openapi.json`
4. Postman will create a collection with all endpoints

#### Using Insomnia

1. Open Insomnia
2. Click "Create" → "Import From" → "URL"
3. Enter: `http://localhost:30080/openapi.json`

#### Using curl with OpenAPI

```bash
# Install openapi-to-postman
npm install -g openapi-to-postman

# Convert to Postman collection
openapi2postmanv2 -s http://localhost:30080/openapi.json \
  -o devices-collection.json -p

# Use with newman (CLI test runner)
npm install -g newman
newman run devices-collection.json
```

## Common Patterns

### Request Headers

```http
Content-Type: application/json
Accept: application/json
```

### Response Format

All successful responses return JSON:

```json
{
  "status": "success",
  "data": { ... }
}
```

Error responses:

```json
{
  "status": "error",
  "message": "Error description",
  "code": "ERROR_CODE"
}
```

### Pagination

For list endpoints:

```http
GET /api/v1/devices?page=1&per_page=50
```

Response:
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 50,
    "total": 150,
    "pages": 3
  }
}
```

### Filtering

```http
GET /api/v1/devices?is_online=true&type=laptop
GET /api/v1/devices?last_seen_after=2024-01-01T00:00:00Z
```

### Sorting

```http
GET /api/v1/devices?sort=last_seen&order=desc
```

## Devices Backend API

### Device Management

#### Register Device

```http
POST /api/v1/devices/register
Content-Type: application/json

{
  "deviceid": "device-001",
  "device_name": "John's Laptop",
  "device_type": "laptop",
  "os": "macOS 14.0",
  "device_location": "Office Building A",
  "ip_address": "192.168.1.100",
  "mac_address": "00:1B:63:84:45:E6",
  "current_user": "john.doe"
}
```

Response:
```json
{
  "deviceid": "device-001",
  "created": true
}
  "mac_address": "00:1B:63:84:45:E6",
  "current_user": "john.doe"
}
```

### Metrics Collection

#### Submit Metrics

```http
POST /api/v1/metrics
Content-Type: application/json

{
  "deviceid": "device-001",
  "cpu_usage": 45.5,
  "cpu_temp": 65.0,
  "memory_total": 17179869184,
  "memory_used": 8589934592,
  "swap_used": 1073741824,
  "disk_total": 500107862016,
  "disk_used": 250053931008,
  "net_bytes_in": 1048576,
  "net_bytes_out": 524288
}
```

Response:
```json
{
  "status": "success",
  "message": "Metrics recorded",
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Metrics Units**:
- `cpu_usage`: Percentage (0-100)
- `cpu_temp`: Celsius
- `memory_*`, `disk_*`: Bytes
- `net_bytes_*`: Bytes per second

### Activity Logging

#### Log Activity

```http
POST /api/v1/activities
Content-Type: application/json

{
  "deviceid": "device-001",
  "activity_type": "app_launch",
  "description": "Opened Visual Studio Code",
  "app": "Visual Studio Code",
  "duration": 3600
}
```

**Activity Types**:
- `app_launch` - Application started
- `file_access` - File opened or modified
- `browser` - Web browser activity
- `network` - Network connection
- `system` - System event

### Alert Management

#### Submit Alert

```http
POST /api/v1/alerts
Content-Type: application/json

{
  "deviceid": "device-001",
  "level": "high",
  "alert_type": "cpu",
  "message": "CPU usage exceeded threshold",
  "value": 95.5,
  "threshold": 90.0
}
```

**Alert Levels**: `low`, `medium`, `high`, `critical`

**Alert Types**: `cpu`, `memory`, `disk`, `network`, `security`

**Note**: Alerts are automatically forwarded to Mentor Backend

### Screenshot Management

#### Upload Screenshot

```http
POST /api/v1/screenshots
Content-Type: multipart/form-data

device_id: device-001
resolution: 1920x1080
file: [binary data]
```

Response:
```json
{
  "id": 123,
  "deviceid": "device-001",
  "timestamp": "2024-11-04T10:30:00Z",
  "path": "screenshots/device-001/2024-11-04-10-30-00.png",
  "resolution": "1920x1080",
  "size": 2457600
}
```

### Process Tracking

#### Update Process List

```http
POST /api/v1/processes
Content-Type: application/json

{
  "deviceid": "device-001",
  "processes": [
    {
      "pid": 1234,
      "process_name": "chrome",
      "cpu": 15.5,
      "memory": 524288000,
      "command_text": "/Applications/Chrome.app/Contents/MacOS/Chrome"
    },
    {
      "pid": 5678,
      "process_name": "vscode",
      "cpu": 8.2,
      "memory": 312458240,
      "command_text": "/usr/local/bin/code"
    }
  ]
}
```

### Health Check

```http
GET /health
```

Response:
```json
{
  "status": "ok",
  "service": "devices-backend",
  "version": "1.0.0",
  "timestamp": "2024-11-04T10:30:00Z"
}
```

## Mentor Backend API

### Device Management

#### List Devices

```http
GET /devices
```

Response:
```json
[
  {
    "deviceid": "device-001",
    "device_name": "John's Laptop",
    "device_type": "laptop",
    "os": "macOS 14.0",
    "last_seen": "2024-11-04T10:30:00Z",
    "is_online": true,
    "location": "Office Building A",
    "ip_address": "192.168.1.100",
    "current_user": "john.doe"
  },
  {
    "deviceid": "device-002",
    "device_name": "Server 1",
    "device_type": "server",
    "os": "Ubuntu 22.04",
    "last_seen": "2024-11-04T10:29:45Z",
    "is_online": true
  }
]
```

#### Get Device Metrics

```http
GET /devices/device-001/metrics?limit=100
```

Query Parameters:
- `limit`: Number of records (default: 100)
- `since`: ISO timestamp for filtering

Response:
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "deviceid": "device-001",
    "timestamp": "2024-11-04T10:30:00Z",
    "cpu_usage": 45.5,
    "cpu_temp": 65.0,
    "memory_total": 17179869184,
    "memory_used": 8589934592,
    "disk_total": 500107862016,
    "disk_used": 250053931008
  }
]
```

#### Get Device Alerts

```http
GET /devices/device-001/alerts?level=high
```

Query Parameters:
- `level`: Filter by severity (`low`, `medium`, `high`, `critical`)
- `type`: Filter by type (`cpu`, `memory`, `disk`, etc.)
- `limit`: Number of records

Response:
```json
[
  {
    "id": 1,
    "deviceid": "device-001",
    "timestamp": "2024-11-04T10:25:00Z",
    "level": "high",
    "alert_type": "cpu",
    "message": "CPU usage exceeded threshold",
    "value": 95.5,
    "threshold": 90.0
  }
]
```

#### Get Device Activities

```http
GET /devices/device-001/activities?limit=50
```

Response:
```json
[
  {
    "id": 1,
    "deviceid": "device-001",
    "timestamp": "2024-11-04T10:20:00Z",
    "activity_type": "app_launch",
    "description": "Opened Visual Studio Code",
    "app": "Visual Studio Code",
    "duration": 3600
  }
]
```

#### Get Device Screenshots

```http
GET /devices/device-001/screenshots?limit=20
```

Response:
```json
[
  {
    "id": 123,
    "deviceid": "device-001",
    "timestamp": "2024-11-04T10:30:00Z",
    "url": "https://minio.example.com/screenshots/device-001/2024-11-04-10-30-00.png?signature=...",
    "resolution": "1920x1080",
    "size": 2457600
  }
]
```

**Note**: URLs are presigned and expire after a set time (typically 1 hour)

#### Get Device Processes

```http
GET /devices/device-001/processes
```

Response:
```json
[
  {
    "id": 1,
    "deviceid": "device-001",
    "timestamp": "2024-11-04T10:30:00Z",
    "pid": 1234,
    "process_name": "chrome",
    "cpu": 15.5,
    "memory": 524288000,
    "command_text": "/Applications/Chrome.app/Contents/MacOS/Chrome"
  }
]
```

### Remote Commands

#### Create Remote Command

```http
POST /devices/commands
Content-Type: application/json

{
  "deviceid": "device-001",
  "command_text": "restart"
}
```

Response:
```json
{
  "id": 1,
  "deviceid": "device-001",
  "command_text": "restart",
  "status": "pending",
  "created_at": "2024-11-04T10:30:00Z"
}
```

#### Get Pending Commands (Device Polls)

```http
GET /devices/device-001/commands/pending
```

Response:
```json
[
  {
    "id": 1,
    "deviceid": "device-001",
    "command_text": "restart",
    "status": "pending",
    "created_at": "2024-11-04T10:30:00Z"
  }
]
```

#### Update Command Status (Device Reports)

```http
POST /commands/status
Content-Type: application/json

{
  "id": 1,
  "status": "completed",
  "result": "System restarted successfully",
  "exit_code": 0
}
```

**Command Statuses**: `pending`, `running`, `completed`, `failed`

### Alert Ingestion

#### Report Alert (From Devices Backend)

```http
POST /devices/device-001/alerts
Content-Type: application/json

{
  "level": "high",
  "alert_type": "cpu",
  "message": "CPU usage exceeded threshold",
  "value": 95.5,
  "threshold": 90.0
}
```

**Note**: This endpoint is called by Devices Backend, not directly by clients

### Health Check

```http
GET /health
```

Response:
```json
{
  "status": "ok",
  "service": "mentor-backend"
}
```

## Error Handling

### HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request succeeded |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Invalid request data |
| 401 | Unauthorized | Authentication required |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 422 | Unprocessable Entity | Validation error |
| 500 | Internal Server Error | Server error |
| 503 | Service Unavailable | Service temporarily unavailable |

### Error Response Format

```json
{
  "detail": "Error message",
  "status_code": 400,
  "error_code": "VALIDATION_ERROR",
  "errors": [
    {
      "field": "cpu_usage",
      "message": "Value must be between 0 and 100"
    }
  ]
}
```

### Common Errors

**Invalid Device ID**:
```json
{
  "detail": "Device not found",
  "status_code": 404,
  "error_code": "DEVICE_NOT_FOUND"
}
```

**Validation Error**:
```json
{
  "detail": "Validation error",
  "status_code": 422,
  "errors": [
    {
      "loc": ["body", "cpu_usage"],
      "msg": "field required",
      "type": "value_error.missing"
    }
  ]
}
```

**Database Error**:
```json
{
  "detail": "Database connection failed",
  "status_code": 503,
  "error_code": "DATABASE_ERROR"
}
```

## Rate Limiting

**Current Status**: No rate limiting implemented in MVP

**Recommended for Production**:
- 100 requests per minute per IP
- 1000 requests per hour per device
- Exponential backoff for repeated failures

**Headers** (when implemented):
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1699102800
```

## Examples

### Complete Device Workflow

#### 1. Register Device

```bash
curl -X POST http://localhost:30080/api/v1/devices/register \
  -H "Content-Type: application/json" \
  -d '{
    "deviceid": "laptop-001",
    "device_name": "Development Laptop",
    "device_type": "laptop",
    "os": "macOS 14.0"
  }'
```

#### 2. Submit Metrics

```bash
curl -X POST http://localhost:30080/api/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "deviceid": "laptop-001",
    "cpu_usage": 45.5,
    "memory_total": 17179869184,
    "memory_used": 8589934592,
    "disk_total": 500107862016,
    "disk_used": 250053931008
  }'
```

#### 3. Log Activity

```bash
curl -X POST http://localhost:30080/api/v1/activities \
  -H "Content-Type: application/json" \
  -d '{
    "deviceid": "laptop-001",
    "activity_type": "app_launch",
    "description": "Opened Visual Studio Code",
    "app": "Visual Studio Code"
  }'
```

#### 4. Submit Alert

```bash
curl -X POST http://localhost:30080/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '{
    "deviceid": "laptop-001",
    "level": "high",
    "alert_type": "cpu",
    "message": "CPU usage exceeded 90%",
    "value": 95.5,
    "threshold": 90.0
  }'
```

#### 5. View Device in Dashboard

```bash
# List all devices
curl http://localhost:30090/devices

# Get device metrics
curl http://localhost:30090/devices/laptop-001/metrics?limit=10

# Get device alerts
curl http://localhost:30090/devices/laptop-001/alerts?level=high
```

### Python Example

```python
import requests
import time

BASE_URL = "http://localhost:30080/api/v1"
DEVICE_ID = "python-device-001"

# Register device
response = requests.post(
    f"{BASE_URL}/devices/register",
    json={
        "id": DEVICE_ID,
        "device_name": "Python Test Device",
        "device_type": "server"
    }
)
print(f"Device registered: {response.json()}")

# Submit metrics periodically
while True:
    metrics = {
        "deviceid": DEVICE_ID,
        "cpu_usage": 45.5,
        "memory_used": 8589934592,
        "disk_used": 250053931008
    }
    
    response = requests.post(f"{BASE_URL}/metrics", json=metrics)
    print(f"Metrics submitted: {response.status_code}")
    
    time.sleep(60)  # Every minute
```

### JavaScript Example

```javascript
const BASE_URL = 'http://localhost:30080/api/v1';
const DEVICE_ID = 'js-device-001';

// Register device
async function registerDevice() {
  const response = await fetch(`${BASE_URL}/devices/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      id: DEVICE_ID,
      name: 'JavaScript Test Device',
      type: 'laptop'
    })
  });
  
  const data = await response.json();
  console.log('Device registered:', data);
}

// Submit metrics
async function submitMetrics() {
  const metrics = {
    device_id: DEVICE_ID,
    cpu_usage: 45.5,
    memory_used: 8589934592,
    disk_used: 250053931008
  };
  
  const response = await fetch(`${BASE_URL}/metrics`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(metrics)
  });
  
  console.log('Metrics submitted:', response.status);
}

// Run
registerDevice().then(() => {
  setInterval(submitMetrics, 60000); // Every minute
});
```

### Go Example

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "time"
)

const (
    baseURL  = "http://localhost:30080/api/v1"
    deviceID = "go-device-001"
)

type Device struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    Type string `json:"type"`
}

type Metrics struct {
    DeviceID   string  `json:"deviceid"`
    CPUUsage   float64 `json:"cpu_usage"`
    MemoryUsed uint64  `json:"memory_used"`
    DiskUsed   uint64  `json:"disk_used"`
}

func registerDevice() error {
    device := Device{
        ID:   deviceID,
        Name: "Go Test Device",
        Type: "server",
    }
    
    data, _ := json.Marshal(device)
    resp, err := http.Post(
        baseURL+"/devices/register",
        "application/json",
        bytes.NewBuffer(data),
    )
    
    return err
}

func submitMetrics() error {
    metrics := Metrics{
        DeviceID:   deviceID,
        CPUUsage:   45.5,
        MemoryUsed: 8589934592,
        DiskUsed:   250053931008,
    }
    
    data, _ := json.Marshal(metrics)
    _, err := http.Post(
        baseURL+"/metrics",
        "application/json",
        bytes.NewBuffer(data),
    )
    
    return err
}

func main() {
    registerDevice()
    
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        submitMetrics()
    }
}
```

## Interactive API Documentation

### Swagger UI (Devices Backend)

Access interactive API documentation for the Devices Backend:

**URL**: `http://localhost:30080/docs`

Features:
- Try out API calls directly in the browser
- View request/response schemas
- Download OpenAPI specification

### ReDoc (Devices Backend)

Alternative documentation interface for the Devices Backend:

**URL**: `http://localhost:30080/redoc`

Features:
- Clean, readable documentation
- Better for reference and sharing
- Printable format

### Swagger UI (Mentor Backend)

Access interactive API documentation for the Mentor Backend:

**URL**: `http://localhost:30090/docs` or `http://localhost:30090/swagger/index.html`

Features:
- Try out API calls directly in the browser
- View request/response schemas for device management endpoints
- Download OpenAPI specification

### OpenAPI Specifications

Download the OpenAPI specifications directly:
- **Devices Backend**: `http://localhost:30080/openapi.json`
- **Mentor Backend**: `http://localhost:30090/swagger/doc.json`

## Best Practices

1. **Always include device_id** in telemetry submissions
2. **Use batch requests** when submitting multiple metrics
3. **Handle errors gracefully** with retries and backoff
4. **Cache presigned URLs** to reduce requests
5. **Submit metrics at regular intervals** (recommended: 60 seconds)
6. **Use appropriate alert levels** (don't overuse "critical")
7. **Include meaningful descriptions** in activities and alerts

## Support

- View complete OpenAPI specs in `docs/` directory
- Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for API issues
- See [DEVELOPMENT.md](DEVELOPMENT.md) for API development
- Open GitHub issues for API bugs or feature requests

## Changelog

### v1.0.0 (MVP Release)
- Initial API release
- No authentication
- Basic CRUD operations
- Alert forwarding
- Screenshot upload
