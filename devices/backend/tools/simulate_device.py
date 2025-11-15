#!/usr/bin/env python3
"""
Simple device telemetry simulator.

Usage:
    # Option A: provide a full base URL
    python simulate_device.py --id device-01 --host http://127.0.0.1:<PORT> --interval 5

    # Option B: set DEVICES_BACKEND_PORT and omit --host
    DEVICES_BACKEND_PORT=<PORT> python simulate_device.py --id device-01

The script will register the device once, then periodically POST metrics, processes and activity.
"""

import argparse
import os
import random
import time
import uuid
from datetime import UTC, datetime

import requests


def rand_cpu():
    return round(random.uniform(0.5, 75.0), 2)


def rand_mem():
    return round(random.uniform(10.0, 4096.0), 2)


def random_process_list(n=5):
    procs = []
    for _i in range(n):
        pid = random.randint(1000, 50000)
        procs.append(
            {
                "pid": pid,
                "process_name": f"proc_{pid}",
                "cpu": round(random.uniform(0.0, 50.0), 2),
                "memory": round(random.uniform(1.0, 500.0), 2),
                "command_text": f"/usr/bin/proc_{pid}",
            }
        )
    return procs


def random_metrics():
    return {
        "cpu_usage": rand_cpu(),
        "memory_usage": rand_mem(),
        "net_bytes_in": random.randint(0, 1000000),
        "net_bytes_out": random.randint(0, 1000000),
    }


def random_activity():
    return {
        "activity_type": random.choice(["RandomActivity", "window_focus", "file_open"]),
        "details": "simulated",
        "timestamp": datetime.now(UTC).isoformat(),
    }


def register_device(base_url, device_id, name=None):
    payload = {
        "deviceid": device_id,
        "device_name": name or f"Sim {device_id}",
        "device_type": "simulator",
        "os": "linux",
        "ip_address": f"10.0.{random.randint(0, 255)}.{random.randint(1, 254)}",
        "current_user": random.choice(["alice", "bob", "jaafar", "guest"]),
    }
    url = f"{base_url.rstrip('/')}/devices/register"
    r = requests.post(url, json=payload, timeout=5)
    r.raise_for_status()
    return r.json()


def post_metrics(base_url, device_id, metrics):
    url = f"{base_url.rstrip('/')}/devices/{device_id}/metrics"
    payload = metrics
    r = requests.post(url, json=payload, timeout=5)
    r.raise_for_status()
    return r.json()


def post_processes(base_url, device_id, processes):
    url = f"{base_url.rstrip('/')}/devices/{device_id}/processes"
    payload = processes
    # router expects a list of process dicts
    r = requests.post(url, json=payload, timeout=5)
    r.raise_for_status()
    return r.json()


def post_activity(base_url, device_id, activity):
    url = f"{base_url.rstrip('/')}/devices/{device_id}/activities"
    # router expects a list of activity dicts
    payload = [activity]
    r = requests.post(url, json=payload, timeout=5)
    r.raise_for_status()
    return r.json()


def run_loop(base_url, device_id, interval):
    print(f"Registering device {device_id} -> {base_url}")
    try:
        res = register_device(base_url, device_id)
        print("Register response:", res)
    except Exception as e:
        print("Register failed:", e)

    count = 0
    while True:
        try:
            metrics = random_metrics()
            post_metrics(base_url, device_id, metrics)
            procs = random_process_list(random.randint(3, 8))
            post_processes(base_url, device_id, procs)
            act = random_activity()
            post_activity(base_url, device_id, act)
            count += 1
            print(
                f"[{datetime.now(UTC).isoformat()}] Sent batch #{count}: metrics={metrics['cpu_usage']}% mem={metrics['memory_usage']}MB procs={len(procs)}"
            )
        except Exception as e:
            print(f"Error sending telemetry: {e}")
        time.sleep(interval)


def parse_args():
    # Build default host from env if provided; otherwise require --host
    default_host = None
    env_port = os.getenv("DEVICES_BACKEND_PORT")
    if env_port:
        default_host = f"http://127.0.0.1:{env_port}"

    p = argparse.ArgumentParser()
    p.add_argument("--id", "-i", required=False, default=f"sim-{uuid.uuid4().hex[:8]}", help="Device id to simulate")
    p.add_argument(
        "--host",
        "-H",
        required=False,
        default=default_host,
        help="Devices backend base URL (e.g., http://127.0.0.1:PORT)",
    )
    p.add_argument("--interval", "-t", required=False, type=int, default=5, help="Seconds between telemetry posts")
    args = p.parse_args()
    if not args.host:
        _msg = "ERROR: provide --host or set DEVICES_BACKEND_PORT"
        raise SystemExit(_msg)
    return args


if __name__ == "__main__":
    args = parse_args()
    try:
        run_loop(args.host, args.id, args.interval)
    except KeyboardInterrupt:
        print("Simulator stopped")
