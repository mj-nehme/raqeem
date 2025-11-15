#!/usr/bin/env python3
"""
End-to-end integration test for alert flow.

Tests the complete pipeline:
1. Device sends alert to devices backend
2. Devices backend stores alert locally
3. Devices backend forwards alert to mentor backend
4. Mentor backend stores alert
5. Alert is retrievable from mentor backend
"""

import sys
import time
import requests
import json
from datetime import datetime

# Configuration
DEVICES_BACKEND_URL = "http://localhost:8081"
MENTOR_BACKEND_URL = "http://localhost:8080"
TEST_DEVICE_ID = f"test-device-{int(time.time())}"

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

def register_device():
    """Register a test device."""
    log(f"Registering device {TEST_DEVICE_ID}...")
    payload = {
        "deviceid": TEST_DEVICE_ID,
        "device_name": f"E2E Test Device {TEST_DEVICE_ID}",
        "device_type": "laptop",
        "os": "macOS",
        "current_user": "e2e-test",
        "device_location": "Test Lab",
        "ip_address": "192.168.1.100",
        "mac_address": "00:11:22:33:44:55"
    }
    
    try:
        response = requests.post(
            f"{DEVICES_BACKEND_URL}/api/v1/devices/register",
            json=payload,
            timeout=5
        )
        response.raise_for_status()
        log(f"✓ Device registered: {response.json()}")
        return True
    except requests.exceptions.RequestException as e:
        log(f"✗ Device registration failed: {e}", "ERROR")
        return False

def send_alert():
    """Send an alert from the device."""
    log("Sending alert...")
    alerts = [{
        "level": "warning",
        "alert_type": "cpu_high",
        "message": "E2E test alert - CPU usage high",
        "value": 95.5,
        "threshold": 80.0
    }]
    
    try:
        response = requests.post(
            f"{DEVICES_BACKEND_URL}/api/v1/devices/{TEST_DEVICE_ID}/alerts",
            json=alerts,
            timeout=5
        )
        response.raise_for_status()
        result = response.json()
        log(f"✓ Alert sent: {result}")
        return result.get("inserted", 0) == 1
    except requests.exceptions.RequestException as e:
        log(f"✗ Alert send failed: {e}", "ERROR")
        return False

def verify_alert_in_mentor():
    """Verify the alert appears in mentor backend."""
    log("Verifying alert in mentor backend...")
    
    # Give forwarding a moment to complete
    time.sleep(2)
    
    try:
        response = requests.get(
            f"{MENTOR_BACKEND_URL}/devices/{TEST_DEVICE_ID}/alerts",
            timeout=5
        )
        response.raise_for_status()
        alerts = response.json()
        
        if not isinstance(alerts, list):
            log(f"✗ Unexpected response format: {alerts}", "ERROR")
            return False
        
        if len(alerts) == 0:
            log("✗ No alerts found in mentor backend", "ERROR")
            return False
        
        # Find our test alert
        test_alert = None
        for alert in alerts:
            if alert.get("message") == "E2E test alert - CPU usage high":
                test_alert = alert
                break
        
        if test_alert:
            log(f"✓ Alert found in mentor backend: {json.dumps(test_alert, indent=2)}")
            
            # Verify alert fields
            checks = [
                (test_alert.get("deviceid") == TEST_DEVICE_ID, "device_id matches"),
                (test_alert.get("level") == "warning", "level is warning"),
                (test_alert.get("type") == "cpu_high", "type is cpu_high"),
                (test_alert.get("value") == 95.5, "value is 95.5"),
                (test_alert.get("threshold") == 80.0, "threshold is 80.0"),
            ]
            
            all_passed = True
            for passed, description in checks:
                if passed:
                    log(f"  ✓ {description}")
                else:
                    log(f"  ✗ {description}", "ERROR")
                    all_passed = False
            
            return all_passed
        else:
            log(f"✗ Test alert not found. Found {len(alerts)} alert(s)", "ERROR")
            return False
            
    except requests.exceptions.RequestException as e:
        log(f"✗ Failed to retrieve alerts: {e}", "ERROR")
        return False

def run_integration_test():
    """Run the complete integration test."""
    log("=" * 60)
    log("Starting E2E Integration Test")
    log("=" * 60)
    
    # Step 1: Wait for services
    if not wait_for_service(DEVICES_BACKEND_URL, "Devices Backend"):
        return False
    if not wait_for_service(MENTOR_BACKEND_URL, "Mentor Backend"):
        return False
    
    # Step 2: Register device
    if not register_device():
        return False
    
    # Step 3: Send alert
    if not send_alert():
        return False
    
    # Step 4: Verify alert in mentor backend
    if not verify_alert_in_mentor():
        return False
    
    log("=" * 60)
    log("✓ All integration tests passed!", "SUCCESS")
    log("=" * 60)
    return True

if __name__ == "__main__":
    success = run_integration_test()
    sys.exit(0 if success else 1)
