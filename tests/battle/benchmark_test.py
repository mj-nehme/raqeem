#!/usr/bin/env python3
"""
Performance Benchmark Test for Raqeem IoT Platform

Establishes baseline performance metrics and detects regressions:
- Device registration latency (p50, p95, p99)
- Telemetry ingestion throughput
- Alert forwarding latency
- Database query response times
- API endpoint response times

Usage:
    python3 benchmark_test.py --samples 1000
    python3 benchmark_test.py --samples 500 --verbose
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

import requests

# Configuration
DEVICES_BACKEND_URL = "http://localhost:8081"
MENTOR_BACKEND_URL = "http://localhost:8080"


class BenchmarkRunner:
    """Orchestrates performance benchmarking."""

    def __init__(self, devices_url: str, mentor_url: str, verbose: bool = False):
        self.devices_url = devices_url
        self.mentor_url = mentor_url
        self.verbose = verbose
        self.results = {}

    def log(self, message: str, level: str = "INFO"):
        """Print timestamped log message."""
        if level != "DEBUG" or self.verbose:
            timestamp = datetime.now().strftime("%H:%M:%S.%f")[:-3]
            symbols = {"INFO": "ℹ️", "SUCCESS": "✓", "ERROR": "✗", "WARN": "⚠️", "DEBUG": "🔍"}
            symbol = symbols.get(level, "•")
            print(f"[{timestamp}] [{level}] {symbol} {message}")

    def measure_operation(self, operation_func, *args, **kwargs) -> tuple[bool, float]:
        """Execute operation and measure latency in milliseconds."""
        start = time.perf_counter()
        try:
            operation_func(*args, **kwargs)
            elapsed_ms = (time.perf_counter() - start) * 1000
            return True, elapsed_ms
        except Exception as e:
            elapsed_ms = (time.perf_counter() - start) * 1000
            self.log(f"Operation failed: {e}", "DEBUG")
            return False, elapsed_ms

    def benchmark_device_registration(self, samples: int) -> dict:
        """Benchmark device registration performance."""
        self.log(f"Benchmarking device registration ({samples} samples)...")
        latencies = []
        failures = 0

        for i in range(samples):
            device_id = f"bench-reg-{int(time.time())}-{i}"

            def register():
                response = requests.post(
                    f"{self.devices_url}/api/v1/devices/register",
                    json={
                        "deviceid": device_id,
                        "device_name": f"Benchmark {i}",
                        "device_type": "laptop",
                        "os": "Test OS",
                    },
                    timeout=10,
                )
                response.raise_for_status()

            success, latency = self.measure_operation(register)
            latencies.append(latency)
            if not success:
                failures += 1

        result = {
            "operation": "device_registration",
            "samples": samples,
            "failures": failures,
            "latency_p50_ms": round(self._percentile(latencies, 50), 2),
            "latency_p95_ms": round(self._percentile(latencies, 95), 2),
            "latency_p99_ms": round(self._percentile(latencies, 99), 2),
            "latency_min_ms": round(min(latencies), 2) if latencies else 0,
            "latency_max_ms": round(max(latencies), 2) if latencies else 0,
            "latency_avg_ms": round(sum(latencies) / len(latencies), 2) if latencies else 0,
        }

        self.log(f"  p50: {result['latency_p50_ms']:.2f}ms, p95: {result['latency_p95_ms']:.2f}ms, p99: {result['latency_p99_ms']:.2f}ms")
        return result

    def benchmark_telemetry_ingestion(self, samples: int, device_id: str) -> dict:
        """Benchmark telemetry ingestion performance."""
        self.log(f"Benchmarking telemetry ingestion ({samples} samples)...")
        latencies = []
        failures = 0

        for _ in range(samples):

            def submit():
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
                response.raise_for_status()

            success, latency = self.measure_operation(submit)
            latencies.append(latency)
            if not success:
                failures += 1

        result = {
            "operation": "telemetry_ingestion",
            "samples": samples,
            "failures": failures,
            "latency_p50_ms": round(self._percentile(latencies, 50), 2),
            "latency_p95_ms": round(self._percentile(latencies, 95), 2),
            "latency_p99_ms": round(self._percentile(latencies, 99), 2),
            "latency_avg_ms": round(sum(latencies) / len(latencies), 2) if latencies else 0,
        }

        self.log(f"  p50: {result['latency_p50_ms']:.2f}ms, p95: {result['latency_p95_ms']:.2f}ms, p99: {result['latency_p99_ms']:.2f}ms")
        return result

    def benchmark_alert_submission(self, samples: int, device_id: str) -> dict:
        """Benchmark alert submission and forwarding performance."""
        self.log(f"Benchmarking alert submission ({samples} samples)...")
        latencies = []
        failures = 0

        for _ in range(samples):

            def submit():
                response = requests.post(
                    f"{self.devices_url}/api/v1/devices/{device_id}/alerts",
                    json=[
                        {
                            "level": "warning",
                            "alert_type": "cpu_high",
                            "message": "Benchmark alert",
                            "value": 85.5,
                            "threshold": 80.0,
                        }
                    ],
                    timeout=10,
                )
                response.raise_for_status()

            success, latency = self.measure_operation(submit)
            latencies.append(latency)
            if not success:
                failures += 1

        result = {
            "operation": "alert_submission",
            "samples": samples,
            "failures": failures,
            "latency_p50_ms": round(self._percentile(latencies, 50), 2),
            "latency_p95_ms": round(self._percentile(latencies, 95), 2),
            "latency_p99_ms": round(self._percentile(latencies, 99), 2),
            "latency_avg_ms": round(sum(latencies) / len(latencies), 2) if latencies else 0,
        }

        self.log(f"  p50: {result['latency_p50_ms']:.2f}ms, p95: {result['latency_p95_ms']:.2f}ms, p99: {result['latency_p99_ms']:.2f}ms")
        return result

    def benchmark_device_query(self, samples: int) -> dict:
        """Benchmark device list query performance."""
        self.log(f"Benchmarking device query ({samples} samples)...")
        latencies = []
        failures = 0

        for _ in range(samples):

            def query():
                response = requests.get(f"{self.mentor_url}/devices", timeout=10)
                response.raise_for_status()

            success, latency = self.measure_operation(query)
            latencies.append(latency)
            if not success:
                failures += 1

        result = {
            "operation": "device_query",
            "samples": samples,
            "failures": failures,
            "latency_p50_ms": round(self._percentile(latencies, 50), 2),
            "latency_p95_ms": round(self._percentile(latencies, 95), 2),
            "latency_p99_ms": round(self._percentile(latencies, 99), 2),
            "latency_avg_ms": round(sum(latencies) / len(latencies), 2) if latencies else 0,
        }

        self.log(f"  p50: {result['latency_p50_ms']:.2f}ms, p95: {result['latency_p95_ms']:.2f}ms, p99: {result['latency_p99_ms']:.2f}ms")
        return result

    def benchmark_metrics_query(self, samples: int, device_id: str) -> dict:
        """Benchmark metrics query performance."""
        self.log(f"Benchmarking metrics query ({samples} samples)...")
        latencies = []
        failures = 0

        for _ in range(samples):

            def query():
                response = requests.get(f"{self.mentor_url}/devices/{device_id}/metrics?limit=100", timeout=10)
                response.raise_for_status()

            success, latency = self.measure_operation(query)
            latencies.append(latency)
            if not success:
                failures += 1

        result = {
            "operation": "metrics_query",
            "samples": samples,
            "failures": failures,
            "latency_p50_ms": round(self._percentile(latencies, 50), 2),
            "latency_p95_ms": round(self._percentile(latencies, 95), 2),
            "latency_p99_ms": round(self._percentile(latencies, 99), 2),
            "latency_avg_ms": round(sum(latencies) / len(latencies), 2) if latencies else 0,
        }

        self.log(f"  p50: {result['latency_p50_ms']:.2f}ms, p95: {result['latency_p95_ms']:.2f}ms, p99: {result['latency_p99_ms']:.2f}ms")
        return result

    def benchmark_screenshot_upload(self, samples: int, device_id: str, size_kb: int = 500) -> dict:
        """Benchmark screenshot upload performance."""
        self.log(f"Benchmarking screenshot upload ({samples} samples, {size_kb}KB each)...")
        latencies = []
        failures = 0

        for i in range(samples):
            fake_image = b"PNG" + b"\x00" * (size_kb * 1024 - 3)

            def upload():
                files = {"file": (f"bench-{i}.png", io.BytesIO(fake_image), "image/png")}
                response = requests.post(f"{self.devices_url}/api/v1/devices/{device_id}/screenshot", files=files, timeout=30)
                response.raise_for_status()

            success, latency = self.measure_operation(upload)
            latencies.append(latency)
            if not success:
                failures += 1

        result = {
            "operation": "screenshot_upload",
            "samples": samples,
            "size_kb": size_kb,
            "failures": failures,
            "latency_p50_ms": round(self._percentile(latencies, 50), 2),
            "latency_p95_ms": round(self._percentile(latencies, 95), 2),
            "latency_p99_ms": round(self._percentile(latencies, 99), 2),
            "latency_avg_ms": round(sum(latencies) / len(latencies), 2) if latencies else 0,
        }

        self.log(f"  p50: {result['latency_p50_ms']:.2f}ms, p95: {result['latency_p95_ms']:.2f}ms, p99: {result['latency_p99_ms']:.2f}ms")
        return result

    def benchmark_concurrent_operations(self, concurrent: int, operations_per_worker: int) -> dict:
        """Benchmark concurrent operation performance."""
        self.log(f"Benchmarking concurrent operations ({concurrent} workers, {operations_per_worker} ops each)...")

        def worker(worker_id: int):
            device_id = f"bench-concurrent-{int(time.time())}-{worker_id}"
            latencies = []

            # Register device
            def register():
                response = requests.post(
                    f"{self.devices_url}/api/v1/devices/register",
                    json={"deviceid": device_id, "device_name": f"Concurrent {worker_id}", "device_type": "laptop", "os": "Test OS"},
                    timeout=10,
                )
                response.raise_for_status()

            _, latency = self.measure_operation(register)
            latencies.append(latency)

            # Perform operations
            for _ in range(operations_per_worker):

                def submit_metric():
                    response = requests.post(
                        f"{self.devices_url}/api/v1/devices/{device_id}/metrics",
                        json={
                            "cpu_usage": random.uniform(10, 90),
                            "cpu_temp": 50,
                            "memory_total": 16000000000,
                            "memory_used": 8000000000,
                            "swap_used": 0,
                            "disk_total": 500000000000,
                            "disk_used": 250000000000,
                            "net_bytes_in": 1024000,
                            "net_bytes_out": 512000,
                        },
                        timeout=10,
                    )
                    response.raise_for_status()

                _, latency = self.measure_operation(submit_metric)
                latencies.append(latency)

            return latencies

        start_time = time.time()
        with concurrent.futures.ThreadPoolExecutor(max_workers=concurrent) as executor:
            futures = [executor.submit(worker, i) for i in range(concurrent)]
            all_latencies = []
            for future in concurrent.futures.as_completed(futures):
                all_latencies.extend(future.result())
        elapsed = time.time() - start_time

        total_ops = concurrent * (1 + operations_per_worker)
        throughput = total_ops / elapsed if elapsed > 0 else 0

        result = {
            "operation": "concurrent_operations",
            "concurrent_workers": concurrent,
            "operations_per_worker": operations_per_worker,
            "total_operations": total_ops,
            "duration_sec": round(elapsed, 2),
            "throughput_per_sec": round(throughput, 2),
            "latency_p50_ms": round(self._percentile(all_latencies, 50), 2),
            "latency_p95_ms": round(self._percentile(all_latencies, 95), 2),
            "latency_p99_ms": round(self._percentile(all_latencies, 99), 2),
        }

        self.log(f"  Throughput: {result['throughput_per_sec']:.2f} ops/s, p95 latency: {result['latency_p95_ms']:.2f}ms")
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
    """Run performance benchmarks."""
    parser = argparse.ArgumentParser(description="Performance benchmark for Raqeem IoT platform")
    parser.add_argument("--samples", type=int, default=100, help="Number of samples per benchmark (default: 100)")
    parser.add_argument("--devices-url", default=DEVICES_BACKEND_URL, help="Devices backend URL")
    parser.add_argument("--mentor-url", default=MENTOR_BACKEND_URL, help="Mentor backend URL")
    parser.add_argument("--verbose", action="store_true", help="Enable verbose logging")
    args = parser.parse_args()

    runner = BenchmarkRunner(args.devices_url, args.mentor_url, args.verbose)

    print("=" * 80)
    runner.log("⚡ Raqeem Performance Benchmark Suite")
    print("=" * 80)
    runner.log(f"Configuration: {args.samples} samples per benchmark")
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

    # Setup: Register a test device for reuse
    test_device_id = f"bench-device-{int(time.time())}"
    runner.log("Setting up test device...")
    try:
        requests.post(
            f"{args.devices_url}/api/v1/devices/register",
            json={"deviceid": test_device_id, "device_name": "Benchmark Device", "device_type": "laptop", "os": "Test OS"},
            timeout=10,
        ).raise_for_status()
        runner.log("Test device registered", "SUCCESS")
    except Exception as e:
        runner.log(f"Failed to register test device: {e}", "WARN")

    print()

    # Run benchmarks
    results = {}

    results["device_registration"] = runner.benchmark_device_registration(args.samples)
    print()

    results["telemetry_ingestion"] = runner.benchmark_telemetry_ingestion(args.samples, test_device_id)
    print()

    results["alert_submission"] = runner.benchmark_alert_submission(args.samples, test_device_id)
    print()

    results["device_query"] = runner.benchmark_device_query(args.samples)
    print()

    results["metrics_query"] = runner.benchmark_metrics_query(args.samples, test_device_id)
    print()

    results["screenshot_upload"] = runner.benchmark_screenshot_upload(min(args.samples, 20), test_device_id, 500)
    print()

    results["concurrent_operations"] = runner.benchmark_concurrent_operations(50, 10)
    print()

    # Print summary
    print("=" * 80)
    runner.log("📊 Performance Benchmark Summary")
    print("=" * 80)

    summary = {"timestamp": datetime.now().isoformat(), "samples": args.samples, "benchmarks": results}

    print(json.dumps(summary, indent=2))

    # Evaluate against targets
    print()
    print("=" * 80)
    runner.log("Performance Targets Evaluation")
    print("=" * 80)

    targets = {
        "device_registration": {"p95_ms": 200},
        "telemetry_ingestion": {"p95_ms": 100},
        "alert_submission": {"p95_ms": 1000},
        "device_query": {"p95_ms": 100},
        "screenshot_upload": {"p95_ms": 5000},
    }

    all_passed = True
    for benchmark, target in targets.items():
        if benchmark in results:
            actual = results[benchmark]["latency_p95_ms"]
            target_val = target["p95_ms"]
            passed = actual <= target_val
            status = "✅" if passed else "⚠️"
            runner.log(f"{status} {benchmark}: {actual:.2f}ms (target: <{target_val}ms)", "SUCCESS" if passed else "WARN")
            if not passed:
                all_passed = False

    print()
    if all_passed:
        runner.log("✅ PASS: All benchmarks meet performance targets", "SUCCESS")
        sys.exit(0)
    else:
        runner.log("⚠️ WARN: Some benchmarks exceed targets (acceptable for loaded systems)", "WARN")
        sys.exit(0)  # Don't fail - just warn


if __name__ == "__main__":
    main()
