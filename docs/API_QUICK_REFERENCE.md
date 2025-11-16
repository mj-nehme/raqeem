# Raqeem API Quick Reference

## Devices Backend API (FastAPI)
**Base URL:** `http://localhost:30080/api/v1`  
**Docs:** http://localhost:30080/docs

### Health Check
```
GET  /health                           → Service health status
```

### Device Management
```
POST /devices/register                 → Register or update device
GET  /devices/                         → List all devices
GET  /devices/{device_id}              → Get device details
```

### Metrics
```
POST /devices/{device_id}/metrics      → Submit performance metrics
GET  /devices/{device_id}/metrics      → Get device metrics (limit=60)
```

### Processes
```
POST /devices/{device_id}/processes    → Update process list
GET  /devices/{device_id}/processes    → Get device processes (limit=100)
GET  /devices/processes                → Get all processes across devices
```

### Activities
```
POST /devices/{device_id}/activities   → Log activities
GET  /devices/{device_id}/activities   → Get device activities (limit=100)
GET  /devices/activities               → Get all activities across devices
```

### Alerts
```
POST /devices/{device_id}/alerts       → Submit alerts
GET  /devices/{device_id}/alerts       → Get device alerts (limit=100)
GET  /devices/alerts                   → Get all alerts across devices
```

### Commands
```
POST /devices/{device_id}/commands     → Create command (from mentor)
GET  /devices/{device_id}/commands/pending → Get pending commands
GET  /devices/{device_id}/commands     → Get command history (limit=100)
POST /devices/commands/{command_id}/result → Submit command result
```

### Screenshots
```
POST /screenshots/                     → Upload screenshot (multipart/form-data)
GET  /devices/{device_id}/screenshots  → Get screenshot metadata (limit=50)
```

---

## Mentor Backend API (Go)
**Base URL:** `http://localhost:30081`  
**Docs:** http://localhost:30081/docs

### Health Check
```
GET  /health                           → Service health status
```

### Device Management
```
POST /devices/register                 → Register device (forwarded)
GET  /devices                          → List all devices
GET  /devices/{id}/metrics             → Get device metrics (limit=60)
GET  /devices/{id}/processes           → Get device processes (limit=100)
GET  /devices/{id}/activities          → Get device activities (limit=100)
GET  /devices/{id}/alerts              → Get device alerts (limit=100)
GET  /devices/{id}/screenshots         → Get screenshots with presigned URLs (limit=50)
```

### Data Ingestion (Forwarded from Devices Backend)
```
POST /devices/metrics                  → Ingest metrics
POST /devices/processes                → Ingest process list
POST /devices/activity                 → Ingest activity
POST /devices/{id}/alerts              → Ingest alert
POST /devices/screenshots              → Ingest screenshot metadata
```

### Commands
```
POST /devices/commands                 → Create remote command
GET  /devices/{id}/commands/pending    → Get pending commands
GET  /devices/{id}/commands            → Get command history (limit=100)
POST /commands/status                  → Update command status
```

### Activities
```
GET  /activities                       → List all activities (filterable)
```

---

## Quick Start Examples

### 1. Register a Device
```bash
curl -X POST http://localhost:30080/api/v1/devices/register \
  -H "Content-Type: application/json" \
  -d '{"deviceid":"a843a399-701f-5011-aff3-4b69d8f21b11","device_name":"My Device"}'
```

### 2. Submit Metrics
```bash
curl -X POST http://localhost:30080/api/v1/devices/a843a399-701f-5011-aff3-4b69d8f21b11/metrics \
  -H "Content-Type: application/json" \
  -d '{"cpu_usage":45.5,"memory_used":8000000000}'
```

### 3. View All Devices
```bash
curl http://localhost:30081/devices
```

### 4. Create Command
```bash
curl -X POST http://localhost:30081/devices/commands \
  -H "Content-Type: application/json" \
  -d '{"deviceid":"a843a399-701f-5011-aff3-4b69d8f21b11","command_text":"get_info"}'
```

### 5. Get Device Metrics
```bash
curl http://localhost:30081/devices/a843a399-701f-5011-aff3-4b69d8f21b11/metrics?limit=60
```

---

## Common Field Names

| Purpose | Field Name | Type | Example |
|---------|-----------|------|---------|
| Device ID | `deviceid` | UUID | `"a843a399-701f-5011-aff3-4b69d8f21b11"` |
| Device Name | `device_name` | string | `"Office Laptop"` |
| Device Location | `device_location` | string | `"Building A"` |
| Activity Type | `activity_type` | string | `"file_access"` |
| Alert Type | `alert_type` | string | `"high_cpu"` |
| Process Name | `process_name` | string | `"chrome"` |
| Command Text | `command_text` | string | `"get_info"` |

---

## HTTP Status Codes

- `200` - Success
- `201` - Created (screenshots)
- `400` - Bad request
- `404` - Not found
- `422` - Validation error
- `500` - Server error

---

## Allowed Remote Commands

- `get_info` - Get device information
- `status` - Get device status
- `restart` - Restart device
- `get_processes` - Get running processes
- `get_logs` - Retrieve logs
- `restart_service` - Restart a service
- `screenshot` - Take screenshot

---

## Data Flow

```
Device → Devices Backend (30080) → Mentor Backend (30081) → Dashboard
```

---

For detailed documentation, visit:
- Devices Backend: http://localhost:30080/docs
- Mentor Backend: http://localhost:30081/docs
- Full Guide: docs/API_INTEGRATION_GUIDE.md
