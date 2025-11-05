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
        "id": TEST_DEVICE_ID,
        "name": f"Backend Communication Test Device",
        "type": "server",
        "os": "CentOS 8",
        "current_user": "integration-test",
        "location": "Backend Integration Test Suite",
        "ip_address": "10.0.1.100",
        "mac_address": "11:22:33:44:55:66"
    }
    
    try:
        response = requests.post(
            f"{DEVICES_BACKEND_URL}/api/v1/devices/register",
            json=payload,
            timeout=5
        )
        response.raise_for_status()
        result = response.json()
        
        if result.get("device_id") == TEST_DEVICE_ID:
            log(f"✓ Device registered: {result}", "SUCCESS")
            return True
        else:
            log(f"✗ Unexpected registration response: {result}", "ERROR")
            return False
            
    except requests.exceptions.RequestException as e:
        log(f"✗ Device registration failed: {e}", "ERROR")
        return False


def test_alert_forwarding():
    """Test alert submission to devices backend and automatic forwarding to mentor."""
    log("Testing alert forwarding...")
    
    # Submit multiple alerts with different severity levels
    alerts = [
        {
            "level": "info",
            "type": "system_update",
            "message": "System update available",
            "value": 1.0,
            "threshold": 0.0
        },
        {
            "level": "warning",
            "type": "disk_space",
            "message": "Disk space running low",
            "value": 85.0,
            "threshold": 80.0
        },
        {
            "level": "critical",
            "type": "service_down",
            "message": "Critical service unavailable",
            "value": 0.0,
            "threshold": 1.0
        }
    ]
    
    try:
        # Submit alerts to devices backend
        response = requests.post(
            f"{DEVICES_BACKEND_URL}/api/v1/devices/{TEST_DEVICE_ID}/alerts",
            json=alerts,
            timeout=5
        )
        response.raise_for_status()
        result = response.json()
        
        if result.get("inserted", 0) == 3:
            log(f"✓ Alerts submitted to devices backend: {result}", "SUCCESS")
            return True
        else:
            log(f"✗ Unexpected alert submission response: {result}", "ERROR")
            return False
            
    except requests.exceptions.RequestException as e:
        log(f"✗ Alert submission failed: {e}", "ERROR")
        return False


def test_alert_retrieval_from_mentor():
    """Verify that alerts were forwarded to mentor backend."""
    log("Verifying alerts in mentor backend...")
    
    # Wait for forwarding to complete
    time.sleep(2)
    
    try:
        response = requests.get(
            f"{MENTOR_BACKEND_URL}/devices/{TEST_DEVICE_ID}/alerts",
            timeout=5
        )
        response.raise_for_status()
        alerts = response.json()
        
        if not isinstance(alerts, list):
            log(f"✗ Unexpected response format: {type(alerts)}", "ERROR")
            return False
        
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
        
        all_found = all(found_alerts.values())
        
        if all_found:
            log(f"✓ All {len(test_messages)} alerts forwarded successfully", "SUCCESS")
            return True
        else:
            missing = [msg for msg, found in found_alerts.items() if not found]
            log(f"✗ Missing alerts: {missing}", "ERROR")
            return False
            
    except requests.exceptions.RequestException as e:
        log(f"✗ Alert retrieval from mentor failed: {e}", "ERROR")
        return False


def test_data_consistency():
    """Verify data consistency between backends."""
    log("Testing data consistency across backends...")
    
    # Submit a test alert with specific values
    test_alert = {
        "level": "error",
        "type": "consistency_check",
        "message": "Backend consistency test alert",
        "value": 42.42,
        "threshold": 40.0
    }
    
    try:
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
        
        if found_alert:
            # Verify all fields match
            checks = [
                (found_alert.get("device_id") == TEST_DEVICE_ID, "device_id"),
                (found_alert.get("level") == "error", "level"),
                (found_alert.get("type") == "consistency_check", "type"),
                (found_alert.get("value") == 42.42, "value"),
                (found_alert.get("threshold") == 40.0, "threshold"),
            ]
            
            all_passed = True
            for passed, field in checks:
                if passed:
                    log(f"  ✓ {field} matches")
                else:
                    log(f"  ✗ {field} mismatch", "ERROR")
                    all_passed = False
            
            if all_passed:
                log("✓ Data consistency verified", "SUCCESS")
                return True
            else:
                return False
        else:
            log("✗ Consistency test alert not found in mentor backend", "ERROR")
            return False
            
    except requests.exceptions.RequestException as e:
        log(f"✗ Data consistency test failed: {e}", "ERROR")
        return False


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
