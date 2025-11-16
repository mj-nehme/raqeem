#!/usr/bin/env python3
"""
Load Test for Raqeem IoT Platform

Tests sustained load to validate normal and peak operations:
- Continuous device operation simulation
- Frontend API load (concurrent dashboard access)
- Alert forwarding pipeline under sustained load
- Query performance monitoring
- Resource utilization tracking

Usage:
    python3 load_test.py --concurrent-users 100 --duration 300
    python3 load_test.py --concurrent-users 50 --duration 120 --verbose
"""

import argparse
import concurrent.futures
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


class LoadTestRunner:
    """Orchestrates load testing scenarios."""

    def __init__(self, devices_url: str, mentor_url: str, verbose: bool = False):
        self.devices_url = devices_url
        self.mentor_url = mentor_url
        self.verbose = verbose
        self.metrics = defaultdict(list)
        self.errors = []

    def log(self, message: str, level: str = "INFO"):
        """Print timestamped log message."""
        if level != "DEBUG" or self.verbose:
            timestamp = datetime.now().strftime("%H:%M:%S.%f")[:-3]
            symbols = {"INFO": "ℹ️", "SUCCESS": "✓", "ERROR": "✗", "WARN": "⚠️", "DEBUG": "🔍"}
            symbol = symbols.get(level, "•")
            print(f"[{timestamp}] [{level}] {symbol} {message}")

    def simulate_device_lifecycle(self, device_id: str, duration_sec: int) -> dict:
        """Simulate a complete device lifecycle under load."""
        operations = {"register": 0, "metrics": 0, "activities": 0, "alerts": 0, "queries": 0}
        latencies = defaultdict(list)
        errors = []

        start_time = time.time()
        end_time = start_time + duration_sec

        # Register device
        try:
            reg_start = time.perf_counter()
            response = requests.post(
                f"{self.devices_url}/api/v1/devices/register",
                json={
                    "deviceid": device_id,
                    "device_name": f"Load Test {device_id}",
                    "device_type": "laptop",
                    "os": "Test OS",
                },
                timeout=10,
            )
            reg_latency = (time.perf_counter() - reg_start) * 1000
            response.raise_for_status()
            operations["register"] = 1
            latencies["register"].append(reg_latency)
        except Exception as e:
            errors.append({"op": "register", "error": str(e)})

        # Continuous operations
        while time.time() < end_time:
            operation = random.choice(["metrics", "metrics", "metrics", "alert", "query"])  # Weight towards metrics

            try:
                op_start = time.perf_counter()

                if operation == "metrics":
                    response = requests.post(
                        f"{self.devices_url}/api/v1/devices/{device_id}/metrics",
                        json={
                            "cpu_usage": random.uniform(10, 90),
                            "cpu_temp": random.uniform(40, 80),
                            "memory_total": 16000000000,
                            "memory_used": random.randint(4000000000, 12000000000),
                            "swap_used": 0,
                            "disk_total": 500000000000,
                            "disk_used": random.randint(100000000000, 400000000000),
                            "net_bytes_in": random.randint(1024000, 10240000),
                            "net_bytes_out": random.randint(512000, 5120000),
                        },
                        timeout=10,
                    )
                elif operation == "alert":
                    response = requests.post(
                        f"{self.devices_url}/api/v1/devices/{device_id}/alerts",
                        json=[
                            {
                                "level": random.choice(["info", "warning", "critical"]),
                                "alert_type": random.choice(["cpu_high", "memory_high", "disk_full"]),
                                "message": f"Load test alert from {device_id}",
                                "value": random.uniform(70, 100),
                                "threshold": 80,
                            }
                        ],
                        timeout=10,
                    )
                elif operation == "query":
                    response = requests.get(f"{self.mentor_url}/devices", timeout=10)

                op_latency = (time.perf_counter() - op_start) * 1000
                response.raise_for_status()
                operations[operation] += 1
                latencies[operation].append(op_latency)

            except Exception as e:
                errors.append({"op": operation, "error": str(e)})

            # Small delay between operations (10 ops/sec per device)
            time.sleep(0.1)

        return {"device_id": device_id, "operations": operations, "latencies": latencies, "errors": errors}

    def simulate_frontend_user(self, user_id: int, duration_sec: int) -> dict:
        """Simulate a frontend user accessing the dashboard."""
        operations = {"device_list": 0, "device_details": 0, "metrics": 0, "alerts": 0}
        latencies = defaultdict(list)
        errors = []

        start_time = time.time()
        end_time = start_time + duration_sec

        while time.time() < end_time:
            operation = random.choice(["device_list", "device_list", "metrics", "alerts"])  # Weight towards listing

            try:
                op_start = time.perf_counter()

                if operation == "device_list":
                    response = requests.get(f"{self.mentor_url}/devices", timeout=10)
                elif operation == "metrics":
                    # Query random device metrics
                    device_id = f"load-device-{random.randint(0, 100)}"
                    response = requests.get(f"{self.mentor_url}/devices/{device_id}/metrics?limit=100", timeout=10)
                elif operation == "alerts":
                    # Query random device alerts
                    device_id = f"load-device-{random.randint(0, 100)}"
                    response = requests.get(f"{self.mentor_url}/devices/{device_id}/alerts", timeout=10)

                op_latency = (time.perf_counter() - op_start) * 1000
                # Don't fail on 404s for non-existent devices
                if response.status_code != 404:
                    response.raise_for_status()
                operations[operation] += 1
                latencies[operation].append(op_latency)

            except Exception as e:
                if "404" not in str(e):  # Ignore 404s
                    errors.append({"op": operation, "error": str(e)})

            # Simulate user think time
            time.sleep(random.uniform(0.5, 2))

        return {"user_id": user_id, "operations": operations, "latencies": latencies, "errors": errors}

    @staticmethod
    def _percentile(values: list[float], percentile: int) -> float:
        """Calculate percentile from list of values."""
        if not values:
            return 0
        sorted_values = sorted(values)
        index = int(len(sorted_values) * percentile / 100)
        return sorted_values[min(index, len(sorted_values) - 1)]

    def aggregate_results(self, results: list[dict], result_type: str) -> dict:
        """Aggregate results from multiple workers."""
        all_operations = defaultdict(int)
        all_latencies = defaultdict(list)
        all_errors = []

        for result in results:
            for op, count in result["operations"].items():
                all_operations[op] += count
            for op, latencies in result["latencies"].items():
                all_latencies[op].extend(latencies)
            all_errors.extend(result.get("errors", []))

        aggregated = {"type": result_type, "operations": dict(all_operations), "latencies": {}, "errors": all_errors}

        for op, latencies in all_latencies.items():
            if latencies:
                aggregated["latencies"][op] = {
                    "count": len(latencies),
                    "p50_ms": round(self._percentile(latencies, 50), 2),
                    "p95_ms": round(self._percentile(latencies, 95), 2),
                    "p99_ms": round(self._percentile(latencies, 99), 2),
                    "max_ms": round(max(latencies), 2),
                    "avg_ms": round(sum(latencies) / len(latencies), 2),
                }

        return aggregated


