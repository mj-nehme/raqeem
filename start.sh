#!/usr/bin/env bash
# Main entry point - delegates to smart service discovery
exec "$(dirname "$0")/scripts/start-smart.sh"
