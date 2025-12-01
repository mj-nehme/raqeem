# Backend API Summary

This is a quick, high-level table of key backend endpoints for both services. For full schemas, see `docs/devices-openapi.yaml` and `docs/mentor-openapi.yaml`.

## Role Separation

- **Mentor Backend**: Read-focused - retrieves device data for dashboards and creates commands for devices
- **Devices Backend**: Write-focused - receives telemetry data from devices and provides command polling

## Devices Backend (FastAPI)

The devices backend is **write-focused** - it receives data from monitored devices.

| Endpoint | Method | Description |
|---|---|---|
| `/health` | GET | Service health check. |
| `/docs` | GET | Swagger UI. |
| `/` | GET | Redirects to `/docs`. |
| `/devices/register` | POST | Register or update a device. |
| `/devices/{device_id}/metrics` | POST | Submit metrics for a device. |
| `/devices/{device_id}/processes` | POST | Submit running processes. |
| `/devices/{device_id}/activities` | POST | Submit activity logs. |
| `/devices/{device_id}/alerts` | POST | Submit alerts. |
| `/devices/{device_id}/commands/pending` | GET | Get pending commands (device reads). |
| `/devices/{device_id}/commands` | POST | Create command (from mentor). |
| `/devices/commands/{command_id}/result` | POST | Submit command result. |
| `/screenshots` | POST | Upload screenshot (multipart). |

## Mentor Backend (Gin)

The mentor backend is **read-focused** - it serves the monitoring dashboard.

| Endpoint | Method | Description |
|---|---|---|
| `/health` | GET | Service health check. |
| `/docs` | GET | Swagger UI. |
| `/` | GET | Redirects to `/docs`. |
| `/devices` | GET | List registered devices. |
| `/devices/{id}/metrics` | GET | Get recent metrics for a device. |
| `/devices/{id}/processes` | GET | Get process list for a device. |
| `/devices/{id}/activities` | GET | Get activities for a device. |
| `/devices/{id}/alerts` | GET | Get alerts for a device. |
| `/devices/{id}/screenshots` | GET | Get screenshots for a device. |
| `/screenshots/{filename}` | GET | Stream screenshot file. |
| `/devices/commands` | POST | Queue remote command (only write). |
| `/devices/{id}/commands/pending` | GET | Get pending commands. |
| `/devices/{id}/commands` | GET | Get command history. |
| `/activities` | GET | List recent activities. |

## Notes

- Endpoint shapes reflect the separation of concerns between backends.
- Mentor is primarily read-only (GET) with the exception of command creation (POST).
- Devices is primarily write-only (POST) with the exception of command polling (GET).
- For exact request/response formats, consult the OpenAPI specs referenced above.
