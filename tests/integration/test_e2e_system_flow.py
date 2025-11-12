#!/usr/bin/env python3
"""
Integration test: End-to-End Complete System Flow

Tests the complete system workflow simulating real device operations:
1. Device registration
2. Continuous metrics submission
3. Activity logging
4. Alert generation and forwarding
5. Screenshot capture and upload
6. Data retrieval and verification from mentor backend
7. Multiple device scenarios

This test simulates a real-world monitoring scenario.
"""

import sys
import time
import requests
import io
from datetime import datetime


# Configuration
DEVICES_BACKEND_URL = "http://localhost:8081"
MENTOR_BACKEND_URL = "http://localhost:8080"
TEST_DEVICE_1 = f"e2e-device-1-{int(time.time())}"
TEST_DEVICE_2 = f"e2e-device-2-{int(time.time())}"


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


def register_device(device_id, name, device_type, os_name):
    """Register a device."""
    log(f"Registering {name}...")
    
    # Use abs(hash()) to ensure positive values for IP and MAC generation
    # Generate IP in range 1-254 to avoid network/broadcast addresses
    hash_val = abs(hash(device_id))
    
    payload = {
        "id": device_id,
        "name": name,
        "type": device_type,
        "os": os_name,
        "current_user": "test-user",
        "location": "E2E Test Lab",
        "ip_address": f"192.168.1.{(hash_val % 254) + 1}",
        "mac_address": f"{hash_val % 256:02X}:BB:CC:DD:EE:FF"
    }
    
    try:
        response = requests.post(
            f"{DEVICES_BACKEND_URL}/api/v1/devices/register",
            json=payload,
            timeout=5
        )
        response.raise_for_status()
        log(f"✓ {name} registered", "SUCCESS")
        return True
    except requests.exceptions.RequestException as e:
        log(f"✗ {name} registration failed: {e}", "ERROR")
        return False


def submit_metrics(device_id, cpu, memory_pct):
    """Submit metrics for a device."""
    payload = {
        "cpu_usage": cpu,
        "cpu_temp": 50 + cpu * 0.5,
        "memory_total": 16000000000,
        "memory_used": int(16000000000 * memory_pct / 100),
        "swap_used": 0,
        "disk_total": 500000000000,
        "disk_used": 250000000000,
        "net_bytes_in": 1024000,
        "net_bytes_out": 512000
    }
    
    try:
        response = requests.post(
            f"{DEVICES_BACKEND_URL}/api/v1/devices/{device_id}/metrics",
            json=payload,
            timeout=5
        )
        response.raise_for_status()
        return True
    except requests.exceptions.RequestException:
        return False


def submit_activity(device_id, activity_type, app, description):
    """Submit activity for a device."""
    activities = [{
        "type": activity_type,
        "app": app,
        "description": description,
        "duration": 60
    }]
    
    try:
        response = requests.post(
            f"{DEVICES_BACKEND_URL}/api/v1/devices/{device_id}/activities",
            json=activities,
            timeout=5
        )
        response.raise_for_status()
        return True
    except requests.exceptions.RequestException:
        return False


def submit_alert(device_id, level, alert_type, message, value, threshold):
    """Submit alert for a device."""
    alerts = [{
        "level": level,
        "type": alert_type,
        "message": message,
        "value": value,
        "threshold": threshold
    }]
    
    try:
        response = requests.post(
            f"{DEVICES_BACKEND_URL}/api/v1/devices/{device_id}/alerts",
            json=alerts,
            timeout=5
        )
        response.raise_for_status()
        return True
    except requests.exceptions.RequestException:
        return False


def upload_screenshot(device_id):
    """Upload a screenshot for a device."""
    fake_image = io.BytesIO(b"fake screenshot data for e2e test")
    
    try:
        response = requests.post(
            f"{DEVICES_BACKEND_URL}/api/v1/screenshots/",
            data={"device_id": device_id},
            files={"file": (f"{device_id}-screenshot.png", fake_image, "image/png")},
            timeout=10
        )
        response.raise_for_status()
        return True
    except requests.exceptions.RequestException:
        return False


def verify_in_mentor(device_id, expected_alerts):
    """Verify device data in mentor backend."""
    log(f"Verifying {device_id} in mentor backend...")
    
    time.sleep(2)  # Wait for forwarding
    
    try:
        # Check device listing
        response = requests.get(f"{MENTOR_BACKEND_URL}/devices", timeout=5)
        response.raise_for_status()
        devices = response.json()
        
        device_found = any(d.get("id") == device_id for d in devices)
        if device_found:
            log("  ✓ Device found in device list")
        else:
            log("  ℹ Device not yet in list (may need registration sync)")
        
        # Check alerts
        response = requests.get(
            f"{MENTOR_BACKEND_URL}/devices/{device_id}/alerts",
            timeout=5
        )
        response.raise_for_status()
        alerts = response.json()
        
        found_count = sum(1 for a in alerts if any(msg in a.get("message", "") for msg in expected_alerts))
        
        if found_count >= len(expected_alerts):
            log(f"  ✓ All {len(expected_alerts)} expected alert(s) found")
            return True
        else:
            log(f"  ✗ Only {found_count}/{len(expected_alerts)} alerts found", "ERROR")
            return False
            
    except requests.exceptions.RequestException as e:
        log(f"  ✗ Verification failed: {e}", "ERROR")
        return False


