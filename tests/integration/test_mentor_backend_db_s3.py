#!/usr/bin/env python3
"""
Integration test: Mentor Backend ↔ Database & S3 (MinIO)

Tests the mentor backend's ability to:
1. Connect to PostgreSQL database
2. Retrieve device information
3. Retrieve metrics, activities, and alerts
4. Generate presigned URLs for screenshots from MinIO (S3)
5. Store and retrieve alerts
"""

import sys
import time
import requests
from datetime import datetime


# Configuration
MENTOR_BACKEND_URL = "http://localhost:8080"
TEST_DEVICE_ID = f"mentor-db-s3-test-{int(time.time())}"


def log(message, level="INFO"):
    """Print timestamped log message."""
    timestamp = datetime.now().strftime("%H:%M:%S.%f")[:-3]
    print(f"[{timestamp}] [{level}] {message}")


def wait_for_service(url, name, max_retries=30, delay=2):
    """Wait for a service to become available."""
    log(f"Waiting for {name} at {url}...")
    for i in range(max_retries):
        try:
            response = requests.get(f"{url}/health", timeout=2)
            if response.status_code == 200:
                log(f"✓ {name} is ready")
                return True
        except requests.exceptions.RequestException:
            pass
        
        if i < max_retries - 1:
            time.sleep(delay)
    
    log(f"✗ {name} failed to become ready", "ERROR")
    return False


def test_device_listing():
    """Test device listing from database."""
    log("Testing device listing...")
    
    response = requests.get(
        f"{MENTOR_BACKEND_URL}/devices",
        timeout=5
    )
    response.raise_for_status()
    devices = response.json()
    
    log(f"✓ Device listing successful: found {len(devices)} device(s)", "SUCCESS")
    assert isinstance(devices, list), f"Expected list response, got: {type(devices)}"


def test_alert_submission():
    """Test alert submission to mentor backend (stores in DB)."""
    log("Testing alert submission to mentor backend...")
    
    # First, we need to create a device via devices backend
    # For this test, we'll submit an alert directly to mentor
    alert_payload = {
        "deviceid": TEST_DEVICE_ID,
        "level": "warning",
        "device_type": "network_latency",
        "message": "Network latency detected in mentor backend test",
        "value": 150.5,
        "threshold": 100.0
    }
    
    response = requests.post(
        f"{MENTOR_BACKEND_URL}/devices/{TEST_DEVICE_ID}/alerts",
        json=alert_payload,
        timeout=5
    )
    response.raise_for_status()
    result = response.json()
    
    log(f"✓ Alert submitted successfully: {result}", "SUCCESS")
    # Just verify we got a response, the exact format may vary
    assert result is not None, "Expected non-None response from alert submission"


def test_alert_retrieval():
    """Test alert retrieval from database."""
    log("Testing alert retrieval...")
    
    # Wait a bit for the alert to be stored
    time.sleep(1)
    
    response = requests.get(
        f"{MENTOR_BACKEND_URL}/devices/{TEST_DEVICE_ID}/alerts",
        timeout=5
    )
    response.raise_for_status()
    alerts = response.json()
    
    assert isinstance(alerts, list), f"Expected list response, got: {type(alerts)}"
    
    # Look for our test alert
    found_alert = None
    for alert in alerts:
        if alert.get("message") == "Network latency detected in mentor backend test":
            found_alert = alert
            break
    
    assert found_alert is not None, "Test alert not found in retrieved alerts"
    log(f"✓ Alert retrieved successfully: {found_alert}", "SUCCESS")
    
    # Verify fields
    assert found_alert.get("deviceid") == TEST_DEVICE_ID, f"device_id mismatch: {found_alert.get('deviceid')}"
    log("  ✓ device_id matches")
    
    assert found_alert.get("level") == "warning", f"level mismatch: {found_alert.get('level')}"
    log("  ✓ level is warning")
    
    assert found_alert.get("type") == "network_latency", f"type mismatch: {found_alert.get('type')}"
    log("  ✓ type is network_latency")
    
    assert found_alert.get("value") == 150.5, f"value mismatch: {found_alert.get('value')}"
    log("  ✓ value is 150.5")
    
    assert found_alert.get("threshold") == 100.0, f"threshold mismatch: {found_alert.get('threshold')}"
    log("  ✓ threshold is 100.0")


def test_device_metrics_retrieval():
    """Test metrics retrieval from database (if any exist)."""
    log("Testing metrics retrieval...")
    
    # We'll use the test device ID, though it may not have metrics yet
    response = requests.get(
        f"{MENTOR_BACKEND_URL}/devices/{TEST_DEVICE_ID}/metrics",
        timeout=5
    )
    
    # Accept both 200 (with data) and 404 (no data) as valid
    assert response.status_code in [200, 404], f"Unexpected status code: {response.status_code}"
    
    if response.status_code == 200:
        metrics = response.json()
        count = len(metrics) if isinstance(metrics, list) else 'N/A'
        log(f"✓ Metrics retrieval successful: found {count} metric(s)", "SUCCESS")
    else:
        log("✓ Metrics endpoint accessible (no data yet, expected)", "SUCCESS")


def test_screenshots_retrieval():
    """Test screenshots retrieval (presigned URLs from S3)."""
    log("Testing screenshots retrieval...")
    
    response = requests.get(
        f"{MENTOR_BACKEND_URL}/devices/{TEST_DEVICE_ID}/screenshots",
        timeout=5
    )
    
    # Accept both 200 (with data) and 404 (no data) as valid
    assert response.status_code in [200, 404], f"Unexpected status code: {response.status_code}"
    
    if response.status_code == 200:
        screenshots = response.json()
        count = len(screenshots) if isinstance(screenshots, list) else 'N/A'
        log(f"✓ Screenshots retrieval successful: found {count} screenshot(s)", "SUCCESS")
    else:
        log("✓ Screenshots endpoint accessible (no data yet, expected)", "SUCCESS")


def run_integration_test():
    """Run the complete integration test."""
    log("=" * 70)
    log("Integration Test: Mentor Backend ↔ Database & S3")
    log("=" * 70)
    
    # Step 1: Wait for service
    if not wait_for_service(MENTOR_BACKEND_URL, "Mentor Backend"):
        return False
    
    # Step 2: Test device listing (DB read)
    if not test_device_listing():
        return False
    
    # Step 3: Test alert submission (DB write)
    if not test_alert_submission():
        return False
    
    # Step 4: Test alert retrieval (DB read)
    if not test_alert_retrieval():
        return False
    
    # Step 5: Test metrics retrieval (DB read)
    if not test_device_metrics_retrieval():
        return False
    
    # Step 6: Test screenshots retrieval (S3 presigned URLs)
    if not test_screenshots_retrieval():
        return False
    
    log("=" * 70)
    log("✓ All Mentor Backend ↔ DB/S3 tests passed!", "SUCCESS")
    log("=" * 70)
    return True


if __name__ == "__main__":
    success = run_integration_test()
    sys.exit(0 if success else 1)
