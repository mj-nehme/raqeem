# OpenAPI Specification Summary

## Overview

This document provides a summary of the OpenAPI specifications for both Raqeem backends.

---

## Devices Backend API (FastAPI)

### Specification Details
- **Format**: OpenAPI 3.1.0
- **Title**: Raqeem Devices Backend API
- **Version**: 1.0.0
- **License**: MIT
- **Base URL**: `/api/v1`

### Access Points
- **Swagger UI**: http://localhost:30080/docs
- **ReDoc**: http://localhost:30080/redoc
- **OpenAPI JSON**: http://localhost:30080/openapi.json

### Statistics
- **Total Paths**: 16
- **Total Tags**: 9
- **Total Schemas**: 15+ (comprehensive Pydantic models)

### Endpoints by Category

#### Device Registration (2 endpoints)
- `POST /devices/register` - Register or update device
- `GET /devices/{device_id}` - Get device details

#### Device Information (2 endpoints)
- `GET /devices/` - List all devices
- `GET /health` - Health check

#### Device Metrics (2 endpoints)
- `POST /devices/{device_id}/metrics` - Submit metrics
- `GET /devices/{device_id}/metrics` - Get device metrics

#### Device Processes (3 endpoints)
- `POST /devices/{device_id}/processes` - Update process list
- `GET /devices/{device_id}/processes` - Get device processes
- `GET /devices/processes` - Get all processes

#### Device Activities (3 endpoints)
- `POST /devices/{device_id}/activities` - Log activities
- `GET /devices/{device_id}/activities` - Get device activities
- `GET /devices/activities` - Get all activities

#### Device Alerts (3 endpoints)
- `POST /devices/{device_id}/alerts` - Submit alerts
- `GET /devices/{device_id}/alerts` - Get device alerts
- `GET /devices/alerts` - Get all alerts

#### Device Commands (4 endpoints)
- `POST /devices/{device_id}/commands` - Create command
- `GET /devices/{device_id}/commands/pending` - Get pending commands
- `GET /devices/{device_id}/commands` - Get command history
- `POST /devices/commands/{command_id}/result` - Submit command result

#### Device Screenshots (2 endpoints)
- `GET /devices/{device_id}/screenshots` - Get screenshot metadata
- `POST /screenshots/` - Upload screenshot

### Request/Response Models

#### Device Models
- `DeviceRegister` - Device registration request
- `DeviceRegisterResponse` - Registration response
- `DeviceInfo` - Device information

#### Metrics Models
- `DeviceMetricsSubmit` - Metrics submission
- `DeviceMetrics` - Metrics response

#### Process Models
- `ProcessSubmit` - Process submission
- `DeviceProcess` - Process response

#### Activity Models
- `ActivitySubmit` - Activity submission
- `DeviceActivity` - Activity response

#### Alert Models
- `AlertSubmit` - Alert submission
- `DeviceAlert` - Alert response

#### Command Models
- `CommandCreate` - Command creation
- `CommandOut` - Command response
- `CommandResultSubmit` - Command result submission

#### Screenshot Models
- `DeviceScreenshot` - Screenshot metadata

#### Standard Responses
- `StatusResponse` - Standard status response
- `InsertedResponse` - Bulk insert response
- `ErrorResponse` - Error response

### Features
- ✅ Comprehensive request/response schemas
- ✅ Detailed descriptions and examples
- ✅ All HTTP status codes documented
- ✅ Query parameters with defaults
- ✅ Tag-based organization
- ✅ Legacy field documentation

---

## Mentor Backend API (Go)

### Specification Details
- **Format**: Swagger 2.0 (OpenAPI 2.0)
- **Title**: Raqeem Mentor Backend API
- **Version**: 1.0
- **License**: MIT
- **Base URL**: `/`

### Access Points
- **Swagger UI**: http://localhost:30081/docs or http://localhost:30081/swagger/index.html
- **OpenAPI JSON**: http://localhost:30081/swagger/doc.json
- **YAML**: http://localhost:30081/swagger/swagger.yaml

### Statistics
- **Total Paths**: 16
- **Total Definitions**: 8
- **Total Tags**: 3

### Endpoints by Category

#### Activities (1 endpoint)
- `GET /activities` - List all activities (with filters)

#### Commands (4 endpoints)
- `POST /commands/status` - Update command status
- `POST /devices/commands` - Create remote command
- `GET /devices/{id}/commands/pending` - Get pending commands
- `GET /devices/{id}/commands` - Get command history

#### Devices (11 endpoints)
- `POST /devices/register` - Register device
- `POST /devices/metrics` - Submit metrics
- `POST /devices/processes` - Update process list
- `POST /devices/activity` - Log activity
- `POST /devices/screenshots` - Store screenshot metadata
- `POST /devices/{id}/alerts` - Report alert
- `GET /devices` - List all devices
- `GET /devices/{id}/metrics` - Get device metrics
- `GET /devices/{id}/processes` - Get device processes
- `GET /devices/{id}/activities` - Get device activities
- `GET /devices/{id}/alerts` - Get device alerts
- `GET /devices/{id}/screenshots` - Get screenshots with presigned URLs

