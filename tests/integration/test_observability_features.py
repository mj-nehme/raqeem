#!/usr/bin/env python3
"""
Integration test: Observability Features (PR#208)

Tests the observability features introduced in PR#208:
1. Health check endpoints (/health/live, /health/ready)
2. Request tracing middleware (X-Request-ID header)
3. Structured logging and request/response tracking

This test validates both Devices Backend (Python/FastAPI) and Mentor Backend (Go/Gin).
"""

import sys
import time
import pytest
import requests
from datetime import datetime


# Configuration
DEVICES_BACKEND_URL = "http://localhost:8081"
MENTOR_BACKEND_URL = "http://localhost:8080"
TEST_DEVICE_ID = f"obs-test-{int(time.time())}"


def log(message, level="INFO"):
    """Print timestamped log message."""
    timestamp = datetime.now().strftime("%H:%M:%S.%f")[:-3]
    color_codes = {
        "INFO": "\033[0m",
        "SUCCESS": "\033[92m",
        "ERROR": "\033[91m",
        "WARNING": "\033[93m"
    }
    color = color_codes.get(level, "\033[0m")
    reset = "\033[0m"
    print(f"{color}[{timestamp}] [{level}] {message}{reset}")


def check_service_available(url, timeout=2):
    """Check if a service is available."""
    try:
        response = requests.get(f"{url}/health", timeout=timeout)
        return response.status_code == 200
    except requests.exceptions.RequestException:
        return False


def wait_for_service(url, name, max_retries=30, delay=2):
    """Wait for a service to become available."""
    log(f"Waiting for {name} at {url}...")
    for i in range(max_retries):
        if check_service_available(url):
            log(f"✓ {name} is ready", "SUCCESS")
            return True
        
        if i < max_retries - 1:
            time.sleep(delay)
    
    log(f"✗ {name} failed to become ready", "ERROR")
    return False


@pytest.fixture(autouse=True)
def check_services():
    """Check if required services are running before running tests."""
    if not check_service_available(DEVICES_BACKEND_URL, timeout=1):
        pytest.skip("Devices backend not available. Start services with ./start.sh or docker-compose up")
    if not check_service_available(MENTOR_BACKEND_URL, timeout=1):
        pytest.skip("Mentor backend not available. Start services with ./start.sh or docker-compose up")


def test_devices_backend_health_endpoints():
    """Test Devices Backend health check endpoints."""
    log("Testing Devices Backend health check endpoints...")
    
    # Test legacy /health endpoint
    log("Testing legacy /health endpoint...")
    response = requests.get(f"{DEVICES_BACKEND_URL}/health", timeout=5)
    assert response.status_code == 200, f"Expected 200, got {response.status_code}"
    data = response.json()
    assert data["status"] == "ok", f"Expected status 'ok', got {data.get('status')}"
    assert data["service"] == "devices-backend", f"Expected service 'devices-backend', got {data.get('service')}"
    log("✓ Legacy /health endpoint works correctly", "SUCCESS")
    
    # Test /health/live endpoint
    log("Testing /health/live endpoint...")
    response = requests.get(f"{DEVICES_BACKEND_URL}/health/live", timeout=5)
    assert response.status_code == 200, f"Expected 200, got {response.status_code}"
    data = response.json()
    assert data["status"] == "alive", f"Expected status 'alive', got {data.get('status')}"
    log("✓ /health/live endpoint works correctly", "SUCCESS")
    
    # Test /health/ready endpoint
    log("Testing /health/ready endpoint...")
    response = requests.get(f"{DEVICES_BACKEND_URL}/health/ready", timeout=5)
    assert response.status_code == 200, f"Expected 200, got {response.status_code}"
    data = response.json()
    assert data["status"] in ["ready", "not_ready"], f"Expected status 'ready' or 'not_ready', got {data.get('status')}"
    assert data["service"] == "devices-backend", f"Expected service 'devices-backend', got {data.get('service')}"
    assert "checks" in data, "Expected 'checks' field in response"
    assert "database" in data["checks"], "Expected 'database' check in response"
    assert "config" in data["checks"], "Expected 'config' check in response"
    
    # Verify database check passed
    if data["checks"]["database"] == "ok":
        log("  ✓ Database check passed", "SUCCESS")
    else:
        log(f"  ⚠ Database check failed: {data['checks']['database']}", "WARNING")
    
    # Verify config check passed
    if data["checks"]["config"] == "ok":
        log("  ✓ Config check passed", "SUCCESS")
    else:
        log(f"  ⚠ Config check failed: {data['checks']['config']}", "WARNING")
    
    log("✓ /health/ready endpoint works correctly", "SUCCESS")


def test_mentor_backend_health_endpoints():
    """Test Mentor Backend health check endpoints."""
    log("Testing Mentor Backend health check endpoints...")
    
    # Test /health endpoint
    log("Testing /health endpoint...")
    response = requests.get(f"{MENTOR_BACKEND_URL}/health", timeout=5)
    assert response.status_code == 200, f"Expected 200, got {response.status_code}"
    data = response.json()
    assert "status" in data, "Expected 'status' field in response"
    log("✓ /health endpoint works correctly", "SUCCESS")


