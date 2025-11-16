# Data Forwarding and Webhook Documentation

## Overview

The Raqeem platform implements a data forwarding mechanism that automatically synchronizes telemetry data from the Devices Backend to the Mentor Backend. This enables centralized monitoring and dashboard visualization while maintaining a distributed, scalable architecture.

## Architecture

```
┌─────────────────┐
│  Device/Client  │
└────────┬────────┘
         │ POST /api/v1/*
         ↓
┌─────────────────────────────────────┐
│      Devices Backend (FastAPI)      │
│  ┌─────────────┐  ┌──────────────┐ │
│  │   Store in  │  │   Forward    │ │
│  │  Local DB   │  │  to Mentor   │ │
│  └─────────────┘  └──────┬───────┘ │
└─────────────────────────┼──────────┘
                          │ HTTP POST
                          ↓
         ┌────────────────────────────┐
         │   Mentor Backend (Go)      │
         │  Store in Central DB       │
         └────────────────────────────┘
```

## How Data Forwarding Works

### 1. Configuration

Data forwarding is enabled by setting the `MENTOR_API_URL` environment variable on the Devices Backend:

```bash
# In devices backend .env file
MENTOR_API_URL=http://mentor-backend-service:8080
```

**Without this variable set**, the Devices Backend operates in standalone mode and does not forward data.

### 2. Forwarded Data Types

The following data types are automatically forwarded:

| Data Type | Devices Endpoint | Mentor Endpoint | Forwarded |
|-----------|------------------|-----------------|-----------|
| Device Registration | `POST /api/v1/devices/register` | `POST /devices/register` | ✅ Yes |
| Metrics | `POST /api/v1/devices/{id}/metrics` | `POST /devices/metrics` | ✅ Yes |
| Activities | `POST /api/v1/devices/{id}/activities` | `POST /devices/activity` | ✅ Yes |
| Processes | `POST /api/v1/devices/{id}/processes` | `POST /devices/processes` | ✅ Yes |
| Alerts | `POST /api/v1/devices/{id}/alerts` | `POST /devices/alerts` | ✅ Yes |
| Screenshots | `POST /api/v1/devices/{id}/screenshots` | `POST /devices/screenshots` | ✅ Yes |
| Commands | `GET /api/v1/devices/{id}/commands/pending` | N/A | ❌ No (Mentor-originated) |

### 3. Forwarding Behavior

#### Fire-and-Forget Pattern

Data forwarding uses a **fire-and-forget** pattern with the following characteristics:

- **Non-blocking**: Forwarding happens asynchronously and does not block the response to the client
- **Failure-tolerant**: If forwarding fails, the data is still stored locally in the Devices Backend
- **Retry logic**: Failed forwards are retried up to 3 times with exponential backoff
- **No rollback**: Local storage succeeds even if forwarding completely fails

#### Success Flow

```python
# Example: Device registration with successful forwarding
1. Client → POST /api/v1/devices/register → Devices Backend
2. Devices Backend validates and stores device in local DB
3. Devices Backend forwards to Mentor Backend (async)
4. Mentor Backend receives and stores device
5. Devices Backend returns success response to client
```

#### Failure Flow

```python
# Example: Device registration with forwarding failure
1. Client → POST /api/v1/devices/register → Devices Backend
2. Devices Backend validates and stores device in local DB
3. Devices Backend attempts to forward to Mentor Backend (async)
4. Forward fails (network error, Mentor down, etc.)
5. Error is logged but not propagated to client
6. Devices Backend returns success response to client
7. Device exists in Devices Backend but not in Mentor Backend
```

### 4. Implementation Details

#### Forwarding Function

The forwarding is implemented in `devices/backend/src/app/util.py`:

```python
async def post_with_retry(
    url: str,
    json: dict,
    max_retries: int = 3,
    timeout: float = 10.0
) -> bool:
    """
    Post data to a URL with retry logic.
    
    Args:
        url: Target URL
        json: JSON payload to send
        max_retries: Maximum number of retry attempts (default: 3)
        timeout: Request timeout in seconds (default: 10.0)
        
    Returns:
        True if successful, False otherwise
    """
    # Implementation with exponential backoff
    # Retry delays: 1s, 2s, 4s
```

#### Example: Metrics Forwarding

From `devices/backend/src/app/api/v1/endpoints/devices.py`:

```python
@router.post("/devices/{device_id}/metrics")
async def submit_metrics(
    device_id: str,
    metrics: DeviceMetricsSubmit,
    db: AsyncSession = Depends(get_db)
):
    # 1. Validate device exists
    device = await get_device_or_404(device_id, db)
    
    # 2. Store metrics in local database
    metric_record = await store_metrics(device_id, metrics, db)
    
    # 3. Forward to Mentor Backend (if configured)
    if settings.MENTOR_API_URL:
        asyncio.create_task(
            post_with_retry(
                f"{settings.MENTOR_API_URL}/devices/metrics",
                json=metric_record.dict()
            )
        )
    
    # 4. Return success immediately
    return {"status": "ok"}
```

## Monitoring and Troubleshooting

### 1. Verify Forwarding Configuration

Check if forwarding is enabled:

```bash
# Check environment variable in devices backend pod
kubectl exec -n raqeem deployment/devices-backend -- env | grep MENTOR_API_URL

# Expected output if enabled:
# MENTOR_API_URL=http://mentor-backend-service:8080
```

### 2. Check Forwarding Logs

Monitor forwarding attempts and failures:

