# Raqeem API Integration Guide

## Overview

This guide demonstrates how to integrate with the Raqeem monitoring platform's two backend APIs:

- **Devices Backend** (FastAPI): High-throughput telemetry ingestion at `/api/v1/`
- **Mentor Backend** (Go): Centralized monitoring dashboard at `/`

## API Documentation

### Access Swagger UI

**Devices Backend:**
- Swagger UI: `http://localhost:30080/docs`
- ReDoc: `http://localhost:30080/redoc`
- OpenAPI JSON: `http://localhost:30080/openapi.json`

**Mentor Backend:**
- Swagger UI: `http://localhost:30090/swagger/index.html` or `http://localhost:30090/docs`
- OpenAPI JSON: `http://localhost:30090/swagger/doc.json`

## Common Integration Patterns

### 1. Device Registration Flow

**Step 1: Register device with Devices Backend**

```bash
curl -X POST http://localhost:30080/api/v1/devices/register \
  -H "Content-Type: application/json" \
  -d '{
    "deviceid": "a843a399-701f-5011-aff3-4b69d8f21b11",
    "device_name": "Office Laptop",
    "device_type": "laptop",
    "os": "Ubuntu 22.04",
    "device_location": "Office Building A",
    "ip_address": "192.168.1.100",
    "mac_address": "00:1B:44:11:3A:B7",
    "current_user": "john.doe"
  }'
```

**Response:**
```json
{
  "deviceid": "a843a399-701f-5011-aff3-4b69d8f21b11",
  "created": true
}
```

The Devices Backend automatically forwards this registration to the Mentor Backend if `MENTOR_API_URL` is configured.

**Step 2: Verify registration in Mentor Backend**

```bash
curl http://localhost:30090/devices
```

### 2. Submit Performance Metrics

**Submit to Devices Backend:**

```bash
curl -X POST http://localhost:30080/api/v1/devices/a843a399-701f-5011-aff3-4b69d8f21b11/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "cpu_usage": 45.5,
    "cpu_temp": 65.0,
    "memory_total": 16000000000,
    "memory_used": 8000000000,
    "swap_used": 100000000,
    "disk_total": 500000000000,
    "disk_used": 250000000000,
    "net_bytes_in": 1048576,
    "net_bytes_out": 524288
  }'
```

**Response:**
```json
{
  "status": "ok"
}
```

**Retrieve metrics from Mentor Backend:**

```bash
curl http://localhost:30090/devices/a843a399-701f-5011-aff3-4b69d8f21b11/metrics?limit=60
```

### 3. Submit Alerts

**Submit to Devices Backend:**

```bash
curl -X POST http://localhost:30080/api/v1/devices/a843a399-701f-5011-aff3-4b69d8f21b11/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "level": "warning",
    "alert_type": "high_cpu",
    "message": "CPU usage exceeded threshold",
    "value": 95.5,
    "threshold": 90.0
  }]'
```

**Response:**
```json
{
  "inserted": 1
}
```

**Retrieve alerts from Mentor Backend:**

```bash
curl http://localhost:30090/devices/a843a399-701f-5011-aff3-4b69d8f21b11/alerts?limit=100
```

### 4. Remote Command Execution

**Step 1: Create command from Mentor Backend**

```bash
curl -X POST http://localhost:30090/devices/commands \
  -H "Content-Type: application/json" \
  -d '{
    "deviceid": "a843a399-701f-5011-aff3-4b69d8f21b11",
    "command_text": "get_info"
  }'
```

**Step 2: Device polls for pending commands**

```bash
curl http://localhost:30080/api/v1/devices/a843a399-701f-5011-aff3-4b69d8f21b11/commands/pending
```

**Step 3: Device submits command result**

```bash
curl -X POST http://localhost:30080/api/v1/devices/commands/123e4567-e89b-12d3-a456-426614174000/result \
  -H "Content-Type: application/json" \
  -d '{
    "status": "completed",
    "result": "Device info: Ubuntu 22.04, 16GB RAM",
    "exit_code": 0
  }'
```

### 5. Log Activities

**Submit to Devices Backend:**

```bash
curl -X POST http://localhost:30080/api/v1/devices/a843a399-701f-5011-aff3-4b69d8f21b11/activities \
  -H "Content-Type: application/json" \
  -d '[{
    "activity_type": "file_access",
    "description": "Opened document.pdf",
    "app": "Adobe Reader",
    "duration": 300
  }]'
```

**Response:**
```json
{
  "inserted": 1
}
```

**Query activities from Mentor Backend:**

```bash
curl http://localhost:30090/activities
```

### 6. Upload Screenshots

**Upload to Devices Backend:**

```bash
curl -X POST http://localhost:30080/api/v1/screenshots/ \
  -F "device_id=a843a399-701f-5011-aff3-4b69d8f21b11" \
  -F "file=@screenshot.png"
```