### Data Models
- `models.Device` - Device information
- `models.DeviceMetric` - Device metrics
- `models.DeviceProcess` - Process information
- `models.DeviceActivity` - Activity information
- `models.DeviceAlert` - Alert information
- `models.DeviceRemoteCommand` - Remote command
- `models.DeviceScreenshot` - Screenshot metadata
- `models.User` - User information

### Features
- ✅ Comprehensive Swagger annotations
- ✅ All parameters documented
- ✅ Request/response models defined
- ✅ Error responses documented
- ✅ Enhanced API description
- ✅ Tag-based organization

---

## Cross-Platform Comparison

### Similarities
- Both expose comprehensive Swagger/OpenAPI documentation
- Both use similar endpoint structures
- Both support the same core operations
- Both use consistent field naming
- Both provide detailed examples

### Differences

| Feature | Devices Backend | Mentor Backend |
|---------|----------------|----------------|
| Framework | FastAPI (Python) | Gin (Go) |
| OpenAPI Version | 3.1.0 | 2.0 |
| Base Path | `/api/v1` | `/` |
| Port | 30080 | 30081 |
| Primary Use | Telemetry ingestion | Dashboard/monitoring |
| Data Flow | Receives from devices | Receives forwarded data |

### Naming Convention Consistency

Both APIs use the same field names:

| Field | Format | Example |
|-------|--------|---------|
| Device ID | `deviceid` | `"a843a399-701f-5011-aff3-4b69d8f21b11"` |
| Device Name | `device_name` | `"Office Laptop"` |
| Device Location | `device_location` | `"Building A"` |
| Activity Type | `activity_type` | `"file_access"` |
| Alert Type | `alert_type` | `"high_cpu"` |
| Process Name | `process_name` | `"chrome"` |
| Command Text | `command_text` | `"get_info"` |

---

## Client Generation

Both specifications can be used to generate API clients:

### Using OpenAPI Generator

```bash
# Install
npm install @openapitools/openapi-generator-cli -g

# Generate Python client for Devices Backend
openapi-generator-cli generate \
  -i http://localhost:30080/openapi.json \
  -g python \
  -o ./clients/devices-python

# Generate TypeScript client for Mentor Backend
openapi-generator-cli generate \
  -i http://localhost:30081/swagger/doc.json \
  -g typescript-axios \
  -o ./clients/mentor-typescript
```

### Supported Generators
- Python (requests, urllib3)
- JavaScript/TypeScript (axios, fetch)
- Go
- Java
- C#
- Ruby
- PHP
- And 50+ more languages

---

## Validation

### Specification Validation
Both specifications can be validated using:

```bash
# Install validator
npm install -g @apidevtools/swagger-cli

# Validate Devices Backend
swagger-cli validate http://localhost:30080/openapi.json

# Validate Mentor Backend
swagger-cli validate http://localhost:30081/swagger/doc.json
```

### Testing
Both APIs can be tested directly from Swagger UI:
1. Navigate to the Swagger UI
2. Click "Try it out" on any endpoint
3. Fill in parameters
4. Click "Execute"
5. View response

---

## Documentation Maintenance

### Devices Backend (FastAPI)
- **Automatic**: FastAPI generates OpenAPI from code
- **Update Process**: Update docstrings and Pydantic models
- **No rebuild needed**: Documentation updates automatically

### Mentor Backend (Go)
- **Semi-automatic**: Uses swaggo/swag
- **Update Process**: 
  1. Update Swagger comments in code
  2. Run `swag init` to regenerate
  3. Commit generated files
- **Command**: `swag init --parseDependency --parseInternal`

---

## Best Practices

### Documentation
- Keep descriptions clear and concise
- Provide realistic examples
- Document all parameters
- Include error responses
- Use consistent terminology

### Maintenance
- Update docs when code changes
- Regenerate Go docs after changes
- Test examples in Swagger UI
- Validate specs regularly

### Versioning
- Current version: 1.0.0 (devices), 1.0 (mentor)
- Version in URL path: `/api/v1`
- Maintain backward compatibility
- Document breaking changes

---

## Resources

### Documentation
- [OpenAPI Specification](https://swagger.io/specification/)
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [Swaggo Documentation](https://github.com/swaggo/swag)

### Tools
- [Swagger UI](https://swagger.io/tools/swagger-ui/)
- [ReDoc](https://github.com/Redocly/redoc)
- [OpenAPI Generator](https://openapi-generator.tech/)
- [Swagger CLI](https://github.com/APIDevTools/swagger-cli)

### Raqeem Docs
- API Integration Guide: `docs/API_INTEGRATION_GUIDE.md`
- API Quick Reference: `docs/API_QUICK_REFERENCE.md`
- Development Guide: `docs/DEVELOPMENT.md`
- Testing Guide: `docs/TESTING.md`

---

**Last Updated**: 2025-11-16  
**Status**: Complete ✅  
**Version**: v0.2.0
