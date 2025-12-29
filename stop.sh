#!/usr/bin/env bash
set -e

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

usage() {
	cat <<USAGE
Usage: ./stop.sh [target] [--clean]

Targets:
	all                      Stop all components and infra
	infra                    Stop Postgres+MinIO (with optional --clean PVCs)
	postgres                 Stop Postgres only (with optional --clean PVCs)
	minio                    Stop MinIO only
	mentor-backend           Stop Mentor backend
	devices-backend          Stop Devices backend
	dashboard                Stop Mentor dashboard frontend
	devices-web-simulator    Stop Devices web simulator frontend
	smart                    Use original smart teardown flow

Flags:
	--clean                  Delete PVCs (only applies to infra/postgres)
USAGE
}

TARGET=${1:-all}
CLEAN_FLAG=""
if [[ "$2" == "--clean" ]] || [[ "$2" == "-c" ]]; then
	CLEAN_FLAG="$2"
fi

case "$TARGET" in
	all)
		# Frontends first, then backends, then infra
		"$ROOT_DIR/mentor/dashboard/scripts/stop.sh" || true
		"$ROOT_DIR/devices/web-simulator/scripts/stop.sh" || true
		"$ROOT_DIR/devices/backend/scripts/stop.sh" || true
		"$ROOT_DIR/mentor/backend/scripts/stop.sh" || true
		"$ROOT_DIR/scripts/stop-infra.sh" "$CLEAN_FLAG" || true
		;;
	infra)
		"$ROOT_DIR/scripts/stop-infra.sh" "$CLEAN_FLAG"
		;;
	postgres)
		"$ROOT_DIR/scripts/stop-infra.sh" "$CLEAN_FLAG"
		;;
	minio)
		# Stop only MinIO by uninstalling release; ignore PVCs
		NAMESPACE=${NAMESPACE:-default} helm uninstall minio -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ minio not found"
		;;
	mentor-backend)
		"$ROOT_DIR/mentor/backend/scripts/stop.sh"
		;;
	devices-backend)
		"$ROOT_DIR/devices/backend/scripts/stop.sh"
		;;
	dashboard)
		"$ROOT_DIR/mentor/dashboard/scripts/stop.sh"
		;;
	devices-web-simulator)
		"$ROOT_DIR/devices/web-simulator/scripts/stop.sh"
		;;
	smart)
		"$ROOT_DIR/scripts/stop-smart.sh" "$CLEAN_FLAG"
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