**Response:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "image_url": "123e4567-e89b-12d3-a456-426614174000.png",
  "status": "success"
}
```

**Retrieve screenshots from Mentor Backend:**

```bash
curl http://localhost:30090/devices/a843a399-701f-5011-aff3-4b69d8f21b11/screenshots?limit=50
```

## API Consistency

### Naming Conventions

Both APIs use consistent field naming:

| Field | Format | Example |
|-------|--------|---------|
| Device ID | `deviceid` | `a843a399-701f-5011-aff3-4b69d8f21b11` |
| Device Name | `device_name` | `"Office Laptop"` |
| Device Location | `device_location` | `"Office Building A"` |
| Activity Type | `activity_type` | `"file_access"` |
| Alert Type | `alert_type` | `"high_cpu"` |
| Process Name | `process_name` | `"chrome"` |
| Command Text | `command_text` | `"get_info"` |

**Legacy Fields:** Both APIs support legacy field names for backward compatibility but will return validation errors if used. Always use the new field names.

### Error Response Format

Both APIs return consistent error responses:

```json
{
  "detail": "Error message describing what went wrong"
}
```

**Common Status Codes:**
- `200`: Success
- `201`: Created (screenshots)
- `400`: Bad request - validation error or invalid data
- `404`: Resource not found
- `422`: Validation error (Devices Backend)
- `500`: Internal server error

### Data Flow Architecture

```
┌─────────────┐
│   Device    │
│  (Client)   │
└──────┬──────┘
       │
       │ 1. Submit telemetry
       ↓
┌──────────────────────┐
│  Devices Backend     │──→ 2. Forward data
│  (FastAPI)           │    (if configured)
│  Port: 30080         │
└──────────────────────┘
                              ┌─────────────────────┐
                              │  Mentor Backend     │←─── 3. Query data
                              │  (Go)               │
                              │  Port: 30090        │
                              └──────────┬──────────┘
                                         │
                                         │ 4. Display
                                         ↓
                                  ┌─────────────┐
                                  │  Frontend   │
                                  │  Dashboard  │
                                  └─────────────┘
```

### Authentication

**Current State:** Neither API requires authentication.

**Future:** Authentication and authorization will be added in upcoming releases. This guide will be updated accordingly.

## Client Generation

Both APIs provide OpenAPI 3.0 specifications that can be used to generate clients:

### Python Client

```bash
# Install openapi-generator
pip install openapi-generator-cli

# Generate Devices Backend client
openapi-generator generate \
  -i http://localhost:30080/openapi.json \
  -g python \
  -o ./clients/devices-python

# Generate Mentor Backend client
openapi-generator generate \
  -i http://localhost:30090/swagger/doc.json \
  -g python \
  -o ./clients/mentor-python
```

### JavaScript/TypeScript Client

```bash
# Generate Devices Backend client
openapi-generator generate \
  -i http://localhost:30080/openapi.json \
  -g typescript-axios \
  -o ./clients/devices-typescript

# Generate Mentor Backend client
openapi-generator generate \
  -i http://localhost:30090/swagger/doc.json \
  -g typescript-axios \
  -o ./clients/mentor-typescript
```

### Go Client

```bash
# Generate client for Go applications
openapi-generator generate \
  -i http://localhost:30080/openapi.json \
  -g go \
  -o ./clients/devices-go
```

## Best Practices

1. **Always use deviceid format**: UUIDs for device identification
2. **Handle errors gracefully**: Check response status codes and handle errors
3. **Use batch operations**: Submit multiple activities/alerts in a single request
4. **Poll efficiently**: Use appropriate intervals for polling pending commands
5. **Respect rate limits**: Implement backoff strategies for high-volume telemetry
6. **Forward errors are non-blocking**: Device operations succeed even if forwarding fails
7. **Use limits wisely**: Query endpoints support limit parameters to control response size

## Troubleshooting

### Connection Issues

**Problem:** Cannot connect to API
**Solution:** Verify services are running:
```bash
curl http://localhost:30080/health
curl http://localhost:30090/health
```

### Data Not Appearing in Mentor Backend

**Problem:** Data submitted to Devices Backend doesn't show in Mentor Backend
**Solution:** 
- Verify `MENTOR_API_URL` environment variable is set on Devices Backend
- Check Devices Backend logs for forwarding errors
- Manually query Devices Backend to verify data was stored

### Validation Errors

**Problem:** Receiving 400/422 errors
**Solution:**
- Check field names match the new naming convention (not legacy)
- Verify UUIDs are valid format
- Ensure required fields are present
- Review API documentation for expected data types

## Support

- GitHub Issues: https://github.com/mj-nehme/raqeem/issues
- API Documentation: Available via Swagger UI on both backends
- License: MIT
