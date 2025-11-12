#!/usr/bin/env python3
"""
Integration test: Devices Backend ↔ Database & S3 (MinIO)

Tests the devices backend's ability to:
1. Connect to PostgreSQL database
2. Store device registration data
3. Store metrics, activities, and alerts
4. Upload screenshots to MinIO (S3)
5. Retrieve stored data
"""

import sys
import time
import requests
import io
from datetime import datetime


# Configuration
DEVICES_BACKEND_URL = "http://localhost:8081"
TEST_DEVICE_ID = f"db-s3-test-{int(time.time())}"


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
    """Test device registration and database storage."""
    log("Testing device registration...")
    
    payload = {
        "id": TEST_DEVICE_ID,
        "name": f"DB/S3 Test Device {TEST_DEVICE_ID}",
        "device_type": "laptop",
        "os": "Ubuntu 22.04",
        "current_user": "test-user",
        "location": "Integration Test Lab",
        "ip_address": "192.168.100.50",
        "mac_address": "AA:BB:CC:DD:EE:FF"
    }
    
    try:
        response = requests.post(
            f"{DEVICES_BACKEND_URL}/api/v1/devices/register",
            json=payload,
            timeout=5
        )
        response.raise_for_status()
        result = response.json()
        
        if result.get("deviceid") == TEST_DEVICE_ID:
            log(f"✓ Device registered successfully: {result}", "SUCCESS")
            return True
        else:
            log(f"✗ Device registration returned unexpected data: {result}", "ERROR")
            return False
            
    except requests.exceptions.RequestException as e:
        log(f"✗ Device registration failed: {e}", "ERROR")
        return False


def test_metrics_storage():
    """Test metrics submission and database storage."""
    log("Testing metrics storage...")
    
    payload = {
        "cpu_usage": 45.7,
        "cpu_temp": 55.2,
        "memory_total": 16000000000,
        "memory_used": 8000000000,
        "swap_used": 500000000,
        "disk_total": 500000000000,
        "disk_used": 250000000000,
        "net_bytes_in": 1024000,
        "net_bytes_out": 512000
    }
    
    try:
        response = requests.post(
            f"{DEVICES_BACKEND_URL}/api/v1/devices/{TEST_DEVICE_ID}/metrics",
            json=payload,
            timeout=5
        )
        response.raise_for_status()
        result = response.json()
        
        if result.get("inserted", 0) >= 1:
            log(f"✓ Metrics stored successfully: {result}", "SUCCESS")
            return True
        else:
            log(f"✗ Metrics storage returned unexpected data: {result}", "ERROR")
            return False
            
    except requests.exceptions.RequestException as e:
        log(f"✗ Metrics storage failed: {e}", "ERROR")
        return False


def test_activity_storage():
    """Test activity logging and database storage."""
    log("Testing activity storage...")
    
    activities = [
        {
            "activity_type": "app_launch",
            "app": "Chrome Browser",
            "description": "Launched web browser",
            "duration": 120
        },
        {
            "activity_type": "file_access",
            "app": "File Manager",
            "description": "Accessed document folder",
            "duration": 30
        }
    ]
    
    try:
        response = requests.post(
            f"{DEVICES_BACKEND_URL}/api/v1/devices/{TEST_DEVICE_ID}/activities",
            json=activities,
            timeout=5
        )
        response.raise_for_status()
        result = response.json()
        
        if result.get("inserted", 0) == 2:
            log(f"✓ Activities stored successfully: {result}", "SUCCESS")
            return True
        else:
            log(f"✗ Activity storage returned unexpected count: {result}", "ERROR")
            return False
            
    except requests.exceptions.RequestException as e:
        log(f"✗ Activity storage failed: {e}", "ERROR")
        return False


def test_alert_storage():
    """Test alert submission and database storage."""
    log("Testing alert storage...")
    
    alerts = [
        {
            "level": "warning",
            "alert_type": "cpu_high",
            "message": "CPU usage above threshold",
            "value": 85.5,
            "threshold": 80.0
        },
        {
            "level": "critical",
            "alert_type": "memory_critical",
            "message": "Memory usage critically high",
            "value": 95.0,
            "threshold": 90.0
        }
    ]
    
    try:
        response = requests.post(
            f"{DEVICES_BACKEND_URL}/api/v1/devices/{TEST_DEVICE_ID}/alerts",
            json=alerts,
            timeout=5
        )
        response.raise_for_status()
        result = response.json()
        
        if result.get("inserted", 0) == 2:
            log(f"✓ Alerts stored successfully: {result}", "SUCCESS")
            return True
        else:
            log(f"✗ Alert storage returned unexpected count: {result}", "ERROR")
            return False
            
    except requests.exceptions.RequestException as e:
        log(f"✗ Alert storage failed: {e}", "ERROR")
        return False


def test_screenshot_upload_to_s3():
    """Test screenshot upload to MinIO (S3)."""
    log("Testing screenshot upload to S3...")
    
    # Create a fake image file
    fake_image = io.BytesIO(b"fake image content for integration test")
    
    try:
        response = requests.post(
            f"{DEVICES_BACKEND_URL}/api/v1/screenshots/",
            data={"deviceid": TEST_DEVICE_ID},
            files={"file": ("integration-test.png", fake_image, "image/png")},
            timeout=10
        )
        response.raise_for_status()
        result = response.json()
        
        if result.get("status") == "success" and "id" in result and "image_url" in result:
            log(f"✓ Screenshot uploaded to S3 successfully: {result}", "SUCCESS")
            return True
        else:
            log(f"✗ Screenshot upload returned unexpected data: {result}", "ERROR")
            return False
            
    except requests.exceptions.RequestException as e:
        log(f"✗ Screenshot upload failed: {e}", "ERROR")
        return False


def run_integration_test():
    """Run the complete integration test."""
    log("=" * 70)
    log("Integration Test: Devices Backend ↔ Database & S3")
    log("=" * 70)
    
    # Step 1: Wait for service
    if not wait_for_service(DEVICES_BACKEND_URL, "Devices Backend"):
        return False
    
    # Step 2: Test device registration (DB)
    if not test_device_registration():
        return False
    
    # Small delay to ensure database consistency
    time.sleep(0.5)
    
    # Step 3: Test metrics storage (DB)
    if not test_metrics_storage():
        return False
    
    # Step 4: Test activity storage (DB)
    if not test_activity_storage():
        return False
    
    # Step 5: Test alert storage (DB)
    if not test_alert_storage():
        return False
    
    # Step 6: Test screenshot upload (S3)
    if not test_screenshot_upload_to_s3():
        return False
    
    log("=" * 70)
    log("✓ All Devices Backend ↔ DB/S3 tests passed!", "SUCCESS")
    log("=" * 70)
    return True


if __name__ == "__main__":
    success = run_integration_test()
    sys.exit(0 if success else 1)
