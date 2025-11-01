#!/usr/bin/env python3
"""
Quick smoke test for running services.
Tests basic connectivity and health of deployed services.
Use this when services are already running (via ./scripts/start.sh or docker-compose).
"""

import sys
import requests
import time
from datetime import datetime

def log(message, level="INFO"):
    """Print timestamped log message."""
    timestamp = datetime.now().strftime("%H:%M:%S")
    symbols = {"INFO": "ℹ️", "SUCCESS": "✓", "ERROR": "✗", "WARN": "⚠️"}
    symbol = symbols.get(level, "•")
    print(f"[{timestamp}] {symbol} {message}")

def check_service(url, name):
    """Check if a service is responding."""
    try:
        response = requests.get(f"{url}/health", timeout=3)
        if response.status_code == 200:
            data = response.json()
            log(f"{name} is healthy: {data}", "SUCCESS")
            return True
        else:
            log(f"{name} returned status {response.status_code}", "ERROR")
            return False
    except requests.exceptions.RequestException as e:
        log(f"{name} is not reachable: {e}", "ERROR")
        return False

def test_alert_flow(devices_url, mentor_url):
    """Quick test of alert forwarding."""
    device_id = f"smoke-test-{int(time.time())}"
    
    # 1. Register device
    log("Registering test device...")
    try:
        response = requests.post(
            f"{devices_url}/api/v1/devices/register",
            json={
                "id": device_id,
                "name": "Smoke Test Device",
                "type": "laptop",
                "os": "Test OS"
            },
            timeout=5
        )
        response.raise_for_status()
        log("Device registered", "SUCCESS")
    except Exception as e:
        log(f"Device registration failed: {e}", "ERROR")
        return False
    
    # 2. Send alert
    log("Sending test alert...")
    try:
        response = requests.post(
            f"{devices_url}/api/v1/devices/{device_id}/alerts",
            json=[{
                "level": "info",
                "type": "smoke_test",
                "message": "Smoke test alert",
                "value": 50,
                "threshold": 80
            }],
            timeout=5
        )
        response.raise_for_status()
        log("Alert sent", "SUCCESS")
    except Exception as e:
        log(f"Alert send failed: {e}", "ERROR")
        return False
    
    # 3. Check alert in mentor backend (with retry)
    log("Checking alert in mentor backend...")
    time.sleep(2)  # Give forwarding time
    
    for attempt in range(3):
        try:
            response = requests.get(
                f"{mentor_url}/devices/{device_id}/alerts",
                timeout=5
            )
            response.raise_for_status()
            alerts = response.json()
            
            if isinstance(alerts, list) and len(alerts) > 0:
                found = any(a.get("message") == "Smoke test alert" for a in alerts)
                if found:
                    log("Alert found in mentor backend", "SUCCESS")
                    return True
                else:
                    log(f"Alert not found (attempt {attempt + 1}/3)", "WARN")
            else:
                log(f"No alerts yet (attempt {attempt + 1}/3)", "WARN")
            
            if attempt < 2:
                time.sleep(2)
        except Exception as e:
            log(f"Failed to check alerts (attempt {attempt + 1}/3): {e}", "WARN")
            if attempt < 2:
                time.sleep(2)
    
    log("Alert was not found in mentor backend after 3 attempts", "ERROR")
    return False

def main():
    """Run smoke tests."""
    # Default URLs (customize as needed)
    devices_url = "http://localhost:8081"
    mentor_url = "http://localhost:8080"
    
    # Allow override via args
    if len(sys.argv) > 1:
        devices_url = sys.argv[1]
    if len(sys.argv) > 2:
        mentor_url = sys.argv[2]
    
    print("=" * 60)
    log("Raqeem Smoke Test")
    print("=" * 60)
    log(f"Devices Backend: {devices_url}")
    log(f"Mentor Backend:  {mentor_url}")
    print()
    
    # Check services
    log("Checking service health...")
    devices_ok = check_service(devices_url, "Devices Backend")
    mentor_ok = check_service(mentor_url, "Mentor Backend")
    print()
    
    if not devices_ok or not mentor_ok:
        log("Services are not healthy. Make sure they're running:", "ERROR")
        log("  ./scripts/start.sh", "INFO")
        log("  OR", "INFO")
        log("  docker-compose -f docker-compose.test.yml up", "INFO")
        return False
    
    # Test alert flow
    log("Testing alert flow...")
    alert_ok = test_alert_flow(devices_url, mentor_url)
    print()
    
    print("=" * 60)
    if devices_ok and mentor_ok and alert_ok:
        log("All smoke tests passed!", "SUCCESS")
        print("=" * 60)
        return True
    else:
        log("Some smoke tests failed", "ERROR")
        print("=" * 60)
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
