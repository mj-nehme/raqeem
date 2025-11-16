#!/usr/bin/env python3
"""
Stress Test for Raqeem IoT Platform

Tests high-volume operations to validate performance under extreme load:
- 1000+ device registration (burst and sustained)
- Continuous telemetry ingestion
- Concurrent alert generation and forwarding
- Bulk screenshot uploads
- Database query performance under load

Usage:
    python3 stress_test.py --devices 1000 --duration 300
    python3 stress_test.py --devices 100 --duration 60 --verbose
"""

import argparse
import concurrent.futures
import io
import json
import random
import sys
import time
from collections import defaultdict
from datetime import datetime
from typing import Any

import requests

# Configuration
DEVICES_BACKEND_URL = "http://localhost:8081"
MENTOR_BACKEND_URL = "http://localhost:8080"


class StressTestRunner:
    """Orchestrates stress testing scenarios."""

    def __init__(self, devices_url: str, mentor_url: str, verbose: bool = False):
        self.devices_url = devices_url
        self.mentor_url = mentor_url
        self.verbose = verbose
        self.results = defaultdict(lambda: defaultdict(list))
        self.errors = []

    def log(self, message: str, level: str = "INFO"):
        """Print timestamped log message."""
        if level != "DEBUG" or self.verbose:
            timestamp = datetime.now().strftime("%H:%M:%S.%f")[:-3]
            symbols = {"INFO": "ℹ️", "SUCCESS": "✓", "ERROR": "✗", "WARN": "⚠️", "DEBUG": "🔍"}
            symbol = symbols.get(level, "•")
            print(f"[{timestamp}] [{level}] {symbol} {message}")

    def measure_latency(self, func, *args, **kwargs) -> tuple[Any, float, bool]:
        """Execute function and measure latency in milliseconds."""
        start = time.perf_counter()
        try:
            result = func(*args, **kwargs)
            elapsed = (time.perf_counter() - start) * 1000
            return result, elapsed, True
        except Exception as e:
            elapsed = (time.perf_counter() - start) * 1000
            self.errors.append({"function": func.__name__, "error": str(e), "latency_ms": elapsed})
            return None, elapsed, False

    def register_device(self, device_id: str) -> tuple[bool, float]:
        """Register a single device and measure latency."""
        hash_val = abs(hash(device_id))
        payload = {
            "deviceid": device_id,
            "device_name": f"Stress-Test-{device_id}",
            "device_type": random.choice(["laptop", "desktop", "server", "mobile"]),
            "os": random.choice(["Ubuntu 22.04", "Windows 11", "macOS 14", "CentOS 7"]),
            "current_user": f"user-{hash_val % 100}",
            "device_location": f"Lab-{hash_val % 10}",
            "ip_address": f"192.168.{(hash_val % 254) + 1}.{(hash_val // 254) % 254 + 1}",
            "mac_address": f"{hash_val % 256:02X}:BB:CC:DD:EE:FF",
        }

        def _register():
            response = requests.post(f"{self.devices_url}/api/v1/devices/register", json=payload, timeout=10)
            response.raise_for_status()
            return response.json()

        _, latency, success = self.measure_latency(_register)
        return success, latency

    def submit_telemetry(self, device_id: str) -> tuple[bool, float]:
        """Submit telemetry metrics for a device."""
        cpu = random.uniform(10, 95)
        mem_pct = random.uniform(30, 85)

        payload = {
            "cpu_usage": cpu,
            "cpu_temp": 50 + cpu * 0.5,
            "memory_total": 16000000000,
            "memory_used": int(16000000000 * mem_pct / 100),
            "swap_used": random.randint(0, 1000000000),
            "disk_total": 500000000000,
            "disk_used": random.randint(100000000000, 450000000000),
            "net_bytes_in": random.randint(1024000, 10240000),
            "net_bytes_out": random.randint(512000, 5120000),
        }

        def _submit():
            response = requests.post(f"{self.devices_url}/api/v1/devices/{device_id}/metrics", json=payload, timeout=10)
            response.raise_for_status()
            return response.json()

        _, latency, success = self.measure_latency(_submit)
        return success, latency

    def submit_alert(self, device_id: str) -> tuple[bool, float]:
        """Submit alert for a device."""
        alert_types = ["cpu_high", "memory_high", "disk_full", "network_slow"]
        levels = ["info", "warning", "critical"]

        payload = [
            {
                "level": random.choice(levels),
                "alert_type": random.choice(alert_types),
                "message": f"Stress test alert from {device_id}",
                "value": random.uniform(50, 100),
                "threshold": random.uniform(70, 90),
            }
        ]

        def _submit():
            response = requests.post(f"{self.devices_url}/api/v1/devices/{device_id}/alerts", json=payload, timeout=10)
            response.raise_for_status()
            return response.json()

        _, latency, success = self.measure_latency(_submit)
        return success, latency

    def upload_screenshot(self, device_id: str, size_kb: int = 500) -> tuple[bool, float]:
        """Upload screenshot for a device."""
        # Generate fake image data
        fake_image = b"PNG" + b"\x00" * (size_kb * 1024 - 3)
        files = {"file": (f"screenshot-{device_id}.png", io.BytesIO(fake_image), "image/png")}

        def _upload():
            response = requests.post(
                f"{self.devices_url}/api/v1/devices/{device_id}/screenshot", files=files, timeout=30
            )
            response.raise_for_status()
            return response.json()

        _, latency, success = self.measure_latency(_upload)
        return success, latency

    def query_devices(self) -> tuple[bool, float]:
        """Query device list from mentor backend."""

        def _query():
            response = requests.get(f"{self.mentor_url}/devices", timeout=10)
            response.raise_for_status()
            return response.json()

        _, latency, success = self.measure_latency(_query)
        return success, latency

    def run_device_registration_test(self, num_devices: int, concurrent: int = 50) -> dict:
        """Test device registration at scale."""
        self.log(f"Starting device registration stress test: {num_devices} devices, {concurrent} concurrent")

        device_ids = [f"stress-device-{int(time.time())}-{i}" for i in range(num_devices)]
        latencies = []
        failures = 0

        start_time = time.time()

        with concurrent.futures.ThreadPoolExecutor(max_workers=concurrent) as executor:
            futures = {executor.submit(self.register_device, device_id): device_id for device_id in device_ids}

            for future in concurrent.futures.as_completed(futures):
                success, latency = future.result()
                latencies.append(latency)
                if not success:
                    failures += 1

        elapsed = time.time() - start_time
        throughput = num_devices / elapsed if elapsed > 0 else 0

        result = {
            "total": num_devices,
            "successful": num_devices - failures,
            "failed": failures,
            "duration_sec": round(elapsed, 2),
            "throughput_per_sec": round(throughput, 2),
            "latency_p50_ms": round(self._percentile(latencies, 50), 2),
            "latency_p95_ms": round(self._percentile(latencies, 95), 2),
            "latency_p99_ms": round(self._percentile(latencies, 99), 2),
            "latency_max_ms": round(max(latencies), 2) if latencies else 0,
        }

        self.log(f"Device registration: {result['successful']}/{result['total']} successful", "SUCCESS")
        self.log(f"  Throughput: {result['throughput_per_sec']:.2f} req/s")
        self.log(f"  Latency p95: {result['latency_p95_ms']:.2f}ms")

        return result

    def run_telemetry_stress_test(self, device_ids: list[str], duration_sec: int, rate_per_sec: int = 10) -> dict:
        """Test continuous telemetry ingestion."""
        self.log(f"Starting telemetry stress test: {len(device_ids)} devices for {duration_sec}s")

        latencies = []
        failures = 0
        sent = 0

        start_time = time.time()
        end_time = start_time + duration_sec

        while time.time() < end_time:
            batch_start = time.time()

            # Send telemetry for random subset of devices
            devices_to_send = random.sample(device_ids, min(rate_per_sec, len(device_ids)))

            with concurrent.futures.ThreadPoolExecutor(max_workers=rate_per_sec) as executor:
                futures = [executor.submit(self.submit_telemetry, device_id) for device_id in devices_to_send]

                for future in concurrent.futures.as_completed(futures):
                    success, latency = future.result()
                    latencies.append(latency)
                    sent += 1
                    if not success:
                        failures += 1

            # Maintain rate
            elapsed_batch = time.time() - batch_start
            if elapsed_batch < 1:
                time.sleep(1 - elapsed_batch)

        elapsed = time.time() - start_time
        throughput = sent / elapsed if elapsed > 0 else 0

        result = {
            "total_sent": sent,
            "successful": sent - failures,
            "failed": failures,
            "duration_sec": round(elapsed, 2),
            "throughput_per_sec": round(throughput, 2),
            "latency_p50_ms": round(self._percentile(latencies, 50), 2),
            "latency_p95_ms": round(self._percentile(latencies, 95), 2),
            "latency_p99_ms": round(self._percentile(latencies, 99), 2),
        }

        self.log(f"Telemetry stress: {result['successful']}/{result['total_sent']} successful", "SUCCESS")
        self.log(f"  Throughput: {result['throughput_per_sec']:.2f} msg/s")

        return result

    def run_alert_stress_test(self, device_ids: list[str], alerts_per_device: int = 10) -> dict:
        """Test alert generation and forwarding at scale."""
        self.log(f"Starting alert stress test: {len(device_ids)} devices, {alerts_per_device} alerts each")

        latencies = []
        failures = 0
        total = len(device_ids) * alerts_per_device

        start_time = time.time()

        with concurrent.futures.ThreadPoolExecutor(max_workers=50) as executor:
            futures = []
            for device_id in device_ids:
                for _ in range(alerts_per_device):
                    futures.append(executor.submit(self.submit_alert, device_id))

            for future in concurrent.futures.as_completed(futures):
                success, latency = future.result()
                latencies.append(latency)
                if not success:
                    failures += 1

        elapsed = time.time() - start_time
        throughput = total / elapsed if elapsed > 0 else 0

        result = {
            "total": total,
            "successful": total - failures,
            "failed": failures,
            "duration_sec": round(elapsed, 2),
            "throughput_per_sec": round(throughput, 2),
            "latency_p50_ms": round(self._percentile(latencies, 50), 2),
            "latency_p95_ms": round(self._percentile(latencies, 95), 2),
            "latency_p99_ms": round(self._percentile(latencies, 99), 2),
        }

        self.log(f"Alert stress: {result['successful']}/{result['total']} successful", "SUCCESS")
        self.log(f"  Throughput: {result['throughput_per_sec']:.2f} alerts/s")

        return result

    def run_screenshot_stress_test(self, device_ids: list[str], screenshots_per_device: int = 5) -> dict:
        """Test screenshot upload at scale."""
        self.log(f"Starting screenshot stress test: {len(device_ids)} devices, {screenshots_per_device} screenshots each")

        latencies = []
        failures = 0
        total = len(device_ids) * screenshots_per_device

        start_time = time.time()

        with concurrent.futures.ThreadPoolExecutor(max_workers=20) as executor:
            futures = []
            for device_id in device_ids:
                for _ in range(screenshots_per_device):
                    futures.append(executor.submit(self.upload_screenshot, device_id, random.randint(100, 1000)))

            for future in concurrent.futures.as_completed(futures):
                success, latency = future.result()
                latencies.append(latency)
                if not success:
                    failures += 1

        elapsed = time.time() - start_time
        throughput = total / elapsed if elapsed > 0 else 0

        result = {
            "total": total,
            "successful": total - failures,
            "failed": failures,
            "duration_sec": round(elapsed, 2),
            "throughput_per_sec": round(throughput, 2),
            "latency_p50_ms": round(self._percentile(latencies, 50), 2),
            "latency_p95_ms": round(self._percentile(latencies, 95), 2),
            "latency_p99_ms": round(self._percentile(latencies, 99), 2),
        }

        self.log(f"Screenshot stress: {result['successful']}/{result['total']} successful", "SUCCESS")
        self.log(f"  Throughput: {result['throughput_per_sec']:.2f} uploads/s")

        return result

    def run_query_stress_test(self, queries: int = 1000) -> dict:
        """Test database query performance under load."""
        self.log(f"Starting query stress test: {queries} queries")

        latencies = []
        failures = 0

        start_time = time.time()

        with concurrent.futures.ThreadPoolExecutor(max_workers=50) as executor:
            futures = [executor.submit(self.query_devices) for _ in range(queries)]

            for future in concurrent.futures.as_completed(futures):
                success, latency = future.result()
                latencies.append(latency)
                if not success:
                    failures += 1

        elapsed = time.time() - start_time
        throughput = queries / elapsed if elapsed > 0 else 0

        result = {
            "total": queries,
            "successful": queries - failures,
            "failed": failures,
            "duration_sec": round(elapsed, 2),
            "throughput_per_sec": round(throughput, 2),
            "latency_p50_ms": round(self._percentile(latencies, 50), 2),
            "latency_p95_ms": round(self._percentile(latencies, 95), 2),
            "latency_p99_ms": round(self._percentile(latencies, 99), 2),
        }

        self.log(f"Query stress: {result['successful']}/{result['total']} successful", "SUCCESS")
        self.log(f"  Throughput: {result['throughput_per_sec']:.2f} queries/s")

        return result

    @staticmethod
    def _percentile(values: list[float], percentile: int) -> float:
        """Calculate percentile from list of values."""
        if not values:
            return 0
        sorted_values = sorted(values)
        index = int(len(sorted_values) * percentile / 100)
        return sorted_values[min(index, len(sorted_values) - 1)]


