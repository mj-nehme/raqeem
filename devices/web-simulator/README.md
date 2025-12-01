# Device Simulator Frontend

An interactive device simulator for testing the Raqeem monitoring platform. Built with React + Vite.

## Features

### 🖥️ Device Registration
- Generate random device IDs or use custom ones
- Configure device properties:
  - Device name
  - Type (laptop/desktop/mobile/tablet)
  - Operating System (macOS/Windows/Linux/iOS/Android)
  - Current user
- Register devices with the backend

### 🚀 Auto-Simulation Mode
Automatically sends realistic data every 5 seconds:
- **Metrics** (always): CPU usage, memory, disk, network stats
- **Activities** (50% chance): App usage, file access, network activity
- **Alerts** (20% chance): CPU high, memory warnings, security alerts
- **Screenshots** (30% chance): Simulated screenshot uploads

### 🎮 Manual Controls
Individual buttons to trigger specific actions:
- 📊 **Send Metrics** - System performance data
- 📝 **Send Activities** - Application usage logs
- ⚠️ **Send Alert** - Warning/critical alerts
- 📸 **Send Screenshot** - Simulated screenshot upload

### 📊 Real-Time Statistics
Track counts of sent data:
- Metrics sent
- Activities sent
- Alerts sent
- Screenshots uploaded

### 📝 Activity Logs
- Real-time log of all actions
- Color-coded by type (success/error/warning/info)
- Timestamps for each action
- Scrollable history (last 50 events)

### 🎨 Modern UI
- Beautiful gradient background
- Responsive card-based layout
- Smooth animations and hover effects
- Mobile-friendly design

## Getting Started

### Prerequisites
- Node.js 18+ and npm
- Devices backend reachable at the URL provided via env (no hardcoded ports)

### Installation

```bash
npm install
```

### Development

Use the repository root `.env` to control ports and URLs:

```bash
cp .env.example .env
# Edit .env to set DEVICES_FRONTEND_PORT and DEVICES_BACKEND_PORT
```

Run the frontend:

```bash
npm run dev
```

The dev server reads its port from `VITE_DEVICES_FRONTEND_PORT`.

### Build for Production

```bash
npm run build
npm run preview
```

## Usage

1. **Open the simulator** at the URL printed by the start script (no fixed port)

2. **Configure your device:**
   - The device ID is auto-generated, or enter your own
   - Set device name, type, and OS
   - Enter current user name

3. **Register the device:**
   - Click "Register Device" button
   - Wait for confirmation in the logs

4. **Start simulation:**
   - Click "▶️ Start Auto Simulation" for continuous data sending
   - Or use manual buttons to send specific data types

5. **Monitor activity:**
   - Watch the statistics counters increase
   - Check the activity logs for detailed information

6. **View results:**
  - Open the Mentor Dashboard at the URL printed by the start script (no fixed port)
  - Your simulated device should appear with all the data

## API Integration

The simulator connects to the Devices Backend API:

- `POST /api/v1/devices/register` - Register device
- `POST /api/v1/devices/{id}/metrics` - Send system metrics
- `POST /api/v1/devices/{id}/activities` - Send activity logs
- `POST /api/v1/devices/{id}/alerts` - Send alerts
- `POST /api/v1/screenshots` - Upload screenshots

Backend URL is configured via env var `VITE_DEVICES_API_URL` (set by the start script based on `.env`).

## Development Tips

### Adding New Data Types

To add a new type of data to send:

1. Add a new function in `DeviceSimulator.jsx`:
```javascript
const sendNewDataType = async () => {
  try {
    const data = { /* your data */ };
    const response = await fetch(`${BACKEND_URL}/devices/${deviceId}/newtype`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    });
    if (response.ok) {
      addLog('✓ New data sent', 'success');
    }
  } catch (error) {
    addLog(`✗ Error: ${error.message}`, 'error');
  }
};
```

2. Add a button in the manual controls section
3. Add the call to the auto-simulation interval if needed

### Customizing Simulation Intervals

Edit the interval timing in the `useEffect` hook:
```javascript
const interval = setInterval(() => {
  // Your simulation logic
}, 5000); // Change this value (in milliseconds)
```

### Adjusting Data Probabilities

Modify the random chance checks:
```javascript
// 70% chance instead of 50%
if (Math.random() > 0.3) {
  sendActivities();
}
```

## Technologies

- **React 18** - UI framework
- **Vite** - Build tool and dev server
- **CSS3** - Custom styling with gradients and animations
- **Fetch API** - HTTP requests to backend

## Project Structure

```
src/
├── App.jsx                          # Main app component
├── App.css                          # App-level styles
├── main.jsx                         # Entry point
├── components/
│   ├── DeviceSimulator.jsx         # Main simulator component
│   └── DeviceSimulator.css         # Simulator styles
└── assets/                          # Static assets
```

## Contributing

When adding new features:
1. Keep the UI responsive and mobile-friendly
2. Add appropriate error handling
3. Update logs for user feedback
4. Maintain color-coding for log types
5. Test with the Mentor Dashboard to verify data flow

## License

Part of the Raqeem monitoring platform.
