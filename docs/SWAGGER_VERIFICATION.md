# Swagger Documentation Verification Guide

This guide helps verify that the Swagger/OpenAPI documentation is working correctly for both the Devices Backend (FastAPI) and Mentor Backend (Go).

## Prerequisites

- Services must be running via `./start.sh`
- All pods must be in "Running" state
- Port forwarding must be active

## Verification Steps

### 1. Check Services are Running

```bash
# Check pod status
kubectl get pods -n raqeem

# Verify all pods are running
# Expected: devices-backend, mentor-backend, mentor-frontend, devices-frontend
```

### 2. Verify Devices Backend Swagger UI (FastAPI)

#### Access Swagger UI

```bash
# Get the devices backend URL
./scripts/discover.sh list | grep "Devices Backend"

# Or directly access (typical port)
open http://localhost:30080/docs
```

**Expected Results:**
- ✅ Swagger UI loads successfully
- ✅ Title shows: "Raqeem Devices Backend API"
- ✅ Version shows: "1.0.0"
- ✅ Detailed API description is visible
- ✅ Tags visible: "devices", "screenshots", "health"

#### Test an Endpoint

1. Expand the `/health` endpoint
2. Click "Try it out"
3. Click "Execute"
4. **Expected Response**:
   ```json
   {
     "status": "ok",
     "service": "devices-backend"
   }
   ```

#### Verify Documentation Quality

Check that each endpoint has:
- ✅ Clear summary and description
- ✅ Request body schema (for POST/PUT)
- ✅ Response schemas with examples
- ✅ Error response documentation (400, 404, 500)
- ✅ Parameter descriptions

#### Access ReDoc (Alternative Documentation)

```bash
open http://localhost:30080/redoc
```

**Expected Results:**
- ✅ ReDoc UI loads with same API information
- ✅ Clean, readable layout
- ✅ All endpoints organized by tags

#### Download OpenAPI Specification

```bash
# Download the OpenAPI JSON
curl http://localhost:30080/openapi.json > devices-openapi.json

# Verify it's valid JSON
jq '.info.title' devices-openapi.json

# Expected output: "Raqeem Devices Backend API"
```

### 3. Verify Mentor Backend Swagger UI (Go)

#### Access Swagger UI

```bash
# Get the mentor backend URL
./scripts/discover.sh list | grep "Mentor Backend"

# Or directly access (typical port)
open http://localhost:30090/swagger/index.html

# Alternative redirect endpoint
open http://localhost:30090/docs
```

**Expected Results:**
- ✅ Swagger UI loads successfully
- ✅ Title shows: "Raqeem Mentor Backend API"
- ✅ Version shows: "1.0"
- ✅ Contact information visible
- ✅ License: MIT
- ✅ Tags visible: "devices", "activities"

#### Test an Endpoint

1. Expand the `/health` endpoint
2. Click "Try it out"
3. Click "Execute"
4. **Expected Response**:
   ```json
   {
     "status": "ok",
     "service": "mentor-backend"
   }
   ```

#### Verify Documentation Quality

Check key endpoints:

**Device Registration** (`POST /devices/register`):
- ✅ Summary: "Register a device"
- ✅ Request body shows `models.Device` schema
- ✅ Response shows 200 with `models.Device`
- ✅ Error responses: 400, 500

**List Devices** (`GET /devices`):
- ✅ Summary: "List all devices"
- ✅ Response shows array of `models.Device`
- ✅ Description explains functionality

**Get Device Metrics** (`GET /devices/{id}/metrics`):
- ✅ Path parameter: `id` (Device ID)
- ✅ Query parameter: `limit` (default 60)
- ✅ Response shows array of `models.DeviceMetric`

#### Check Model Definitions

1. Scroll down to "Models" or "Definitions" section
2. Verify these models are documented:
   - ✅ `models.Device`
   - ✅ `models.DeviceMetric`
   - ✅ `models.DeviceActivity`
   - ✅ `models.DeviceAlert`
   - ✅ `models.DeviceProcess`
   - ✅ `models.DeviceRemoteCommand`
   - ✅ `models.DeviceScreenshot`

#### Access Swagger JSON

```bash
# The swagger.json is embedded, but you can check the source
cat mentor/backend/src/docs/swagger.json | jq '.info'

# Expected output shows API metadata
```

### 4. Cross-Platform Consistency Verification

#### Compare Naming Conventions

**Devices Backend Fields:**
- `deviceid` (UUID)
- `device_name` (string)
- `device_location` (string)
- `activity_type` (string)
- `alert_type` (string)
- `command_text` (string)

**Mentor Backend Fields:**
- Same field names (canonical lowercase_underscore)
- UUID types properly handled
- Consistent timestamp formats (RFC3339/ISO8601)

#### Verify Error Response Formats

Both APIs should return errors in consistent format:

```json
{
  "error": "Error message"
}
```

