#!/bin/bash
# CI lint check for legacy field names
# This script ensures that legacy field names are not used in source code

set -e

echo "🔍 Checking for legacy field names in source code..."

FAILED=0

# Directories to check
DIRS_TO_CHECK=(
  "devices/backend/src/app"
  "mentor/backend/src"
  "devices/frontend/src"
  "mentor/frontend/src"
  "tests"
  "devices/backend/tools"
)

# Patterns to search for (excluding comments and documentation)
check_pattern() {
  local pattern=$1
  local message=$2
  local files_to_check=$3
  
  echo "Checking for: $message"
  
  # Search for the pattern, excluding certain files/patterns
  if grep -rn "$pattern" $files_to_check \
      --include="*.py" \
      --include="*.go" \
      --include="*.jsx" \
      --include="*.js" \
      --include="*.ts" \
      --include="*.tsx" \
      --exclude="*test*.py" \
      --exclude="*test*.go" \
      --exclude="*test*.js" \
      --exclude="*test*.jsx" \
      --exclude="*.md" \
      --exclude="*.txt" \
      --exclude="*.log" \
      --exclude-dir=node_modules \
      --exclude-dir=.git \
      --exclude-dir=dist \
      --exclude-dir=build \
      --exclude-dir=coverage \
      2>/dev/null; then
    echo "❌ FAIL: Found legacy field pattern: $message"
    FAILED=1
  else
    echo "✅ PASS: No legacy field pattern found: $message"
  fi
}

# Check for legacy fields in JSON payloads
check_pattern '"id":.*"device' 'Legacy "id" field in device context (use "deviceid")' "${DIRS_TO_CHECK[@]}"
check_pattern '"name":.*device' 'Legacy "name" field in device context (use "device_name")' "${DIRS_TO_CHECK[@]}"
check_pattern '"location":' 'Legacy "location" field (use "device_location")' "${DIRS_TO_CHECK[@]}"
check_pattern '"type":.*activity' 'Legacy "type" field in activity context (use "activity_type")' "${DIRS_TO_CHECK[@]}"
check_pattern '"type":.*alert' 'Legacy "type" field in alert context (use "alert_type")' "${DIRS_TO_CHECK[@]}"
check_pattern '"command":.*"/.*"' 'Legacy "command" field in process context (use "command_text")' "${DIRS_TO_CHECK[@]}"

# Check for legacy field access patterns in code
check_pattern '\.get\("id"\)' 'Legacy field access: .get("id") (use .get("deviceid"))' "${DIRS_TO_CHECK[@]}"
check_pattern '\.get\("name"\)' 'Legacy field access: .get("name") (use specific field name)' "${DIRS_TO_CHECK[@]}"
check_pattern '\.get\("type"\)' 'Legacy field access: .get("type") (use specific type field)' "${DIRS_TO_CHECK[@]}"
check_pattern '\.get\("command"\)' 'Legacy field access: .get("command") (use .get("command_text"))' "${DIRS_TO_CHECK[@]}"
check_pattern '\.get\("location"\)' 'Legacy field access: .get("location") (use .get("device_location"))' "${DIRS_TO_CHECK[@]}"

if [ $FAILED -eq 1 ]; then
  echo ""
  echo "❌ Legacy field names found! Please update to use canonical naming."
  echo "See docs/API.md for the complete field name mapping."
  exit 1
else
  echo ""
  echo "✅ All checks passed! No legacy field names found."
  exit 0
fi
