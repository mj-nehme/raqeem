Device Agent (macOS)

Overview
- A lightweight macOS command-line agent that registers this machine as a device and periodically sends system metrics to the Devices Backend.
- Targets the Devices Backend API described in `docs/devices-openapi.yaml`.

Build (macOS)
```zsh
cd devices/agent-macos
swift build -c release
```

Run
```zsh
# Defaults: addr=127.0.0.1:30080 (http), interval=30s
# First run will generate and persist a device ID at ~/.raqeem/device_id.json

DEVICES_BACKEND_ADDR=127.0.0.1:30080 \
DEVICE_ID=<optional-uuid> \
INTERVAL_SECONDS=30 \
.build/release/device-agent
```

What it does
- Registers the device at `POST /devices/register` (if not yet registered) with `deviceid` and basic info (name, OS, user).
- Every interval, posts metrics to `POST /devices/{device_id}/metrics`:
  - cpu_usage (approximate), memory_total/used (MB), disk_total/used (MB), net_bytes_in/out.

Notes
- CPU temperature is not sent (requires SMC access). Set to omitted.
- The agent uses simple shell parsing for some metrics (ps, vm_stat, netstat). It’s sufficient for MVP.
- To auto-start on login, consider a LaunchAgent plist (not included by default).

Configuration
- `DEVICES_BACKEND_ADDR`: Text `ip:port` (default `127.0.0.1:30080`). Used to build `http://<addr>/api/v1`.
- `DEVICE_ID`: Optional. If not provided, a UUID is generated and stored at `~/.raqeem/device_id.json`.
- `INTERVAL_SECONDS`: Metrics collection period (default 30).
- Alternatively, put the backend address in `~/.raqeem/backend_addr.txt` (first line `ip:port`).

Uninstall
- Remove the binary and `~/.raqeem/device_id.json`.
