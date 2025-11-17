# Commands Tab UI Description

## Visual Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│ Devices                                                    [↻]      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│ Test Device #1  [●] Online                                         │
│ Test Device #2  [○] Offline                                        │
│ Test Device #3  [●] Online                                         │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│ Test Device #1                                           [Refresh] │
│ [●] Online  [Linux]  [laptop]                                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│ [CPU: 45.2%]  [Memory: 62.1%]  [Disk: 75.3%]  [Network: ↓ 1.2MB/s]│
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│ [Processes] [Activity] [Alerts] [Screenshots] [Commands]          │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│ Commands                                                            │
│                                                                     │
│ ┌────────────────────────────────────────────────────────────────┐ │
│ │ ✅ Command sent successfully!                              [X] │ │
│ └────────────────────────────────────────────────────────────────┘ │
│                                                                     │
│ ┌─────────────────────────────────────────────────────────┐        │
│ │ Enter command (e.g., get_info, status, restart)...      │ [Send]│
│ └─────────────────────────────────────────────────────────┘        │
│                                                                     │
│ Command History                                                     │
│ ┌─────────────────────────────────────────────────────────────────┐│
│ │ [ℹ] get_info                                   [completed]       ││
│ │     Status: completed • 2024-01-15 10:30:22                     ││
│ │     Result: {"cpu": 45.2, "memory": 62.1, "disk": 75.3}        ││
│ ├─────────────────────────────────────────────────────────────────┤│
│ │ [○] status                                      [pending]        ││
│ │     Status: pending • 2024-01-15 10:32:15                       ││
│ ├─────────────────────────────────────────────────────────────────┤│
│ │ [✓] check_network                              [completed]       ││
│ │     Status: completed • 2024-01-15 10:25:10                     ││
│ │     Result: Network connectivity OK                             ││
│ ├─────────────────────────────────────────────────────────────────┤│
│ │ [X] invalid_command                              [failed]        ││
│ │     Status: failed • 2024-01-15 10:20:05                        ││
│ │     Result: Command not recognized                              ││
│ └─────────────────────────────────────────────────────────────────┘│
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## UI Elements

### 1. Success Alert (Green)
- Appears after successfully sending a command
- Auto-dismisses after 3 seconds
- Can be manually dismissed with [X] button
- Text: "Command sent successfully!"

### 2. Error Alert (Red) 
- Appears when command sending fails
- Persists until manually dismissed
- Shows error message from backend
- Example: "Failed to send command: Invalid device ID"

### 3. Command Input Field
- Placeholder text: "Enter command (e.g., get_info, status, restart)..."
- Full width text field
- Clears after successful command send
- Enter key also sends command

### 4. Send Button
- Blue primary button
- Contains send icon (→) and "Send" text
- Disabled when input is empty
- Disabled when no device is selected

### 5. Command History List
- Shows most recent commands first (newest at top)
- Each command item displays:
  - Icon indicating status (info, spinner, checkmark, error)
  - Command text (bold)
  - Status badge (colored chip)
  - Timestamp in local format
  - Result text (if completed, shown in monospace font)
  - Status colors:
    - Green: completed
    - Orange/Yellow: pending (with spinner)
    - Red: failed
    - Gray: unknown

### 6. Real-time Updates
- Command list refreshes every 10 seconds automatically
- Manual refresh available via refresh button in header
- New commands appear at the top of the list
- Status updates happen automatically

## Responsive Behavior

- On mobile/tablet: Command input and button stack vertically
- On desktop: Command input and button side-by-side
- Command history items remain full width
- Long command text truncates with ellipsis
- Long results show first 100 characters with "..." indicator

## Accessibility

- All buttons have aria-labels
- Alert messages have appropriate ARIA roles
- Color is not the only indicator of status (icons + text)
- Keyboard navigation works for all interactive elements
- Focus visible on all focusable elements

## Material-UI Components Used

- `Alert` for success/error messages
- `TextField` for command input
- `Button` with `startIcon` for Send button
- `List`, `ListItem`, `ListItemText`, `ListItemIcon` for command history
- `Chip` for status badges
- `Typography` for text elements
- `CircularProgress` for pending status indicator
- `Info`, `Error`, `History` icons from `@mui/icons-material`
