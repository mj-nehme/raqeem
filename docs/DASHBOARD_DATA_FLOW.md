# Dashboard Data Flow Configuration

This document explains how to configure the data forwarding between devices backend and mentor backend so that device data (activities, alerts, processes, screenshots) appears in the mentor dashboard.

## Overview

The Raqeem platform has two backends:
- **Devices Backend** (Python/FastAPI): Receives data from device simulators
- **Mentor Backend** (Go/Gin): Serves the mentor dashboard

For the mentor dashboard to display device data, the devices backend must forward data to the mentor backend.

## Required Configuration

### Devices Backend Environment Variables

Set the `MENTOR_API_URL` environment variable in the devices backend to point to the mentor backend:

```bash
# In devices/backend/.env
MENTOR_API_URL=http://localhost:8080

# In Kubernetes deployment
MENTOR_API_URL=http://mentor-backend.default.svc.cluster.local:8080
```

### Data Flow

With `MENTOR_API_URL` configured, the devices backend will automatically forward:

1. **Device Registration** → `POST /devices/register`
2. **Metrics** → `POST /devices/metrics`
3. **Processes** → `POST /devices/processes`
4. **Activities** → `POST /devices/activity`
5. **Alerts** → `POST /devices/{device_id}/alerts`
6. **Screenshots** → `POST /devices/screenshots`
7. **Command Results** → `POST /commands/status`

## Verification

To verify the data flow is working:

1. Start both backends with `MENTOR_API_URL` configured
2. Use the device simulator to send data
3. Check the mentor dashboard to see if data appears

### Test with curl

```bash
# Register a device through devices backend
curl -X POST http://localhost:8081/api/v1/devices/register \
  -H "Content-Type: application/json" \
  -d '{"id":"test-device","name":"Test Device","type":"laptop","os":"Linux"}'

# Send metrics
curl -X POST http://localhost:8081/api/v1/devices/test-device/metrics \
  -H "Content-Type: application/json" \
  -d '{"cpu_usage":50,"memory_total":16384,"memory_used":8000}'

# Check mentor backend
curl http://localhost:8080/devices
curl http://localhost:8080/devices/test-device/metrics
```

## Troubleshooting

### Dashboard shows "Load failed" errors

**Cause**: Data is not being forwarded from devices backend to mentor backend

**Solution**: 
1. Check that `MENTOR_API_URL` is set in devices backend environment
2. Verify mentor backend is accessible from devices backend
3. Check devices backend logs for forwarding errors

### Screenshots not appearing

**Cause**: Screenshots need to be uploaded to MinIO and presigned URLs generated

**Solution**:
1. Verify MinIO is running and accessible
2. Check MinIO credentials in mentor backend (defaults: minioadmin/minioadmin1234)
3. Upload a test screenshot through the device simulator
4. Check that presigned URLs are being generated in the response

### Activities/Alerts/Processes not showing

**Cause**: Similar to above - data not being forwarded

**Solution**:
1. Ensure `MENTOR_API_URL` is configured
2. Check that mentor backend endpoints are working: `curl http://localhost:8080/health`
3. Verify database is accessible from both backends

## Architecture

```
Device Simulator → Devices Backend → Mentor Backend → Mentor Dashboard
                         |                |
                         v                v
                    PostgreSQL      PostgreSQL
                                         |
                                         v
                                      MinIO (Screenshots)
```

Both backends share the same PostgreSQL database, so data can be queried from either backend. The forwarding ensures that the mentor backend's tables are populated in real-time.