Or for FastAPI validation errors:

```json
{
  "detail": "Error message or validation details"
}
```

### 5. Test Client Generation

#### Generate Python Client

```bash
# Install openapi-generator-cli
npm install -g @openapitools/openapi-generator-cli

# Generate Python client
openapi-generator-cli generate \
  -i http://localhost:30080/openapi.json \
  -g python \
  -o ./test-client-python \
  --additional-properties=packageName=raqeem_test

# Check generated files
ls -la ./test-client-python
```

**Expected Results:**
- ✅ Directory created with Python package structure
- ✅ API client classes generated
- ✅ Model classes for all schemas
- ✅ README with usage instructions

#### Test Generated Client (Optional)

```bash
cd ./test-client-python
pip install -e .

# Test in Python
python << EOF
from raqeem_test import ApiClient, Configuration
config = Configuration(host="http://localhost:30080/api/v1")
client = ApiClient(configuration=config)
print("Client created successfully!")
EOF
```

### 6. Validate OpenAPI Schemas

#### Use OpenAPI Validator

```bash
# Install validator
npm install -g @apidevtools/swagger-cli

# Validate Devices Backend
swagger-cli validate http://localhost:30080/openapi.json

# Validate Mentor Backend (from file)
swagger-cli validate mentor/backend/src/docs/swagger.json
```

**Expected Output:**
- ✅ "The API is valid" or no errors
- No warnings about missing required fields
- No structural issues

### 7. Test Documentation Updates

#### Verify Auto-Generation

After making code changes to endpoints:

**For FastAPI (Devices Backend):**
1. Restart the service
2. Refresh http://localhost:30080/docs
3. ✅ Changes appear automatically

**For Go (Mentor Backend):**
1. Run `swag init -g main.go --output docs`
2. Rebuild and restart the service
3. Refresh http://localhost:30090/swagger/index.html
4. ✅ Changes appear after regeneration

### 8. Accessibility Tests

#### Check from Different Browsers

Test Swagger UIs in:
- ✅ Chrome/Edge
- ✅ Firefox
- ✅ Safari

#### Verify Mobile Responsiveness

1. Open DevTools (F12)
2. Toggle device toolbar
3. Test on mobile viewport sizes
4. ✅ UI remains usable

## Troubleshooting

### Swagger UI Not Loading

**Problem**: Swagger UI shows blank page or 404

**Solutions:**
```bash
# Check pod logs
kubectl logs -n raqeem -l app=devices-backend
kubectl logs -n raqeem -l app=mentor-backend

# Verify port forwarding
kubectl get svc -n raqeem

# Restart services
kubectl rollout restart deployment -n raqeem
```

### OpenAPI JSON Not Accessible

**Problem**: Cannot download openapi.json

**For Devices Backend:**
```bash
# Check if app is running
curl http://localhost:30080/health

# Try with verbose output
curl -v http://localhost:30080/openapi.json
```

**For Mentor Backend:**
```bash
# Swagger JSON is embedded, check docs directory
ls -la mentor/backend/src/docs/
```

### Models Not Showing Up

**Problem**: Swagger shows endpoints but no model definitions

**FastAPI**: Models are generated automatically from Pydantic schemas. Check:
```bash
# Verify schemas exist
ls devices/backend/src/app/schemas/
```

**Go**: Models must be used in annotated endpoints. Check:
```bash
# Regenerate Swagger
cd mentor/backend/src
swag init -g main.go --output docs
```

### Client Generation Fails

**Problem**: openapi-generator returns errors

**Solutions:**
```bash
# Update generator
npm update -g @openapitools/openapi-generator-cli

# Try specific generator version
openapi-generator-cli version-manager set 6.6.0

# Use alternative tool
swagger-codegen generate -i <spec-url> -l python -o ./client
```

## Success Criteria

All checks should pass:

- [x] Devices Backend Swagger UI accessible at `/docs`
- [x] Devices Backend ReDoc accessible at `/redoc`
- [x] Devices Backend OpenAPI JSON downloadable
- [x] Mentor Backend Swagger UI accessible at `/swagger/index.html`
- [x] Mentor Backend `/docs` redirects correctly
- [x] All endpoints have documentation (15+ for mentor, 20+ for devices)
- [x] All models/schemas are documented
- [x] Examples are present for complex payloads
- [x] Error responses are documented
- [x] Client generation works for Python, JS, Go
- [x] OpenAPI specs validate without errors
- [x] Field naming is consistent across both APIs

## Report Issues

If any verification step fails:

1. Document the specific failure
2. Include error messages and logs
3. Note the environment (OS, browser, etc.)
4. Create a GitHub issue with details

## Next Steps

After verification:
1. Update this checklist with actual results
2. Take screenshots of Swagger UIs for documentation
3. Add any discovered issues to backlog
4. Update release notes with API documentation improvements
