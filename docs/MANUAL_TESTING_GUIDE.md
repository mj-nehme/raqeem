# Manual Testing Guide for Dashboard Data Display Fix

## Overview
This guide helps verify that device data (activities, alerts, processes, screenshots) now appears correctly in the mentor dashboard after the fix.

## Prerequisites
1. Both backends running (devices backend on port 8081, mentor backend on port 8080)
2. PostgreSQL database accessible
3. MinIO running for screenshot storage
4. **IMPORTANT**: `MENTOR_API_URL` environment variable set in devices backend

## Setup

### 1. Configure Environment
In `devices/backend/.env`, add:
```bash
MENTOR_API_URL=http://localhost:8080
```

Or for Kubernetes:
```bash
MENTOR_API_URL=http://mentor-backend.default.svc.cluster.local:8080
```

### 2. Start Services
```bash
# Start all services
./start.sh

# Or manually start backends if needed
cd devices/backend && uvicorn app.main:app --reload --port 8081
cd mentor/backend/src && go run main.go
```

## Test Scenarios

### Test 1: Device Registration
```bash
# Register a test device
curl -X POST http://localhost:8081/api/v1/devices/register \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-device-001",
    "name": "Test Laptop",
    "device_type": "laptop",
    "os": "macOS",
    "current_user": "test-user"
  }'

# Verify in mentor backend
curl http://localhost:8080/devices
# Should show the registered device
```

### Test 2: Metrics
```bash
# Send metrics
curl -X POST http://localhost:8081/api/v1/devices/test-device-001/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "cpu_usage": 45.5,
    "memory_total": 16384,
    "memory_used": 8000,
    "disk_total": 500000,
    "disk_used": 250000
  }'

# Check in mentor backend
curl http://localhost:8080/devices/test-device-001/metrics
# Should return the metrics
```

### Test 3: Activities
```bash
# Send activities
curl -X POST http://localhost:8081/api/v1/devices/test-device-001/activities \
  -H "Content-Type: application/json" \
  -d '[{
    "type": "app_usage",
    "app": "Chrome",
    "description": "User browsing web",
    "duration": 120
  }]'

# Check in mentor backend
curl http://localhost:8080/devices/test-device-001/activities
# Should return the activity
```

### Test 4: Alerts
```bash
# Send alert
curl -X POST http://localhost:8081/api/v1/devices/test-device-001/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "level": "warning",
    "type": "cpu_high",
    "message": "CPU usage is high",
    "value": 95.5,
    "threshold": 80.0
  }]'

# Check in mentor backend
curl http://localhost:8080/devices/test-device-001/alerts
# Should return the alert
```

### Test 5: Processes
```bash
# Send process list
curl -X POST http://localhost:8081/api/v1/devices/test-device-001/processes \
  -H "Content-Type: application/json" \
  -d '[
    {"pid": 1234, "name": "chrome", "cpu": 15.2, "memory": 512000},
    {"pid": 5678, "name": "vscode", "cpu": 8.5, "memory": 256000}
  ]'

# Check in mentor backend
curl http://localhost:8080/devices/test-device-001/processes
# Should return the processes
```

### Test 6: Screenshots (Key Fix)
```bash
# Upload screenshot
curl -X POST http://localhost:8081/api/v1/screenshots/ \
  -F "device_id=test-device-001" \
  -F "file=@/path/to/test-image.png"

# Check in mentor backend
curl http://localhost:8080/devices/test-device-001/screenshots
# Should return screenshots with presigned URLs

# Verify the response includes:
# - "url" field with presigned MinIO URL
# - "screenshot_url" field with presigned MinIO URL
# - "path" field with the filename
# - "deviceid" matching "test-device-001"
```

## Dashboard Verification

### 1. Open Mentor Dashboard
Navigate to the mentor frontend (usually http://localhost:5000)

### 2. Select Device
Click on "test-device-001" in the device list

### 3. Verify Tabs
Check each tab displays data without "Load failed" errors:
- ✅ **Overview**: Shows latest metrics (CPU, memory, disk, network)
- ✅ **Processes**: Lists running processes
- ✅ **Activity**: Shows recent activities
- ✅ **Alerts**: Displays alerts with proper severity
- ✅ **Screenshots**: Shows screenshot thumbnails (click to view full size)

### 4. Use Device Simulator
1. Open device simulator (usually http://localhost:4000)
2. Register a device
3. Click "Start Auto-Simulation"
4. Watch mentor dashboard update in real-time
5. All data should appear within 30 seconds

## Expected Results

✅ **Success Indicators:**
- No "Load failed" errors in any tab
- Screenshots show with clickable thumbnails
- Screenshot URLs are presigned MinIO URLs (contain signatures)
- All data appears within 30 seconds of sending
- Real-time updates work when simulator is running

❌ **Failure Indicators:**
- "Load failed" errors → Check MENTOR_API_URL is set
- Empty lists → Check data was sent successfully
- Screenshots don't load → Check MinIO is running and accessible
- No URLs in screenshots → Check presigned URL generation

## Troubleshooting

### Issue: Screenshots not appearing
**Check:**
1. Is MinIO running? `curl http://localhost:9000/minio/health/live`
2. Are presigned URLs being generated? Check response has "url" field
3. Can browser access MinIO? Try opening a presigned URL directly

### Issue: No data in any tab
**Check:**
1. Is MENTOR_API_URL set in devices backend?
2. Can devices backend reach mentor backend? `curl http://localhost:8080/health`
3. Check devices backend logs for forwarding errors

### Issue: Data appears in DB but not dashboard
**Check:**
1. Are both backends using the same database?
2. Check mentor backend logs for errors
3. Verify mentor backend API returns data: `curl http://localhost:8080/devices/{id}/activities`

## Logs to Monitor

### Devices Backend
```bash
# Watch for forwarding logs
tail -f devices/backend/logs/app.log | grep -i "mentor\|forward"
```

### Mentor Backend
```bash
# Watch for ingestion logs
tail -f mentor/backend/logs/app.log | grep -i "devices\|screenshot"
```

## Success Criteria
- [ ] All 6 curl tests return expected data
- [ ] Dashboard displays all sections without errors
- [ ] Screenshots show with proper URLs
- [ ] Device simulator data appears in dashboard
- [ ] Real-time updates work correctly
