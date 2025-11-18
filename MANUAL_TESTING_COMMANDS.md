# Manual Testing Guide: Command Display on Frontend

This document provides step-by-step instructions for manually testing the command display functionality in the Mentor Dashboard.

## Prerequisites

1. Start the Raqeem platform:
   ```bash
   ./start.sh
   ```

2. Wait for all services to be ready (check with `./scripts/discover.sh list`)

## Test Scenarios

### Scenario 1: View Commands Tab

1. Open the Mentor Dashboard (typically at http://localhost:5001 or similar)
2. Wait for devices to appear in the sidebar
3. Click on any device to select it
4. Click on the "Commands" tab
5. **Expected**: You should see a command input field with a "Send" button

### Scenario 2: Send a Command

1. Follow Scenario 1 to open the Commands tab
2. Type a command in the input field (e.g., "get_info", "status", "restart")
3. Click the "Send" button
4. **Expected**: 
   - A green success message "Command sent successfully!" should appear
   - The message should auto-dismiss after 3 seconds
   - The command input should be cleared
   - The command list should refresh automatically

### Scenario 3: View Command History

1. After sending commands (Scenario 2), scroll down in the Commands tab
2. **Expected**: You should see a list of commands with:
   - Command text (e.g., "get_info")
   - Status badge (pending, completed, or failed)
   - Timestamp when the command was created
   - Result text (if command is completed)
   - Status chip with appropriate color:
     - Green for "completed"
     - Orange/Yellow for "pending"
     - Red for "failed"

### Scenario 4: Command Status Updates

1. Send a command that will be executed by a device
2. Wait for the device to execute the command
3. Observe the command list (refreshes every 10 seconds)
4. **Expected**: 
   - Status should update from "pending" to "completed" or "failed"
   - Result text should appear for completed commands
   - Status chip color should change accordingly

### Scenario 5: Error Handling

1. To test error handling, you can either:
   - Disconnect the backend service
   - Send an invalid command (if backend validation exists)
2. Try to send a command
3. **Expected**: 
   - A red error message should appear showing the error details
   - The error message should be dismissible via the "X" button
   - The command should not be added to the history

### Scenario 6: Real-time Updates

1. Open the Mentor Dashboard in two browser windows
2. In Window 1: Select a device and open Commands tab
3. In Window 2: Select the same device and send a command
4. In Window 1: Wait up to 10 seconds
5. **Expected**: The command should appear in Window 1's command list

## Field Name Verification

### Backend API Response Format

The backend returns commands with these fields:
```json
{
  "commandid": "uuid",
  "deviceid": "uuid",
  "command_text": "get_info",
  "status": "pending",
  "created_at": "2024-01-01T12:00:00Z",
  "completed_at": "2024-01-01T12:00:05Z",
  "result": "Device information retrieved",
  "exit_code": 0
}
```

### Frontend Display

The frontend now correctly uses:
- `cmd.commandid` (not `cmd.id`)
- `cmd.command_text` (not `cmd.command`)
- `cmd.deviceid` (not `cmd.device_id`)
- `cmd.status`
- `cmd.created_at`
- `cmd.result`

### POST Request Format

When sending a command, the frontend sends:
```json
{
  "deviceid": "uuid",
  "command_text": "get_info"
}
```

**NOT** (old incorrect format):
```json
{
  "device_id": "uuid",
  "command": "get_info"
}
```

## Troubleshooting

### Commands not appearing
- Check that the backend API is running (http://localhost:30090/health)
- Check browser console for any JavaScript errors
- Verify the device is selected in the sidebar
- Wait for auto-refresh (10 seconds) or click the refresh button

### Send button disabled
- Ensure the command input is not empty
- Ensure a device is selected

### Error messages persist
- Click the "X" button on the error alert to dismiss it
- Try refreshing the page if errors continue

## API Testing with curl

You can also test the API directly:

```bash
# Get device list
curl http://localhost:30090/api/devices

# Send a command
curl -X POST http://localhost:30090/api/devices/commands \
  -H "Content-Type: application/json" \
  -d '{"deviceid": "YOUR_DEVICE_ID", "command_text": "get_info"}'

# Get command history for a device
curl http://localhost:30090/api/devices/YOUR_DEVICE_ID/commands

# Get pending commands for a device
curl http://localhost:30090/api/devices/YOUR_DEVICE_ID/commands/pending
```

## Screenshots to Capture

When testing manually, capture these screenshots:

1. **Commands Tab Empty State**: No commands yet
2. **Command Input**: Showing the input field and Send button
3. **Success Message**: Green success alert after sending a command
4. **Command List with Pending**: Commands showing "pending" status
5. **Command List with Completed**: Commands showing "completed" status with results
6. **Error Message**: Red error alert when command fails
7. **Full Dashboard**: Showing the Commands tab in context with other tabs

## Acceptance Criteria Checklist

- [ ] Commands are visible on the Mentor Dashboard frontend
- [ ] Command history is displayed with status
- [ ] Users can view command results
- [ ] Real-time updates work correctly (10s polling)
- [ ] Filtering and search work as expected (not implemented in this PR)
- [ ] UI is intuitive and responsive
- [ ] Tests cover command display functionality ✅
- [ ] Documentation is updated ✅
- [ ] Success/error messages provide clear feedback ✅
- [ ] Field names match backend API ✅
