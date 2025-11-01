#!/usr/bin/env bash
set -euo pipefail

# Load .env if present
if [[ -f ../../.env ]]; then
	set -a
	# shellcheck disable=SC1091
	source ../../.env
	set +a
fi

if [[ -z "${BASE_URL:-}" ]]; then
	if [[ -z "${MENTOR_BACKEND_PORT:-}" ]]; then
		echo "ERROR: Set BASE_URL or MENTOR_BACKEND_PORT in .env"
		exit 1
	fi
	BASE_URL="http://localhost:${MENTOR_BACKEND_PORT}"
fi

echo "Fetching all activities:"
curl "$BASE_URL/activities"
echo -e "\n"

echo "Fetching filtered activities:"
curl "$BASE_URL/activities?user_id=abc123&location=lab1"
echo -e "\n"
