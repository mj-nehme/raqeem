# ✅ CORS Issue Tracking & Resolutions

## Current Symptom
Browser reported CORS error + 500 response when calling the devices registration endpoint.

## Actual Underlying Causes Encountered
### 1. Historical Device Metrics ID Type Mismatch (Earlier Investigation)
```
sqlalchemy.exc.ProgrammingError: column "id" is of type bigint but expression is of type uuid
[SQL: INSERT INTO device_metrics (...)]
```
Action: Adjust schema or model types (tracked separately).

### 2. Invalid Device Identifier During Registration (Latest)
```
asyncpg.exceptions.DataError: invalid input for query argument $1: 'device-b6ow609xr'
invalid UUID 'device-b6ow609xr': length must be between 32..36 characters
```
Cause: Incoming `deviceid` sometimes not a valid UUID and was bound directly to a UUID column.
Resolution: `register_device` now requires only `deviceid` field; if not a valid UUID it generates one and returns mapping (`original_id`, `normalized: true`). Support for alternate `id` field removed for consistency.

## Why It Looked Like CORS
The 500 error short-circuited normal response flow. Although CORS middleware is in place, error responses thrown before the main handler completion may omit expected `Access-Control-Allow-*` headers, leading browsers to surface a CORS message even though the core issue is application logic.

## Dynamic CORS Configuration (Current Design)
Precedence order:
1. Explicit list via `FRONTEND_ORIGINS` (comma-separated)
2. Regex via `FRONTEND_ORIGIN_REGEX` (single pattern) if list not provided
3. Derived ports: `DEVICES_FRONTEND_PORT`, `MENTOR_FRONTEND_PORT` (fallback defaults, e.g. 4001) forming `http://localhost:<port>` origins
4. Final minimal fallback single origin (`http://localhost:4001`) if nothing else is set

Behavior:
- If a list is supplied it is used directly (`allow_origins`).
- If only regex is supplied it is compiled and used (`allow_origin_regex`).
- Credentials remain disabled; wide wildcard not used.
- Safe failure: unexpected errors in forwarding to mentor backend do not affect CORS.

## Recent Code Changes
File: `devices/backend/src/app/api/v1/endpoints/devices.py`
- Added UUID normalization logic in `/register` using strictly the `deviceid` field.
- Removed fallback to generic `id` to enforce consistent payload schema.
- Forwarded registration now uses normalized UUID.
File: `devices/backend/src/app/core/cors.py`
- Refactored to implement precedence above and environment-driven port derivation.

## Next Recommended Checks
- Confirm frontend now receives 200 with normalized UUID.
- Decide whether to persist original short id (currently returned, not stored) in a separate column if traceability is required.
- Align Helm charts: expose `FRONTEND_ORIGIN_REGEX` and minimize hardcoded port lists.
- Add integration test posting a short non-UUID id and asserting normalized response.

## Quick Verification Commands
```
curl -i -X POST \
	-H "Origin: http://localhost:4001" \
	-H "Content-Type: application/json" \
	http://<devices-service-host>/api/v1/devices/register \
	-d '{"id":"device-short123","name":"Test"}'
```
Expect: `200 OK` with JSON containing `deviceid` (UUID), `original_id`, `normalized: true`.

## Status
✅ UUID normalization implemented
✅ Dynamic CORS confirmed via manual curl preflight & POST
⏳ Chart/env alignment & integration test pending

---
*Updated automatically to reflect ongoing CORS and registration debugging.*