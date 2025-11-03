#!/usr/bin/env bash
# Main entry point - delegates to smart service discovery cleanup
exec "$(dirname "$0")/scripts/stop-smart.sh"