def simulate_device_scenario_1():
    """Scenario 1: Normal device operation with occasional alerts."""
    log("\n--- Scenario 1: Normal Device Operation ---")
    
    if not register_device(TEST_DEVICE_1, "E2E Laptop 1", "laptop", "Windows 11"):
        return False
    
    # Submit normal metrics
    log("Submitting normal metrics...")
    for i in range(3):
        if not submit_metrics(TEST_DEVICE_1, 45.0 + i * 5, 50.0):
            log("✗ Metrics submission failed", "ERROR")
            return False
        time.sleep(0.3)
    log("✓ Metrics submitted", "SUCCESS")
    
    # Submit activities
    log("Submitting activities...")
    activities = [
        ("app_launch", "Chrome", "User opened web browser"),
        ("file_access", "Word", "User edited document"),
    ]
    for act_type, app, desc in activities:
        if not submit_activity(TEST_DEVICE_1, act_type, app, desc):
            log("✗ Activity submission failed", "ERROR")
            return False
    log("✓ Activities submitted", "SUCCESS")
    
    # Submit a warning alert
    log("Submitting warning alert...")
    if not submit_alert(TEST_DEVICE_1, "warning", "cpu_high", 
                       "CPU usage elevated", 75.0, 70.0):
        log("✗ Alert submission failed", "ERROR")
        return False
    log("✓ Alert submitted", "SUCCESS")
    
    # Upload screenshot
    log("Uploading screenshot...")
    if not upload_screenshot(TEST_DEVICE_1):
        log("✗ Screenshot upload failed", "ERROR")
        return False
    log("✓ Screenshot uploaded", "SUCCESS")
    
    # Verify in mentor
    if not verify_in_mentor(TEST_DEVICE_1, ["CPU usage elevated"]):
        return False
    
    log("✓ Scenario 1 completed", "SUCCESS")
    return True


def simulate_device_scenario_2():
    """Scenario 2: Critical device with multiple alerts."""
    log("\n--- Scenario 2: Critical Device Alerts ---")
    
    if not register_device(TEST_DEVICE_2, "E2E Server 1", "server", "Ubuntu 22.04"):
        return False
    
    # Submit high metrics
    log("Submitting critical metrics...")
    if not submit_metrics(TEST_DEVICE_2, 95.0, 90.0):
        log("✗ Metrics submission failed", "ERROR")
        return False
    log("✓ Critical metrics submitted", "SUCCESS")
    
    # Submit multiple critical alerts
    log("Submitting multiple critical alerts...")
    alerts = [
        ("critical", "cpu_critical", "CPU at maximum capacity", 98.0, 90.0),
        ("critical", "memory_critical", "Memory exhausted", 95.0, 90.0),
        ("error", "disk_error", "Disk I/O errors detected", 100.0, 10.0),
    ]
    
    for level, alert_type, message, value, threshold in alerts:
        if not submit_alert(TEST_DEVICE_2, level, alert_type, message, value, threshold):
            log("✗ Alert submission failed", "ERROR")
            return False
    log("✓ All critical alerts submitted", "SUCCESS")
    
    # Verify in mentor
    expected_messages = [
        "CPU at maximum capacity",
        "Memory exhausted",
        "Disk I/O errors detected"
    ]
    if not verify_in_mentor(TEST_DEVICE_2, expected_messages):
        return False
    
    log("✓ Scenario 2 completed", "SUCCESS")
    return True


def verify_cross_device_data():
    """Verify that both devices are visible in mentor backend."""
    log("\n--- Cross-Device Verification ---")
    
    try:
        response = requests.get(f"{MENTOR_BACKEND_URL}/devices", timeout=5)
        response.raise_for_status()
        devices = response.json()
        
        log(f"Total devices in system: {len(devices)}")
        
        # Check that we can see data from both test devices
        test_devices = [TEST_DEVICE_1, TEST_DEVICE_2]
        found = 0
        
        for device_id in test_devices:
            response = requests.get(
                f"{MENTOR_BACKEND_URL}/devices/{device_id}/alerts",
                timeout=5
            )
            if response.status_code == 200:
                alerts = response.json()
                if len(alerts) > 0:
                    log(f"  ✓ {device_id}: {len(alerts)} alert(s)")
                    found += 1
        
        if found == len(test_devices):
            log("✓ Cross-device data verified", "SUCCESS")
            return True
        else:
            log(f"✗ Only {found}/{len(test_devices)} devices verified", "ERROR")
            return False
            
    except requests.exceptions.RequestException as e:
        log(f"✗ Cross-device verification failed: {e}", "ERROR")
        return False


def run_integration_test():
    """Run the complete end-to-end integration test."""
    log("=" * 70)
    log("Integration Test: End-to-End Complete System Flow")
    log("=" * 70)
    
    # Step 1: Wait for services
    if not wait_for_service(DEVICES_BACKEND_URL, "Devices Backend"):
        return False
    if not wait_for_service(MENTOR_BACKEND_URL, "Mentor Backend"):
        return False
    
    # Step 2: Run device scenarios
    if not simulate_device_scenario_1():
        return False
    
    if not simulate_device_scenario_2():
        return False
    
    # Step 3: Verify cross-device data
    if not verify_cross_device_data():
        return False
    
    log("\n" + "=" * 70)
    log("✓ All End-to-End System tests passed!", "SUCCESS")
    log("=" * 70)
    return True


if __name__ == "__main__":
    success = run_integration_test()
    sys.exit(0 if success else 1)
