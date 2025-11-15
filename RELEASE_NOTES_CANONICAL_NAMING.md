# Release Notes: Canonical Field Naming (Breaking Change)

## Version: TBD

### ⚠️ **BREAKING CHANGE: API Field Names Standardized**

All JSON field names across the Raqeem platform now use **canonical lowercase underscore naming**. Legacy field names are no longer accepted and will result in HTTP 400 errors with clear migration messages.

---

## What Changed

### Field Name Mapping

| Context | Legacy Field | Canonical Field | 
|---------|-------------|-----------------|
| Device Registration | `id` | `deviceid` |
| Device Info | `name` | `device_name` |
| Device Info | `location` | `device_location` |
| Process Info | `name` | `process_name` |
| Process Info | `command` | `command_text` |
| Activities | `type` | `activity_type` |
| Alerts | `type` | `alert_type` |
| Commands | `command` | `command_text` |

### Complete Canonical Field List

**IDs:**
- `deviceid`, `metricid`, `processid`, `activityid`, `alertid`, `commandid`, `screenshotid`, `userid`

**Device Fields:**
- `deviceid`, `device_name`, `device_type`, `os`, `last_seen`, `is_online`, `device_location`, `ip_address`, `mac_address`, `current_user`

**Metric Fields:**
- `metricid`, `deviceid`, `timestamp`, `cpu_usage`, `cpu_temp`, `memory_total`, `memory_used`, `swap_used`, `disk_total`, `disk_used`, `net_bytes_in`, `net_bytes_out`

**Process Fields:**
- `processid`, `deviceid`, `timestamp`, `pid`, `process_name`, `cpu`, `memory`, `command_text`

**Activity Fields:**
- `activityid`, `deviceid`, `timestamp`, `activity_type`, `description`, `app`, `duration`

**Alert Fields:**
- `alertid`, `deviceid`, `timestamp`, `level`, `alert_type`, `message`, `value`, `threshold`

**Command Fields:**
- `commandid`, `deviceid`, `command_text`, `status`, `created_at`, `completed_at`, `result`, `exit_code`

**Screenshot Fields:**
- `screenshotid`, `deviceid`, `timestamp`, `path`, `resolution`, `size`

**User Fields:**
- `userid`, `deviceid`, `username`, `created_at`

---

## Migration Guide

### 1. Update Device Registration

**Before:**
```json
{
  "id": "device-001",
  "name": "My Laptop",
  "location": "Office"
}
```

**After:**
```json
{
  "deviceid": "device-001",
  "device_name": "My Laptop",
  "device_location": "Office"
}
```

### 2. Update Process Submissions

**Before:**
```json
{
  "pid": 1234,
  "name": "chrome",
  "command": "/usr/bin/chrome"
}
```

**After:**
```json
{
  "pid": 1234,
  "process_name": "chrome",
  "command_text": "/usr/bin/chrome"
}
```

### 3. Update Activity Logs

**Before:**
```json
{
  "type": "app_launch",
  "description": "Launched Chrome"
}
```

**After:**
```json
{
  "activity_type": "app_launch",
  "description": "Launched Chrome"
}
```

### 4. Update Alert Submissions

**Before:**
```json
{
  "level": "warning",
  "type": "cpu_high",
  "message": "CPU usage exceeded threshold"
}
```

**After:**
```json
{
  "level": "warning",
  "alert_type": "cpu_high",
  "message": "CPU usage exceeded threshold"
}
```

### 5. Update Command Submissions

**Before:**
```json
{
  "command": "restart"
}
```

**After:**
```json
{
  "command_text": "restart"
}
```

---

## Error Messages

When using legacy field names, you will receive a `400 Bad Request` response with a clear migration message:

```json
{
  "detail": "unsupported legacy field: name; use device_name"
}
```

```json
{
  "detail": "unsupported legacy field: id; use deviceid"
}
```

```json
{
  "detail": "unsupported legacy field: type; use activity_type"
}
```

---

## Impact

### What's Affected
- ✅ All API endpoints (Devices Backend & Mentor Backend)
- ✅ Frontend applications (React dashboards)
- ✅ Simulation scripts
- ✅ Integration tests
- ✅ API documentation

### What's NOT Affected
- ❌ Database schema (column names remain unchanged)
- ❌ Internal processing logic
- ❌ Database migrations

---

## Testing Your Migration

### 1. Update Your Code
Follow the migration guide above to update all API calls.

### 2. Test Against Development Environment
```bash
# Test device registration
curl -X POST http://localhost:30080/api/v1/devices/register \
  -H "Content-Type: application/json" \
  -d '{
    "deviceid": "test-001",
    "device_name": "Test Device",
    "device_type": "laptop"
  }'
```

### 3. Run Your Integration Tests
Ensure all tests pass with the new field names.

### 4. Validate Error Handling
Test that your application properly handles 400 errors if any legacy fields are accidentally used.

---

## Timeline

- **Effective Date:** TBD
- **Deprecation Notice:** N/A (immediate breaking change)
- **Support for Legacy Names:** None (removed immediately)

---

## Support

If you encounter issues during migration:

1. Check the error message for the specific field that needs updating
2. Refer to the field mapping table above
3. Review the updated API documentation: `docs/API.md`
4. Run the CI lint check: `./scripts/lint-field-names.sh`

---

## Technical Details

### Changes Made

1. **Devices Backend (FastAPI)**
   - Added strict validation to reject legacy field names
   - Updated all response serialization to use canonical names
   - Updated request payload parsing

2. **Mentor Backend (Gin)**
   - Already using canonical names via GORM JSON tags
   - No changes required

3. **Frontends**
   - Updated all API calls to use canonical field names
   - Updated data model mappings

4. **Tests & Documentation**
   - Updated all test fixtures
   - Updated API documentation
   - Added comprehensive legacy field rejection tests

5. **CI/CD**
   - Added lint check to prevent legacy field usage
   - CI pipeline updated to run field name validation

---

## Rollback

**This is a breaking change with no rollback path.** All clients must be updated to use canonical field names.

If you discover issues after deployment:
1. Fix the integration by updating to canonical field names
2. Deploy the fix immediately
3. There is no backward-compatible version available

---

## Questions?

Contact the development team or file an issue on GitHub with the label `breaking-change-migration`.
