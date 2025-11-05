---
name: CORS Error - Origin not allowed despite dynamic configuration
about: Frontend requests failing with CORS errors even with dynamic port detection
title: 'Fix CORS configuration: Origin http://localhost:4001 not allowed by Access-Control-Allow-Origin'
labels: bug, cors, backend, frontend
assignees: ''
---

## 🐛 Bug Description
Despite implementing dynamic CORS configuration with environment variables, the devices frontend is still experiencing CORS errors when making API calls to the backend.

## 🔍 Error Details
**Console Errors:**
```
[Error] Origin http://localhost:4001 is not allowed by Access-Control-Allow-Origin. Status code: 500
[Error] Fetch API cannot load http://localhost:30080/api/v1/devices/device-xvztc0x4a/metrics due to access control checks.
[Error] Failed to load resource: Origin http://localhost:4001 is not allowed by Access-Control-Allow-Origin. Status code: 500 (metrics, line 0)
[Error] Origin http://localhost:4001 is not allowed by Access-Control-Allow-Origin. Status code: 500
[Error] Fetch API cannot load http://localhost:30080/api/v1/screenshots/ due to access control checks.
[Error] Failed to load resource: Origin http://localhost:4001 is not allowed by Access-Control-Allow-Origin. Status code: 500 (screenshots, line 0)
[Error] Unhandled Promise Rejection: TypeError: Load failed
```

## 🔧 Current Configuration
**Environment Variables (confirmed in containers):**
- `FRONTEND_ORIGINS=http://localhost:4001,http://localhost:5002` (devices-backend)
- `FRONTEND_ORIGIN=http://localhost:4001,http://localhost:5002` (mentor-backend)

**API Endpoints Affected:**
- `GET /api/v1/devices/{device_id}/metrics`
- `GET /api/v1/screenshots/`
- Other device-related endpoints

## 🧐 Investigation Points

### 1. Status Code 500 Issue
The CORS error comes with status code 500, suggesting the backend is throwing an exception before CORS middleware can handle the response properly.

**Check:**
- Backend logs for 500 errors
- Database connectivity issues
- Missing environment variables
- Exception handling in API endpoints

### 2. CORS Middleware Order
**Verify:**
- CORS middleware is applied before route handlers
- No conflicting middleware interfering with CORS headers
- Preflight OPTIONS requests are handled correctly

### 3. Backend CORS Implementation
**Files to check:**
- `devices/backend/src/app/core/cors.py` - CORS setup function
- `devices/backend/src/app/main.py` - Middleware registration order
- Container environment variable parsing

### 4. Request Analysis
**Test:**
- Direct curl requests with Origin header
- Preflight OPTIONS requests
- Different endpoint paths

## 🎯 Expected Behavior
- Frontend at `http://localhost:4001` should successfully make API calls to `http://localhost:30080/api/v1/*`
- CORS headers should be properly returned
- No CORS-related errors in browser console

## 🔍 Debugging Steps

### Step 1: Check Backend Logs
```bash
kubectl logs deployment/devices-backend -f
```

### Step 2: Test CORS with curl
```bash
# Test preflight request
curl -X OPTIONS http://localhost:30080/api/v1/devices/register \
  -H "Origin: http://localhost:4001" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -v

# Test actual request
curl -X GET http://localhost:30080/api/v1/devices/device-test/metrics \
  -H "Origin: http://localhost:4001" \
  -v
```

### Step 3: Verify Environment Variables
```bash
kubectl exec deployment/devices-backend -- env | grep FRONTEND
```

### Step 4: Check Database Connectivity
```bash
kubectl exec deployment/devices-backend -- python -c "
import asyncio
from app.database.connection import get_database_url
print('Database URL:', get_database_url())
"
```

## 📋 Acceptance Criteria
- [ ] Frontend can successfully register devices
- [ ] Frontend can fetch device metrics without CORS errors
- [ ] Frontend can upload/fetch screenshots
- [ ] All API endpoints return proper CORS headers
- [ ] No 500 status codes related to CORS issues
- [ ] Dynamic port detection works for any frontend port assignment

## 🔧 Potential Solutions

### Option 1: Fix 500 Errors First
Address the underlying 500 errors that prevent CORS middleware from responding properly.

### Option 2: CORS Middleware Enhancement
Ensure CORS middleware handles errors gracefully and always returns appropriate headers.

### Option 3: Environment Variable Debugging
Verify the comma-separated origins are being parsed correctly by the FastAPI CORS middleware.

### Option 4: Preflight Request Handling
Ensure OPTIONS requests are handled correctly for all endpoints.

## 📝 Additional Context
- This issue occurred after implementing dynamic CORS configuration
- The backend containers show correct environment variables
- Issue affects multiple API endpoints
- Frontend works when CORS is disabled (dev mode)

## 🏷️ Priority
**High** - Blocks frontend functionality and user workflow