def main():
    """Run comprehensive stress tests."""
    parser = argparse.ArgumentParser(description="Stress test for Raqeem IoT platform")
    parser.add_argument("--devices", type=int, default=100, help="Number of devices to simulate (default: 100)")
    parser.add_argument("--duration", type=int, default=60, help="Test duration in seconds (default: 60)")
    parser.add_argument("--devices-url", default=DEVICES_BACKEND_URL, help="Devices backend URL")
    parser.add_argument("--mentor-url", default=MENTOR_BACKEND_URL, help="Mentor backend URL")
    parser.add_argument("--verbose", action="store_true", help="Enable verbose logging")
    args = parser.parse_args()

    runner = StressTestRunner(args.devices_url, args.mentor_url, args.verbose)

    print("=" * 80)
    runner.log("🔥 Raqeem Stress Test Suite")
    print("=" * 80)
    runner.log(f"Configuration: {args.devices} devices, {args.duration}s duration")
    runner.log(f"Devices Backend: {args.devices_url}")
    runner.log(f"Mentor Backend:  {args.mentor_url}")
    print()

    # Check service health
    runner.log("Checking service health...")
    try:
        requests.get(f"{args.devices_url}/health", timeout=5).raise_for_status()
        requests.get(f"{args.mentor_url}/health", timeout=5).raise_for_status()
        runner.log("Services are healthy", "SUCCESS")
    except Exception as e:
        runner.log(f"Services are not healthy: {e}", "ERROR")
        sys.exit(1)

    print()

    # Run stress tests
    results = {}

    # 1. Device registration stress
    results["device_registration"] = runner.run_device_registration_test(args.devices)
    device_ids = [f"stress-device-{int(time.time())}-{i}" for i in range(min(args.devices, 100))]
    print()

    # 2. Telemetry stress (use subset of devices)
    results["telemetry"] = runner.run_telemetry_stress_test(device_ids[:50], duration_sec=min(args.duration, 60))
    print()

    # 3. Alert stress
    results["alerts"] = runner.run_alert_stress_test(device_ids[:20], alerts_per_device=5)
    print()

    # 4. Screenshot stress
    results["screenshots"] = runner.run_screenshot_stress_test(device_ids[:10], screenshots_per_device=3)
    print()

    # 5. Query stress
    results["queries"] = runner.run_query_stress_test(queries=500)
    print()

    # Print summary
    print("=" * 80)
    runner.log("📊 Stress Test Summary")
    print("=" * 80)

    summary = {
        "timestamp": datetime.now().isoformat(),
        "config": {"devices": args.devices, "duration": args.duration},
        "results": results,
        "errors": runner.errors[:10],  # Limit to first 10 errors
        "total_errors": len(runner.errors),
    }

    print(json.dumps(summary, indent=2))

    # Determine pass/fail
    total_failures = sum(r.get("failed", 0) for r in results.values())
    total_operations = sum(r.get("total", 0) for r in results.values())
    success_rate = ((total_operations - total_failures) / total_operations * 100) if total_operations > 0 else 0

    print()
    print("=" * 80)
    if success_rate >= 95:
        runner.log(f"✅ PASS: {success_rate:.1f}% success rate ({total_operations - total_failures}/{total_operations})", "SUCCESS")
        sys.exit(0)
    else:
        runner.log(f"❌ FAIL: {success_rate:.1f}% success rate ({total_operations - total_failures}/{total_operations})", "ERROR")
        sys.exit(1)


if __name__ == "__main__":
    main()
