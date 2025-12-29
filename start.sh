#!/usr/bin/env bash
set -e

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

usage() {
	cat <<USAGE
Usage: ./start.sh [target]

Targets:
	all                      Start infra (Postgres+MinIO) and all components
	infra                    Start Postgres+MinIO
	postgres                 Start Postgres only
	minio                    Start MinIO only
	mentor-backend           Start Mentor backend
	devices-backend          Start Devices backend
	dashboard                Start Mentor dashboard frontend
	devices-web-simulator    Start Devices web simulator frontend
	smart                    Start the original smart all-in-one flow

Environment:
	NAMESPACE                Kubernetes namespace (default: "default")
	AUTO_CLEAN_PVC=true      Auto-delete lingering Postgres PVC during infra start
USAGE
}

TARGET=${1:-all}

case "$TARGET" in
	all)
		"$ROOT_DIR/scripts/start-infra.sh"
		"$ROOT_DIR/mentor/backend/scripts/start.sh"
		"$ROOT_DIR/devices/backend/scripts/start.sh"
		"$ROOT_DIR/mentor/dashboard/scripts/start.sh"
		"$ROOT_DIR/devices/web-simulator/scripts/start.sh"
		;;
	infra|minio|postgres)
		"$ROOT_DIR/scripts/start-infra.sh" "$TARGET"
		;;
	mentor-backend)
		"$ROOT_DIR/mentor/backend/scripts/start.sh"
		;;
	devices-backend)
		"$ROOT_DIR/devices/backend/scripts/start.sh"
		;;
	dashboard)
		"$ROOT_DIR/mentor/dashboard/scripts/start.sh"
		;;
	devices-web-simulator)
		"$ROOT_DIR/devices/web-simulator/scripts/start.sh"
		;;
	smart)
		"$ROOT_DIR/scripts/start-smart.sh"
		;;
	-h|--help)
		usage
		;;
	*)
		echo "❌ Unknown target: $TARGET"
		usage
		exit 1
		;;
esac