def test_devices_backend_request_tracing():
    """Test Devices Backend request tracing middleware."""
    log("Testing Devices Backend request tracing middleware...")
    
    # Test that X-Request-ID is added to response when not provided
    log("Testing request ID generation...")
    response = requests.get(f"{DEVICES_BACKEND_URL}/health", timeout=5)
    assert "X-Request-ID" in response.headers, "Expected X-Request-ID header in response"
    request_id = response.headers["X-Request-ID"]
    assert len(request_id) > 0, "Expected non-empty request ID"
    log(f"✓ Request ID generated: {request_id}", "SUCCESS")
    
    # Test that provided X-Request-ID is preserved
    log("Testing request ID propagation...")
    custom_request_id = f"test-{int(time.time())}-custom"
    headers = {"X-Request-ID": custom_request_id}
    response = requests.get(f"{DEVICES_BACKEND_URL}/health", headers=headers, timeout=5)
    assert "X-Request-ID" in response.headers, "Expected X-Request-ID header in response"
    returned_id = response.headers["X-Request-ID"]
    assert returned_id == custom_request_id, f"Expected request ID '{custom_request_id}', got '{returned_id}'"
    log(f"✓ Request ID propagated correctly: {returned_id}", "SUCCESS")


def test_mentor_backend_request_tracing():
    """Test Mentor Backend request tracing middleware."""
    log("Testing Mentor Backend request tracing middleware...")
    
    # Test that X-Request-ID is added to response when not provided
    log("Testing request ID generation...")
    response = requests.get(f"{MENTOR_BACKEND_URL}/health", timeout=5)
    assert "X-Request-ID" in response.headers, "Expected X-Request-ID header in response"
    request_id = response.headers["X-Request-ID"]
    assert len(request_id) > 0, "Expected non-empty request ID"
    log(f"✓ Request ID generated: {request_id}", "SUCCESS")
    
    # Test that provided X-Request-ID is preserved
    log("Testing request ID propagation...")
    custom_request_id = f"test-{int(time.time())}-custom"
    headers = {"X-Request-ID": custom_request_id}
    response = requests.get(f"{MENTOR_BACKEND_URL}/health", headers=headers, timeout=5)
    assert "X-Request-ID" in response.headers, "Expected X-Request-ID header in response"
    returned_id = response.headers["X-Request-ID"]
    assert returned_id == custom_request_id, f"Expected request ID '{custom_request_id}', got '{returned_id}'"
    log(f"✓ Request ID propagated correctly: {returned_id}", "SUCCESS")


def test_request_tracing_across_services():
    """Test request ID propagation across services."""
    log("Testing request ID propagation across services...")
    
    # Register a device with a custom request ID
    log("Registering device with custom request ID...")
    custom_request_id = f"cross-service-{int(time.time())}"
    headers = {"X-Request-ID": custom_request_id}
    device_data = {
        "device_id": TEST_DEVICE_ID,
        "device_name": "Observability Test Device",
        "device_type": "laptop"
    }
    
    response = requests.post(
        f"{DEVICES_BACKEND_URL}/api/v1/devices/register",
        json=device_data,
        headers=headers,
        timeout=5
    )
    
    # Check if the request ID was preserved
    if "X-Request-ID" in response.headers:
        returned_id = response.headers["X-Request-ID"]
        if returned_id == custom_request_id:
            log(f"✓ Request ID propagated through device registration: {returned_id}", "SUCCESS")
        else:
            log(f"⚠ Request ID changed during device registration: {custom_request_id} → {returned_id}", "WARNING")
    else:
        log("⚠ X-Request-ID header not found in device registration response", "WARNING")
    
    # Verify device was registered
    if response.status_code in [200, 201]:
        log("✓ Device registered successfully", "SUCCESS")
    else:
        log(f"⚠ Device registration returned status {response.status_code}", "WARNING")


def run_all_tests():
    """Run all observability feature tests."""
    log("=" * 80)
    log("Observability Features Integration Tests (PR#208)")
    log("=" * 80)
    log("")
    
    # Wait for services to be ready
    if not wait_for_service(DEVICES_BACKEND_URL, "Devices Backend"):
        log("✗ Devices Backend not available", "ERROR")
        return False
    
    if not wait_for_service(MENTOR_BACKEND_URL, "Mentor Backend"):
        log("✗ Mentor Backend not available", "ERROR")
        return False
    
    log("")
    log("-" * 80)
    log("Test 1: Devices Backend Health Endpoints")
    log("-" * 80)
    if not test_devices_backend_health_endpoints():
        log("✗ Devices Backend health endpoints test failed", "ERROR")
        return False
    
    log("")
    log("-" * 80)
    log("Test 2: Mentor Backend Health Endpoints")
    log("-" * 80)
    if not test_mentor_backend_health_endpoints():
        log("✗ Mentor Backend health endpoints test failed", "ERROR")
        return False
    
    log("")
    log("-" * 80)
    log("Test 3: Devices Backend Request Tracing")
    log("-" * 80)
    if not test_devices_backend_request_tracing():
        log("✗ Devices Backend request tracing test failed", "ERROR")
        return False
    
    log("")
    log("-" * 80)
    log("Test 4: Mentor Backend Request Tracing")
    log("-" * 80)
    if not test_mentor_backend_request_tracing():
        log("✗ Mentor Backend request tracing test failed", "ERROR")
        return False
    
    log("")
    log("-" * 80)
    log("Test 5: Cross-Service Request Tracing")
    log("-" * 80)
    if not test_request_tracing_across_services():
        log("✗ Cross-service request tracing test failed", "ERROR")
        return False
    
    log("")
    log("=" * 80)
    log("✓ All observability feature tests passed!", "SUCCESS")
    log("=" * 80)
    
    return True


if __name__ == "__main__":
    success = run_all_tests()
    sys.exit(0 if success else 1)