```bash
# View devices backend logs
kubectl logs -n raqeem deployment/devices-backend -f | grep -i "forward\|mentor"

# Look for patterns like:
# INFO: Successfully forwarded metrics to mentor backend
# WARNING: Failed to forward alert to mentor backend (attempt 1/3)
# ERROR: Forward failed after 3 retries: connection timeout
```

### 3. Test Forwarding Manually

Verify end-to-end forwarding:

```bash
# 1. Submit data to Devices Backend
curl -X POST http://localhost:30080/api/v1/devices/register \
  -H "Content-Type: application/json" \
  -d '{
    "deviceid": "test-device-123",
    "device_name": "Test Device"
  }'

# 2. Query Devices Backend (should exist)
curl http://localhost:30080/api/v1/devices/test-device-123

# 3. Query Mentor Backend (should also exist if forwarding works)
curl http://localhost:30081/devices/test-device-123

# If device exists in Devices but not Mentor:
# - Forwarding may be disabled (no MENTOR_API_URL)
# - Forwarding may be failing (check logs)
# - Network connectivity issues between services
```

### 4. Common Issues

#### Issue: Data Not Appearing in Mentor Backend

**Symptoms**: Data submitted to Devices Backend is stored locally but doesn't appear in Mentor Backend.

**Diagnosis**:
```bash
# Check if MENTOR_API_URL is set
kubectl get configmap -n raqeem devices-backend-config -o yaml | grep MENTOR_API_URL

# Check network connectivity
kubectl exec -n raqeem deployment/devices-backend -- \
  curl -v http://mentor-backend-service:8080/health
```

**Solutions**:
- Verify `MENTOR_API_URL` environment variable is set correctly
- Ensure Mentor Backend service is running and healthy
- Check Kubernetes network policies allow traffic between services
- Review devices backend logs for forwarding errors

#### Issue: Forwarding Causing Performance Issues

**Symptoms**: Slow response times on Devices Backend endpoints.

**Diagnosis**:
```bash
# Check if forwarding is blocking
# Look for synchronous forward calls in logs
kubectl logs -n raqeem deployment/devices-backend | grep "blocking\|sync.*forward"
```

**Solutions**:
- Ensure forwarding is async (using `asyncio.create_task`)
- Increase forwarding timeout if network is slow
- Consider implementing a message queue for large volumes

#### Issue: Inconsistent Data Between Backends

**Symptoms**: Some devices/metrics exist in Devices Backend but not in Mentor Backend.

**Diagnosis**:
```bash
# Compare counts
curl http://localhost:30080/api/v1/devices | jq 'length'
curl http://localhost:30081/devices | jq 'length'
```

**Solutions**:
- Implement a reconciliation script to sync missing data
- Check for forwarding failures during specific time periods
- Consider implementing a dead-letter queue for failed forwards

## Advanced Configuration

### 1. Custom Forwarding Timeout

Adjust timeout for slower networks:

```python
# In devices/backend/src/app/core/config.py
class Settings(BaseSettings):
    MENTOR_API_URL: str | None = None
    FORWARD_TIMEOUT: float = 10.0  # seconds
    FORWARD_MAX_RETRIES: int = 3
```

### 2. Selective Forwarding

Disable forwarding for specific data types:

```python
# In devices/backend/src/app/core/config.py
class Settings(BaseSettings):
    FORWARD_METRICS: bool = True
    FORWARD_ACTIVITIES: bool = True
    FORWARD_ALERTS: bool = True
    FORWARD_PROCESSES: bool = False  # Example: disable process forwarding
```

### 3. Forwarding Metrics

Track forwarding success/failure rates:

```python
# Add Prometheus metrics
from prometheus_client import Counter

forward_success = Counter('forward_success_total', 'Total successful forwards', ['data_type'])
forward_failure = Counter('forward_failure_total', 'Total failed forwards', ['data_type'])

# In forwarding code:
if success:
    forward_success.labels(data_type='metrics').inc()
else:
    forward_failure.labels(data_type='metrics').inc()
```

## Webhook Support (Future Enhancement)

Currently, the platform uses a push-based forwarding model. Future versions may include webhook support where Mentor Backend can register webhooks to receive data.

### Planned Webhook Features

- **Webhook Registration**: API endpoints to register/manage webhooks
- **Event Types**: Subscribe to specific event types (device.registered, metric.submitted, etc.)
- **Retry Logic**: Automatic retries for failed webhook deliveries
- **Signature Verification**: HMAC signatures for webhook payload verification
- **Delivery Logs**: Audit trail of webhook deliveries and failures

Example webhook registration (planned):

```bash
# Register a webhook
curl -X POST http://localhost:30080/api/v1/webhooks \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://external-service.com/webhook",
    "events": ["device.registered", "metric.submitted"],
    "secret": "webhook-secret-key"
  }'
```

## Best Practices

1. **Always enable forwarding in production** to ensure centralized monitoring
2. **Monitor forwarding logs** to detect and resolve issues quickly
3. **Use health checks** to verify both backends are operational
4. **Implement reconciliation** to sync missing data periodically
5. **Set appropriate timeouts** based on network latency
6. **Use retry limits** to prevent cascading failures
7. **Test forwarding** in staging before deploying to production

## Related Documentation

- [API Integration Guide](API_INTEGRATION_GUIDE.md)
- [Architecture Overview](ARCHITECTURE.md)
- [Deployment Guide](DEPLOYMENT.md)
- [Troubleshooting Guide](TROUBLESHOOTING.md)

## Support

For issues related to data forwarding:
1. Check logs on both Devices and Mentor backends
2. Verify network connectivity between services
3. Review this documentation for common issues
4. Open an issue at https://github.com/mj-nehme/raqeem/issues
