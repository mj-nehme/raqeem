#!/usr/bin/env bash
# Back-compat shim: delegate to new location
exec "$(dirname "$0")/scripts/start.sh"
