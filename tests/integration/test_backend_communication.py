#!/usr/bin/env python3
"""
Integration test: Backend-to-Backend Communication

Tests the communication between Devices Backend and Mentor Backend:
1. Device registration in devices backend
2. Alert submission to devices backend
3. Automatic alert forwarding from devices → mentor
4. Alert retrieval from mentor backend
5. Verification of data consistency across backends
"""

import sys
import time
import requests
from datetime import datetime


# Configuration
DEVICES_BACKEND_URL = "http://localhost:8081"
MENTOR_BACKEND_URL = "http://localhost:8080"
TEST_DEVICE_ID = f"backend-comm-test-{int(time.time())}"


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


def test_device_registration():
    """Register a device in devices backend."""
    log("Registering device in devices backend...")
    
    payload = {
        "deviceid": TEST_DEVICE_ID,
        "device_name": "Backend Communication Test Device",
        "device_type": "server",
        "os": "CentOS 8",
        "current_user": "integration-test",
        "device_location": "Backend Integration Test Suite",
        "ip_address": "10.0.1.100",
        "mac_address": "11:22:33:44:55:66"
    }
    
    response = requests.post(
        f"{DEVICES_BACKEND_URL}/api/v1/devices/register",
        json=payload,
        timeout=5
    )
    response.raise_for_status()
    result = response.json()
    
    log(f"✓ Device registered: {result}", "SUCCESS")
    assert result.get("deviceid") == TEST_DEVICE_ID, f"Unexpected deviceid in response: {result}"


def test_alert_forwarding():
    """Test alert submission to devices backend and automatic forwarding to mentor."""
    log("Testing alert forwarding...")
    
    # Submit multiple alerts with different severity levels
    alerts = [
        {
            "level": "info",
            "alert_type": "system_update",
            "message": "System update available",
            "value": 1.0,
            "threshold": 0.0
        },
        {
            "level": "warning",
            "alert_type": "disk_space",
            "message": "Disk space running low",
            "value": 85.0,
            "threshold": 80.0
        },
        {
            "level": "critical",
            "alert_type": "service_down",
            "message": "Critical service unavailable",
            "value": 0.0,
            "threshold": 1.0
        }
    ]
    
    # Submit alerts to devices backend
    response = requests.post(
        f"{DEVICES_BACKEND_URL}/api/v1/devices/{TEST_DEVICE_ID}/alerts",
        json=alerts,
        timeout=5
    )
    response.raise_for_status()
    result = response.json()
    
    log(f"✓ Alerts submitted to devices backend: {result}", "SUCCESS")
    assert result.get("inserted", 0) == 3, f"Expected 3 alerts to be inserted, got: {result}"


def test_alert_retrieval_from_mentor():
    """Verify that alerts were forwarded to mentor backend."""
    log("Verifying alerts in mentor backend...")
    
    # Wait for forwarding to complete
    time.sleep(2)
    
    response = requests.get(
        f"{MENTOR_BACKEND_URL}/devices/{TEST_DEVICE_ID}/alerts",
        timeout=5
    )
    response.raise_for_status()
    alerts = response.json()
    
    assert isinstance(alerts, list), f"Expected list response, got: {type(alerts)}"
    
    # Check that we have at least our 3 test alerts
    test_messages = [
        "System update available",
        "Disk space running low",
        "Critical service unavailable"
    ]
    
    found_alerts = {msg: False for msg in test_messages}
    
    for alert in alerts:
        msg = alert.get("message")
        if msg in found_alerts:
            found_alerts[msg] = True
            log(f"  ✓ Found alert: {msg} (level: {alert.get('level')}, type: {alert.get('type')})")
    
    missing = [msg for msg, found in found_alerts.items() if not found]
    assert not missing, f"Missing alerts in mentor backend: {missing}"
    
    log(f"✓ All {len(test_messages)} alerts forwarded successfully", "SUCCESS")


def test_data_consistency():
    """Verify data consistency between backends."""
    log("Testing data consistency across backends...")
    
    # Submit a test alert with specific values
    test_alert = {
        "level": "error",
        "alert_type": "consistency_check",
        "message": "Backend consistency test alert",
        "value": 42.42,
        "threshold": 40.0
    }
    
    # Submit to devices backend
    response = requests.post(
        f"{DEVICES_BACKEND_URL}/api/v1/devices/{TEST_DEVICE_ID}/alerts",
        json=[test_alert],
        timeout=5
    )
    response.raise_for_status()
    
    # Wait for forwarding
    time.sleep(2)
    
    # Retrieve from mentor backend
    response = requests.get(
        f"{MENTOR_BACKEND_URL}/devices/{TEST_DEVICE_ID}/alerts",
        timeout=5
    )
    response.raise_for_status()
    alerts = response.json()
    
    # Find our test alert
    found_alert = None
    for alert in alerts:
        if alert.get("message") == "Backend consistency test alert":
            found_alert = alert
            break
    
    assert found_alert is not None, "Consistency test alert not found in mentor backend"
    
    # Verify all fields match
    assert found_alert.get("deviceid") == TEST_DEVICE_ID, f"deviceid mismatch: {found_alert.get('deviceid')}"
    log("  ✓ deviceid matches")
    
    assert found_alert.get("level") == "error", f"level mismatch: {found_alert.get('level')}"
    log("  ✓ level matches")
    
    assert found_alert.get("type") == "consistency_check", f"type mismatch: {found_alert.get('type')}"
    log("  ✓ type matches")
    
    assert found_alert.get("value") == 42.42, f"value mismatch: {found_alert.get('value')}"
    log("  ✓ value matches")
    
    assert found_alert.get("threshold") == 40.0, f"threshold mismatch: {found_alert.get('threshold')}"
    log("  ✓ threshold matches")
    
    log("✓ Data consistency verified", "SUCCESS")


def run_integration_test():
    """Run the complete integration test."""
    log("=" * 70)
    log("Integration Test: Backend-to-Backend Communication")
    log("=" * 70)
    
    # Step 1: Wait for services
    if not wait_for_service(DEVICES_BACKEND_URL, "Devices Backend"):
        return False
    if not wait_for_service(MENTOR_BACKEND_URL, "Mentor Backend"):
        return False
    
    # Step 2: Register device
    if not test_device_registration():
        return False
    
    # Step 3: Test alert forwarding
    if not test_alert_forwarding():
        return False
    
    # Step 4: Verify alerts in mentor backend
    if not test_alert_retrieval_from_mentor():
        return False
    
    # Step 5: Test data consistency
    if not test_data_consistency():
        return False
    
    log("=" * 70)
    log("✓ All Backend-to-Backend communication tests passed!", "SUCCESS")
    log("=" * 70)
    return True


if __name__ == "__main__":
    success = run_integration_test()
    sys.exit(0 if success else 1)
