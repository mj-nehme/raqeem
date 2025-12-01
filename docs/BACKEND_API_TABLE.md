# Backend API Summary

This is a quick, high-level table of key backend endpoints for both services. For full schemas, see `docs/devices-openapi.yaml` and `docs/mentor-openapi.yaml`.

## Devices Backend (FastAPI)

| Endpoint | Method | Description |
|---|---|---|
| `/health` | GET | Service health check. |
| `/docs` | GET | Swagger UI. |
| `/` | GET | Redirects to `/docs`. |
| `/devices/register` | POST | Register or update a device. |
| `/devices/` | GET | List all devices. |
| `/devices/{device_id}/metrics` | POST | Submit metrics for a device. |
| `/devices/{device_id}/processes` | POST | Submit running processes. |
| `/devices/{device_id}/activities` | POST | Submit activity logs. |
| `/devices/{device_id}/alerts` | POST | Submit alerts. |
| `/screenshots` | POST | Upload screenshot (multipart). |
| `/users/` | POST | Create user (legacy). |

## Mentor Backend (Gin)

| Endpoint | Method | Description |
|---|---|---|
| `/health` | GET | Service health check. |
| `/docs` | GET | Swagger UI. |
| `/` | GET | Redirects to `/docs`. |
| `/devices` | GET | List registered devices. |
| `/devices` | POST | Register/update device (alias). |
| `/devices/register` | POST | Register/update device. |
| `/devices/metrics` | POST | Ingest metrics (bulk). |
| `/devices/{id}/metrics` | GET | Get recent metrics for a device. |
| `/devices/{id}/processes` | POST | Replace process list. |
| `/devices/{id}/activities` | POST | Post activity entries. |
| `/devices/{id}/alerts` | POST | Post alerts. |
| `/devices/commands` | POST | Queue remote command. |
| `/devices/{id}/commands/pending` | GET | Get pending commands. |
| `/commands/status` | POST | Update command status. |
| `/activities` | GET | List recent activities. |

Notes
- Endpoint shapes reflect unified health/docs routes added across services.
- For exact request/response formats, consult the OpenAPI specs referenced above.
