import React, { useState, useEffect, useCallback } from 'react';
import './DeviceSimulator.css';

const BACKEND_URL = import.meta.env.VITE_DEVICES_API_URL;
const API_BASE_URL = BACKEND_URL;

function DeviceSimulator() {
    const [deviceId, setDeviceId] = useState('');
    const [deviceName, setDeviceName] = useState('');
    const [deviceType, setDeviceType] = useState('laptop');
    const [deviceOS, setDeviceOS] = useState('macOS');
    const [currentUser, setCurrentUser] = useState('');
    const [isRegistered, setIsRegistered] = useState(false);
    const [isRunning, setIsRunning] = useState(false);
    const [logs, setLogs] = useState([]);
    const [stats, setStats] = useState({
        metricsCount: 0,
        activitiesCount: 0,
        alertsCount: 0,
        screenshotsCount: 0,
        processesCount: 0
    });

    // Generate random device ID
    const generateDeviceId = () => {
        return `device-${Math.random().toString(36).substr(2, 9)}`;
    };

    useEffect(() => {
        if (!deviceId) {
            setDeviceId(generateDeviceId());
        }
    }, [deviceId]);

    const addLog = useCallback((message, type = 'info') => {
        const timestamp = new Date().toLocaleTimeString();
        setLogs(prev => [{ timestamp, message, type }, ...prev.slice(0, 49)]);
    }, []);

    const registerDevice = async () => {
        try {
            const payload = {
                id: deviceId,
                name: deviceName || `${deviceType}-${deviceId.slice(-4)}`,
                type: deviceType,
                os: deviceOS,
                current_user: currentUser || 'simulator-user',
                location: 'Simulated Location',
                ip_address: '192.168.1.' + Math.floor(Math.random() * 255),
                mac_address: Array.from({ length: 6 }, () =>
                    Math.floor(Math.random() * 256).toString(16).padStart(2, '0')
                ).join(':')
            };

            const response = await fetch(`${API_BASE_URL}/devices/register`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            if (response.ok) {
                setIsRegistered(true);
                addLog('✓ Device registered successfully', 'success');
            } else {
                addLog('✗ Failed to register device', 'error');
            }
        } catch (error) {
            addLog(`✗ Error: ${error.message}`, 'error');
        }
    };

    const sendMetrics = useCallback(async () => {
        try {
            const metrics = {
                cpu_usage: Math.random() * 100,
                cpu_temp: 40 + Math.random() * 40,
                memory_total: 16384,
                memory_used: Math.floor(Math.random() * 12000),
                swap_used: Math.floor(Math.random() * 2000),
                disk_total: 500000,
                disk_used: Math.floor(Math.random() * 400000),
                net_bytes_in: Math.floor(Math.random() * 1000000),
                net_bytes_out: Math.floor(Math.random() * 500000)
            };

            const response = await fetch(`${API_BASE_URL}/devices/${deviceId}/metrics`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(metrics)
            });

            if (response.ok) {
                setStats(prev => ({ ...prev, metricsCount: prev.metricsCount + 1 }));
                addLog('📊 Metrics sent', 'success');
            }
        } catch (error) {
            addLog(`✗ Metrics error: ${error.message}`, 'error');
        }
    }, [deviceId, addLog]);

    const sendActivities = useCallback(async () => {
        try {
            const apps = ['Chrome', 'VSCode', 'Slack', 'Terminal', 'Spotify', 'Zoom'];
            const types = ['app_usage', 'file_access', 'network_activity'];

            const activities = Array.from({ length: Math.floor(Math.random() * 3) + 1 }, () => ({
                type: types[Math.floor(Math.random() * types.length)],
                app: apps[Math.floor(Math.random() * apps.length)],
                description: `User activity on ${apps[Math.floor(Math.random() * apps.length)]}`,
                duration: Math.floor(Math.random() * 300)
            }));

            const response = await fetch(`${API_BASE_URL}/devices/${deviceId}/activities`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(activities)
            });

            if (response.ok) {
                setStats(prev => ({ ...prev, activitiesCount: prev.activitiesCount + activities.length }));
                addLog(`📝 ${activities.length} activities sent`, 'success');
            }
        } catch (error) {
            addLog(`✗ Activities error: ${error.message}`, 'error');
        }
    }, [deviceId, addLog]);

    const sendAlert = useCallback(async () => {
        try {
            const alertTypes = ['cpu_high', 'memory_high', 'disk_full', 'security_warning'];
            const levels = ['warning', 'critical'];

            const alerts = [{
                level: levels[Math.floor(Math.random() * levels.length)],
                type: alertTypes[Math.floor(Math.random() * alertTypes.length)],
                message: 'Simulated alert condition detected',
                value: Math.floor(Math.random() * 100),
                threshold: 80
            }];

            const response = await fetch(`${API_BASE_URL}/devices/${deviceId}/alerts`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(alerts)
            });

            if (response.ok) {
                setStats(prev => ({ ...prev, alertsCount: prev.alertsCount + 1 }));
                addLog('⚠️ Alert sent', 'warning');
            }
        } catch (error) {
            addLog(`✗ Alert error: ${error.message}`, 'error');
        }
    }, [deviceId, addLog]);

    const sendScreenshot = useCallback(async () => {
        try {
            // Create a simple colored canvas as screenshot
            const canvas = document.createElement('canvas');
            canvas.width = 800;
            canvas.height = 600;
            const ctx = canvas.getContext('2d');

            // Random background color
            ctx.fillStyle = `hsl(${Math.random() * 360}, 70%, 50%)`;
            ctx.fillRect(0, 0, 800, 600);

            // Add some text
            ctx.fillStyle = 'white';
            ctx.font = '48px Arial';
            ctx.fillText('Simulated Screenshot', 50, 300);
            ctx.font = '24px Arial';
            ctx.fillText(new Date().toLocaleString(), 50, 350);
            ctx.fillText(`Device: ${deviceId}`, 50, 400);

            canvas.toBlob(async (blob) => {
                const formData = new FormData();
                formData.append('device_id', deviceId);
                formData.append('file', blob, `screenshot-${Date.now()}.png`);

                const response = await fetch(`${API_BASE_URL}/screenshots`, {
                    method: 'POST',
                    body: formData
                });

                if (response.ok) {
                    setStats(prev => ({ ...prev, screenshotsCount: prev.screenshotsCount + 1 }));
                    addLog('📸 Screenshot uploaded', 'success');
                }
            }, 'image/png');
        } catch (error) {
            addLog(`✗ Screenshot error: ${error.message}`, 'error');
        }
    }, [deviceId, addLog]);

    const executeCommand = useCallback(async (cmd) => {
        try {
            addLog(`⚙️ Executing command: ${cmd.command}`, 'info');

            // Whitelist of allowed commands
            const allowedCommands = [
                'get_info',
                'status',
                'restart',
                'get_processes',
                'get_logs',
                'restart_service',
                'screenshot'
            ];

            const commandBase = cmd.command.toLowerCase().split(' ')[0];
            if (!allowedCommands.includes(commandBase)) {
                throw new Error('Command not allowed');
            }

            let result = '';
            let exitCode = 0;

            // Simple command execution simulation
            switch (commandBase) {
                case 'get_info':
                    result = JSON.stringify({
                        device_id: deviceId,
                        name: deviceName || `${deviceType}-${deviceId.slice(-4)}`,
                        type: deviceType,
                        os: deviceOS,
                        user: currentUser || 'simulator-user',
                        uptime: Math.floor(Math.random() * 86400),
                    });
                    break;
                case 'status':
                    result = 'Device is online and operational';
                    break;
                case 'restart':
                    result = 'Device restart initiated';
                    break;
                case 'get_processes':
                    result = JSON.stringify([
                        { name: 'chrome', cpu: 15.2, memory: 512000 },
                        { name: 'vscode', cpu: 8.5, memory: 256000 },
                        { name: 'terminal', cpu: 2.1, memory: 64000 },
                    ]);
                    break;
                case 'get_logs':
                    result = 'Log line 1\nLog line 2\nLog line 3';
                    break;
                case 'restart_service': {
                    const service = cmd.command.split(' ')[1] || 'unknown';
                    result = `Service ${service} restarted successfully`;
                    break;
                }
                case 'screenshot':
                    result = 'Screenshot captured successfully';
                    break;
                default:
                    result = `Command executed: ${cmd.command}`;
            }

            // Submit result back to backend
            const submitResponse = await fetch(`${API_BASE_URL}/commands/${cmd.id}/result`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    status: 'completed',
                    result: result,
                    exit_code: exitCode,
                }),
            });

            if (submitResponse.ok) {
                addLog(`✓ Command completed: ${cmd.command}`, 'success');
            }
        } catch (error) {
            addLog(`✗ Command error: ${error.message}`, 'error');
            // Try to report failure
            try {
                await fetch(`${API_BASE_URL}/commands/${cmd.id}/result`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        status: 'failed',
                        result: error.message,
                        exit_code: 1,
                    }),
                });
            } catch {
                // Ignore if we can't report the failure
            }
        }
    }, [deviceId, deviceName, deviceType, deviceOS, currentUser, addLog]);

    const pollCommands = useCallback(async () => {
        try {
            const response = await fetch(`${API_BASE_URL}/devices/${deviceId}/commands/pending`);
            if (response.ok) {
                const commands = await response.json();
                for (const cmd of commands) {
                    await executeCommand(cmd);
                }
            }
        } catch {
            // Silently fail command polling to avoid log spam
        }
    }, [deviceId, executeCommand]);

    const sendProcesses = useCallback(async () => {
        try {
            // Generate realistic process list
            const processNames = [
                'chrome', 'firefox', 'vscode', 'slack', 'terminal', 'spotify',
                'docker', 'node', 'python', 'postgres', 'redis', 'nginx'
            ];

            const processes = Array.from({ length: Math.floor(Math.random() * 8) + 5 }, () => ({
                pid: 1000 + Math.floor(Math.random() * 9000),
                name: processNames[Math.floor(Math.random() * processNames.length)],
                cpu: Math.random() * 50,
                memory: Math.floor(Math.random() * 1024 * 1024 * 1024), // Up to 1GB
                command: `/usr/bin/${processNames[Math.floor(Math.random() * processNames.length)]}`
            }));

            const response = await fetch(`${API_BASE_URL}/devices/${deviceId}/processes`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(processes)
            });

            if (response.ok) {
                setStats(prev => ({ ...prev, processesCount: (prev.processesCount || 0) + processes.length }));
                addLog(`💻 ${processes.length} processes sent`, 'success');
            }
        } catch (error) {
            addLog(`✗ Processes error: ${error.message}`, 'error');
        }
    }, [deviceId]);

    const startSimulation = () => {
        if (!isRegistered) {
            addLog('Please register device first', 'warning');
            return;
        }
        setIsRunning(true);
        addLog('🚀 Simulation started', 'info');
    };

    const stopSimulation = () => {
        setIsRunning(false);
        addLog('⏸️ Simulation stopped', 'info');
    };

    // Poll for commands every 5 seconds when device is registered
    useEffect(() => {
        if (!isRegistered) return;

        const interval = setInterval(() => {
            pollCommands();
        }, 5000);

        return () => clearInterval(interval);
    }, [isRegistered, pollCommands]);

    // Auto-send data when simulation is running
    useEffect(() => {
        if (!isRunning) return;

        const interval = setInterval(() => {
            sendMetrics();

            // 50% chance to send activities
            if (Math.random() > 0.5) {
                sendActivities();
            }

            // 20% chance to send alert
            if (Math.random() > 0.8) {
                sendAlert();
            }

            // 30% chance to send screenshot
            if (Math.random() > 0.7) {
                sendScreenshot();
            }

            // 40% chance to send processes
            if (Math.random() > 0.6) {
                sendProcesses();
            }
        }, 5000); // Every 5 seconds

        return () => clearInterval(interval);
    }, [isRunning, isRegistered, sendMetrics, sendActivities, sendAlert, sendScreenshot, sendProcesses]);

    return (
        <div className="simulator">
            <header className="simulator-header">
                <h1>🖥️ Device Simulator</h1>
                <p>Simulate device registration and data transmission</p>
            </header>

            <div className="simulator-content">
                {/* Device Configuration */}
                <div className="card">
                    <h2>Device Configuration</h2>
                    <div className="form-grid">
                        <div className="form-group">
                            <label>Device ID</label>
                            <input
                                type="text"
                                value={deviceId}
                                onChange={(e) => setDeviceId(e.target.value)}
                                disabled={isRegistered}
                                placeholder="device-xxxxx"
                            />
                        </div>

                        <div className="form-group">
                            <label>Device Name</label>
                            <input
                                type="text"
                                value={deviceName}
                                onChange={(e) => setDeviceName(e.target.value)}
                                disabled={isRegistered}
                                placeholder="My Device"
                            />
                        </div>

                        <div className="form-group">
                            <label>Type</label>
                            <select
                                value={deviceType}
                                onChange={(e) => setDeviceType(e.target.value)}
                                disabled={isRegistered}
                            >
                                <option value="laptop">Laptop</option>
                                <option value="desktop">Desktop</option>
                                <option value="mobile">Mobile</option>
                                <option value="tablet">Tablet</option>
                            </select>
                        </div>

                        <div className="form-group">
                            <label>Operating System</label>
                            <select
                                value={deviceOS}
                                onChange={(e) => setDeviceOS(e.target.value)}
                                disabled={isRegistered}
                            >
                                <option value="macOS">macOS</option>
                                <option value="Windows">Windows</option>
                                <option value="Linux">Linux</option>
                                <option value="iOS">iOS</option>
                                <option value="Android">Android</option>
                            </select>
                        </div>

                        <div className="form-group">
                            <label>Current User</label>
                            <input
                                type="text"
                                value={currentUser}
                                onChange={(e) => setCurrentUser(e.target.value)}
                                disabled={isRegistered}
                                placeholder="simulator-user"
                            />
                        </div>
                    </div>

                    <div className="button-group">
                        {!isRegistered ? (
                            <button className="btn btn-primary" onClick={registerDevice}>
                                Register Device
                            </button>
                        ) : (
                            <button className="btn btn-secondary" disabled>
                                ✓ Device Registered
                            </button>
                        )}

                        <button
                            className="btn btn-outline"
                            onClick={() => {
                                setDeviceId(generateDeviceId());
                                setIsRegistered(false);
                                setIsRunning(false);
                                setStats({ metricsCount: 0, activitiesCount: 0, alertsCount: 0, screenshotsCount: 0, processesCount: 0 });
                                addLog('Device reset', 'info');
                            }}
                        >
                            Reset
                        </button>
                    </div>
                </div>

                {/* Simulation Controls */}
                <div className="card">
                    <h2>Simulation Controls</h2>
                    <div className="button-group">
                        {!isRunning ? (
                            <button
                                className="btn btn-success"
                                onClick={startSimulation}
                                disabled={!isRegistered}
                            >
                                ▶️ Start Auto Simulation
                            </button>
                        ) : (
                            <button className="btn btn-danger" onClick={stopSimulation}>
                                ⏸️ Stop Simulation
                            </button>
                        )}
                    </div>

                    <div className="manual-controls">
                        <h3>Manual Actions</h3>
                        <div className="button-group">
                            <button
                                className="btn btn-small"
                                onClick={sendMetrics}
                                disabled={!isRegistered}
                            >
                                📊 Send Metrics
                            </button>
                            <button
                                className="btn btn-small"
                                onClick={sendActivities}
                                disabled={!isRegistered}
                            >
                                📝 Send Activities
                            </button>
                            <button
                                className="btn btn-small"
                                onClick={sendAlert}
                                disabled={!isRegistered}
                            >
                                ⚠️ Send Alert
                            </button>
                            <button
                                className="btn btn-small"
                                onClick={sendScreenshot}
                                disabled={!isRegistered}
                            >
                                📸 Send Screenshot
                            </button>
                        </div>
                    </div>
                </div>

                {/* Statistics */}
                <div className="card stats-card">
                    <h2>Statistics</h2>
                    <div className="stats-grid">
                        <div className="stat">
                            <div className="stat-value">{stats.metricsCount}</div>
                            <div className="stat-label">Metrics Sent</div>
                        </div>
                        <div className="stat">
                            <div className="stat-value">{stats.activitiesCount}</div>
                            <div className="stat-label">Activities Sent</div>
                        </div>
                        <div className="stat">
                            <div className="stat-value">{stats.alertsCount}</div>
                            <div className="stat-label">Alerts Sent</div>
                        </div>
                        <div className="stat">
                            <div className="stat-value">{stats.screenshotsCount}</div>
                            <div className="stat-label">Screenshots Sent</div>
                        </div>
                        <div className="stat">
                            <div className="stat-value">{stats.processesCount}</div>
                            <div className="stat-label">Processes Sent</div>
                        </div>
                    </div>
                </div>

                {/* Activity Logs */}
                <div className="card logs-card">
                    <h2>Activity Logs</h2>
                    <div className="logs">
                        {logs.length === 0 ? (
                            <p className="no-logs">No activity yet. Register a device to get started.</p>
                        ) : (
                            logs.map((log, idx) => (
                                <div key={idx} className={`log-entry log-${log.type}`}>
                                    <span className="log-time">{log.timestamp}</span>
                                    <span className="log-message">{log.message}</span>
                                </div>
                            ))
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}

export default DeviceSimulator;