def main():
    """Run load tests."""
    parser = argparse.ArgumentParser(description="Load test for Raqeem IoT platform")
    parser.add_argument("--concurrent-users", type=int, default=50, help="Concurrent users/devices (default: 50)")
    parser.add_argument("--duration", type=int, default=60, help="Test duration in seconds (default: 60)")
    parser.add_argument("--devices-url", default=DEVICES_BACKEND_URL, help="Devices backend URL")
    parser.add_argument("--mentor-url", default=MENTOR_BACKEND_URL, help="Mentor backend URL")
    parser.add_argument("--verbose", action="store_true", help="Enable verbose logging")
    args = parser.parse_args()

    runner = LoadTestRunner(args.devices_url, args.mentor_url, args.verbose)

    print("=" * 80)
    runner.log("📊 Raqeem Load Test Suite")
    print("=" * 80)
    runner.log(f"Configuration: {args.concurrent_users} concurrent users, {args.duration}s duration")
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

    # Test 1: Device simulation load
    runner.log(f"Starting device simulation load test ({args.concurrent_users} devices for {args.duration}s)...")
    device_ids = [f"load-device-{int(time.time())}-{i}" for i in range(args.concurrent_users)]

    start_time = time.time()
    with concurrent.futures.ThreadPoolExecutor(max_workers=args.concurrent_users) as executor:
        device_futures = [
            executor.submit(runner.simulate_device_lifecycle, device_id, args.duration) for device_id in device_ids
        ]
        device_results = [future.result() for future in concurrent.futures.as_completed(device_futures)]
    device_duration = time.time() - start_time

    device_summary = runner.aggregate_results(device_results, "device_simulation")
    runner.log(f"Device simulation completed in {device_duration:.1f}s", "SUCCESS")
    runner.log(f"  Total operations: {sum(device_summary['operations'].values())}")
    runner.log(f"  Total errors: {len(device_summary['errors'])}")

    print()

    # Test 2: Frontend user load
    runner.log(f"Starting frontend user load test ({args.concurrent_users} users for {args.duration}s)...")

    start_time = time.time()
    with concurrent.futures.ThreadPoolExecutor(max_workers=args.concurrent_users) as executor:
        user_futures = [
            executor.submit(runner.simulate_frontend_user, i, args.duration) for i in range(args.concurrent_users)
        ]
        user_results = [future.result() for future in concurrent.futures.as_completed(user_futures)]
    user_duration = time.time() - start_time

    user_summary = runner.aggregate_results(user_results, "frontend_users")
    runner.log(f"Frontend user simulation completed in {user_duration:.1f}s", "SUCCESS")
    runner.log(f"  Total operations: {sum(user_summary['operations'].values())}")
    runner.log(f"  Total errors: {len(user_summary['errors'])}")

    print()

    # Print detailed summary
    print("=" * 80)
    runner.log("📊 Load Test Summary")
    print("=" * 80)

    summary = {
        "timestamp": datetime.now().isoformat(),
        "config": {"concurrent_users": args.concurrent_users, "duration": args.duration},
        "device_simulation": device_summary,
        "frontend_users": user_summary,
    }

    print(json.dumps(summary, indent=2))

    # Determine pass/fail
    total_errors = len(device_summary["errors"]) + len(user_summary["errors"])
    total_ops = sum(device_summary["operations"].values()) + sum(user_summary["operations"].values())
    success_rate = ((total_ops - total_errors) / total_ops * 100) if total_ops > 0 else 0

    print()
    print("=" * 80)
    if success_rate >= 95:
        runner.log(f"✅ PASS: {success_rate:.1f}% success rate ({total_ops - total_errors}/{total_ops})", "SUCCESS")
        sys.exit(0)
    else:
        runner.log(f"❌ FAIL: {success_rate:.1f}% success rate ({total_ops - total_errors}/{total_ops})", "ERROR")
        sys.exit(1)


if __name__ == "__main__":
    main()
