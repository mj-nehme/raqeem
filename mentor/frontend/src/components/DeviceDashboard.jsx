import React, { useEffect, useMemo, useState } from 'react';
import {
    Box,
    Container,
    Grid,
    Card,
    CardContent,
    Typography,
    List,
    ListItem,
    ListItemButton,
    ListItemText,
    ListItemIcon,
    Chip,
    TextField,
    Button,
    Paper,
    Divider,
    Avatar,
    IconButton,
    Alert,
    CircularProgress,
    Tab,
    Tabs,
} from '@mui/material';
import {
    Computer,
    Laptop,
    PhoneAndroid,
    Tablet,
    Circle,
    Memory,
    Storage,
    NetworkCheck,
    Warning,
    Error,
    Info,
    Refresh,
    Send,
    History,
} from '@mui/icons-material';
import {
    AreaChart,
    Area,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip as RechartsTooltip,
    ResponsiveContainer,
    PieChart,
    Pie,
    Cell,
} from 'recharts';

const BACKEND_URL = 'http://localhost:30081'; // Direct to mentor backend
const COLORS = ['#1976d2', '#4caf50', '#ff9800', '#f44336', '#9c27b0'];

function getDeviceIcon(type) {
    const iconType = (type || '').toLowerCase();
    if (iconType.includes('laptop')) return <Laptop />;
    if (iconType.includes('phone') || iconType.includes('mobile')) return <PhoneAndroid />;
    if (iconType.includes('tablet')) return <Tablet />;
    return <Computer />;
}

export default function DeviceDashboard() {
    const [devices, setDevices] = useState([]);
    const [selectedDevice, setSelectedDevice] = useState(null);
    const [metrics, setMetrics] = useState([]);
    const [processes, setProcesses] = useState([]);
    const [activities, setActivities] = useState([]);
    const [alerts, setAlerts] = useState([]);
    const [screenshots, setScreenshots] = useState([]);
    const [commands, setCommands] = useState([]);
    const [command, setCommand] = useState('');
    const [loading, setLoading] = useState(false);
    const [refreshing, setRefreshing] = useState(false);
    const [tabValue, setTabValue] = useState(0);

    // Poll devices list
    useEffect(() => {
        let cancelled = false;
        const fetchDevices = async () => {
            try {
                const res = await fetch(`${BACKEND_URL}/devices`);
                const data = await res.json();
                if (!cancelled) setDevices(Array.isArray(data) ? data : []);
            } catch (err) {
                console.error('Failed to fetch devices:', err);
            }
        };
        fetchDevices();
        const interval = setInterval(fetchDevices, 10000);
        return () => {
            cancelled = true;
            clearInterval(interval);
        };
    }, []);

    // Fetch selected device details
    const fetchDeviceDetails = async () => {
        if (!selectedDevice) return;
        setRefreshing(true);
        setLoading(true);
        try {
            const [metricsRes, processesRes, activitiesRes, alertsRes, screenshotsRes, commandsRes] = await Promise.all([
                fetch(`${BACKEND_URL}/devices/${selectedDevice.id}/metrics`),
                fetch(`${BACKEND_URL}/devices/${selectedDevice.id}/processes`),
                fetch(`${BACKEND_URL}/devices/${selectedDevice.id}/activities`),
                fetch(`${BACKEND_URL}/devices/${selectedDevice.id}/alerts`),
                fetch(`${BACKEND_URL}/devices/${selectedDevice.id}/screenshots`),
                fetch(`${BACKEND_URL}/devices/${selectedDevice.id}/commands`),
            ]);
            const [metricsData, processesData, activitiesData, alertsData, screenshotsData, commandsData] = await Promise.all([
                metricsRes.json(),
                processesRes.json(),
                activitiesRes.json(),
                alertsRes.json(),
                screenshotsRes.json(),
                commandsRes.json(),
            ]);
            setMetrics(Array.isArray(metricsData) ? metricsData.slice(-50) : []);
            setProcesses(Array.isArray(processesData) ? processesData : []);
            setActivities(Array.isArray(activitiesData) ? activitiesData : []);
            setAlerts(Array.isArray(alertsData) ? alertsData : []);
            setScreenshots(Array.isArray(screenshotsData) ? screenshotsData : []);
            setCommands(Array.isArray(commandsData) ? commandsData : []);
        } catch (err) {
            console.error('Failed to fetch device details:', err);
        } finally {
            setRefreshing(false);
            setLoading(false);
        }
    };

    useEffect(() => {
        if (!selectedDevice) return;
        fetchDeviceDetails();
        const interval = setInterval(fetchDeviceDetails, 10000);
        return () => clearInterval(interval);
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [selectedDevice?.id]);

    const latestMetrics = metrics.length ? metrics[metrics.length - 1] : null;
    const chartData = useMemo(
        () =>
            metrics.map((m, idx) => ({
                name: idx,
                cpu: Number(m.cpu_usage || 0),
                memory: m.memory_total ? Number(((m.memory_used / m.memory_total) * 100).toFixed(1)) : 0,
            })),
        [metrics]
    );
    const processChartData = useMemo(
        () =>
            processes.slice(0, 5).map((p) => ({
                name: String(p.name || '').substring(0, 15),
                cpu: Number(p.cpu || 0),
            })),
        [processes]
    );

    const sendCommand = async () => {
        if (!command.trim() || !selectedDevice) return;
        try {
            await fetch(`${BACKEND_URL}/devices/commands`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ device_id: selectedDevice.id, command }),
            });
            setCommand('');
        } catch (err) {
            console.error('Failed to send command:', err);
        }
    };

    return (
        <Box sx={{ display: 'flex', minHeight: '100vh', bgcolor: 'background.default' }}>
            {/* Sidebar */}
            <Paper elevation={0} sx={{ width: 300, borderRight: '1px solid', borderColor: 'divider', display: 'flex', flexDirection: 'column' }}>
                <Box sx={{ p: 2 }}>
                    <Typography variant="h6" fontWeight={600}>
                        Devices
                    </Typography>
                </Box>
                <Divider />
                <Box sx={{ flexGrow: 1, overflowY: 'auto' }}>
                    <List>
                        {devices.map((device) => {
                            const isOnline = device.isOnline || device.is_online;
                            const isSelected = selectedDevice?.id === device.id;
                            return (
                                <ListItem key={device.id} disablePadding>
                                    <ListItemButton selected={isSelected} onClick={() => setSelectedDevice(device)} sx={{ '&.Mui-selected': { bgcolor: 'action.selected' } }}>
                                        <ListItemIcon>
                                            <Avatar sx={{ width: 32, height: 32, bgcolor: isOnline ? 'success.main' : 'error.main' }}>{getDeviceIcon(device.type || device.device_type)}</Avatar>
                                        </ListItemIcon>
                                        <ListItemText
                                            primary={device.name}
                                            secondary={<Chip size="small" label={isOnline ? 'Online' : 'Offline'} color={isOnline ? 'success' : 'error'} sx={{ height: 20, fontSize: '0.7rem' }} />}
                                        />
                                    </ListItemButton>
                                </ListItem>
                            );
                        })}
                    </List>
                </Box>
            </Paper>

            {/* Main content */}
            <Box sx={{ flexGrow: 1, overflowY: 'auto' }}>
                <Container maxWidth="xl" sx={{ py: 4 }}>
                    {selectedDevice ? (
                        <>
                            {/* Header */}
                            <Box sx={{ mb: 3, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                                <Box>
                                    <Typography variant="h4" fontWeight={600} gutterBottom>
                                        {selectedDevice.name}
                                    </Typography>
                                    <Box sx={{ display: 'flex', gap: 1, alignItems: 'center', flexWrap: 'wrap' }}>
                                        <Chip icon={<Circle sx={{ fontSize: 12 }} />} label={(selectedDevice.isOnline || selectedDevice.is_online) ? 'Online' : 'Offline'} color={(selectedDevice.isOnline || selectedDevice.is_online) ? 'success' : 'error'} size="small" />
                                        <Chip label={selectedDevice.os || 'Unknown OS'} size="small" variant="outlined" />
                                        <Chip label={selectedDevice.type || selectedDevice.device_type || 'Unknown Type'} size="small" variant="outlined" />
                                    </Box>
                                </Box>
                                <IconButton onClick={fetchDeviceDetails} disabled={refreshing} aria-label="refresh">
                                    <Refresh sx={{ animation: refreshing ? 'spin 1s linear infinite' : 'none' }} />
                                </IconButton>
                            </Box>

                            {loading ? (
                                <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
                                    <CircularProgress />
                                </Box>
                            ) : (
                                <>
                                    {/* Stats */}
                                    <Grid container spacing={3} sx={{ mb: 3 }}>
                                        {latestMetrics && (
                                            <>
                                                <Grid item xs={12} sm={6} md={3}>
                                                    <Card><CardContent>
                                                        <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                                                            <Avatar sx={{ bgcolor: 'primary.main', mr: 2 }}><Memory /></Avatar>
                                                            <Typography variant="body2" color="text.secondary">CPU Usage</Typography>
                                                        </Box>
                                                        <Typography variant="h4" fontWeight={600}>{Number(latestMetrics.cpu_usage || 0).toFixed(1)}%</Typography>
                                                    </CardContent></Card>
                                                </Grid>
                                                <Grid item xs={12} sm={6} md={3}>
                                                    <Card><CardContent>
                                                        <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                                                            <Avatar sx={{ bgcolor: 'success.main', mr: 2 }}><Memory /></Avatar>
                                                            <Typography variant="body2" color="text.secondary">Memory</Typography>
                                                        </Box>
                                                        <Typography variant="h4" fontWeight={600}>
                                                            {latestMetrics.memory_total ? ((latestMetrics.memory_used / latestMetrics.memory_total) * 100).toFixed(1) : '0.0'}%
                                                        </Typography>
                                                    </CardContent></Card>
                                                </Grid>
                                                <Grid item xs={12} sm={6} md={3}>
                                                    <Card><CardContent>
                                                        <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                                                            <Avatar sx={{ bgcolor: 'warning.main', mr: 2 }}><Storage /></Avatar>
                                                            <Typography variant="body2" color="text.secondary">Disk</Typography>
                                                        </Box>
                                                        <Typography variant="h4" fontWeight={600}>
                                                            {latestMetrics.disk_total ? ((latestMetrics.disk_used / latestMetrics.disk_total) * 100).toFixed(1) : '0.0'}%
                                                        </Typography>
                                                    </CardContent></Card>
                                                </Grid>
                                                <Grid item xs={12} sm={6} md={3}>
                                                    <Card><CardContent>
                                                        <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                                                            <Avatar sx={{ bgcolor: 'info.main', mr: 2 }}><NetworkCheck /></Avatar>
                                                            <Typography variant="body2" color="text.secondary">Network</Typography>
                                                        </Box>
                                                        <Typography variant="body1" fontWeight={600}>↓ {Number((latestMetrics.net_bytes_in || 0) / 1024 / 1024).toFixed(2)} MB/s</Typography>
                                                        <Typography variant="body1" fontWeight={600}>↑ {Number((latestMetrics.net_bytes_out || 0) / 1024 / 1024).toFixed(2)} MB/s</Typography>
                                                    </CardContent></Card>
                                                </Grid>
                                            </>
                                        )}
                                    </Grid>

                                    {/* Charts */}
                                    {metrics.length > 0 && (
                                        <Grid container spacing={3} sx={{ mb: 3 }}>
                                            <Grid item xs={12} md={8}>
                                                <Card><CardContent>
                                                    <Typography variant="h6" gutterBottom>System Performance</Typography>
                                                    <ResponsiveContainer width="100%" height={300}>
                                                        <AreaChart data={chartData}>
                                                            <defs>
                                                                <linearGradient id="colorCpu" x1="0" y1="0" x2="0" y2="1">
                                                                    <stop offset="5%" stopColor="#1976d2" stopOpacity={0.8} />
                                                                    <stop offset="95%" stopColor="#1976d2" stopOpacity={0} />
                                                                </linearGradient>
                                                                <linearGradient id="colorMemory" x1="0" y1="0" x2="0" y2="1">
                                                                    <stop offset="5%" stopColor="#4caf50" stopOpacity={0.8} />
                                                                    <stop offset="95%" stopColor="#4caf50" stopOpacity={0} />
                                                                </linearGradient>
                                                            </defs>
                                                            <CartesianGrid strokeDasharray="3 3" />
                                                            <XAxis dataKey="name" />
                                                            <YAxis domain={[0, 100]} />
                                                            <RechartsTooltip />
                                                            <Area type="monotone" dataKey="cpu" stroke="#1976d2" fillOpacity={1} fill="url(#colorCpu)" name="CPU %" />
                                                            <Area type="monotone" dataKey="memory" stroke="#4caf50" fillOpacity={1} fill="url(#colorMemory)" name="Memory %" />
                                                        </AreaChart>
                                                    </ResponsiveContainer>
                                                </CardContent></Card>
                                            </Grid>
                                            <Grid item xs={12} md={4}>
                                                <Card><CardContent>
                                                    <Typography variant="h6" gutterBottom>Top Processes (CPU)</Typography>
                                                    <ResponsiveContainer width="100%" height={300}>
                                                        <PieChart>
                                                            <Pie data={processChartData} cx="50%" cy="50%" labelLine={false} label={(entry) => entry.name} outerRadius={80} fill="#8884d8" dataKey="cpu">
                                                                {processChartData.map((entry, index) => (
                                                                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                                                                ))}
                                                            </Pie>
                                                            <RechartsTooltip />
                                                        </PieChart>
                                                    </ResponsiveContainer>
                                                </CardContent></Card>
                                            </Grid>
                                        </Grid>
                                    )}

                                    {/* Device Information */}
                                    <Card sx={{ mb: 3 }}>
                                        <CardContent>
                                            <Typography variant="h6" gutterBottom>Device Information</Typography>
                                            <Grid container spacing={2} sx={{ mt: 1 }}>
                                                <Grid item xs={12} sm={6} md={3}><Typography variant="caption" color="text.secondary">Device ID</Typography><Typography variant="body2" fontWeight={500}>{selectedDevice.id}</Typography></Grid>
                                                <Grid item xs={12} sm={6} md={3}><Typography variant="caption" color="text.secondary">MAC Address</Typography><Typography variant="body2" fontWeight={500}>{selectedDevice.mac_address || selectedDevice.mac || '—'}</Typography></Grid>
                                                <Grid item xs={12} sm={6} md={3}><Typography variant="caption" color="text.secondary">IP Address</Typography><Typography variant="body2" fontWeight={500}>{selectedDevice.ip_address || selectedDevice.ip || '—'}</Typography></Grid>
                                                <Grid item xs={12} sm={6} md={3}><Typography variant="caption" color="text.secondary">Current User</Typography><Typography variant="body2" fontWeight={500}>{selectedDevice.current_user || selectedDevice.user || '—'}</Typography></Grid>
                                                <Grid item xs={12} sm={6} md={3}><Typography variant="caption" color="text.secondary">Location</Typography><Typography variant="body2" fontWeight={500}>{selectedDevice.location || '—'}</Typography></Grid>
                                                <Grid item xs={12} sm={6} md={3}><Typography variant="caption" color="text.secondary">Last Seen</Typography><Typography variant="body2" fontWeight={500}>{selectedDevice.last_seen ? new Date(selectedDevice.last_seen).toLocaleString() : '—'}</Typography></Grid>
                                            </Grid>
                                        </CardContent>
                                    </Card>

                                    {/* Tabs */}
                                    <Card>
                                        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
                                            <Tabs value={tabValue} onChange={(e, v) => setTabValue(v)}>
                                                <Tab label="Processes" />
                                                <Tab label="Activity" />
                                                <Tab label="Alerts" />
                                                <Tab label="Screenshots" />
                                                <Tab label="Commands" />
                                            </Tabs>
                                        </Box>
                                        <CardContent>
                                            {tabValue === 0 && (
                                                <Box>
                                                    <Typography variant="h6" gutterBottom>Running Processes</Typography>
                                                    {processes.length === 0 ? (
                                                        <Typography variant="body2" color="text.secondary" sx={{ py: 2 }}>No processes to display</Typography>
                                                    ) : (
                                                        <List>
                                                            {processes.slice(0, 15).map((proc) => (
                                                                <ListItem key={proc.pid} divider>
                                                                    <ListItemText primary={proc.name} secondary={`PID: ${proc.pid}`} />
                                                                    <Box sx={{ display: 'flex', gap: 2 }}>
                                                                        <Chip label={`CPU: ${Number(proc.cpu || 0).toFixed(1)}%`} size="small" />
                                                                        <Chip label={`MEM: ${Number((proc.memory || 0) / 1024 / 1024).toFixed(1)}MB`} size="small" />
                                                                    </Box>
                                                                </ListItem>
                                                            ))}
                                                        </List>
                                                    )}
                                                </Box>
                                            )}
                                            {tabValue === 1 && (
                                                <Box>
                                                    <Typography variant="h6" gutterBottom>Recent Activity</Typography>
                                                    {activities.length === 0 ? (
                                                        <Typography variant="body2" color="text.secondary" sx={{ py: 2 }}>No activity yet</Typography>
                                                    ) : (
                                                        <List>
                                                            {activities.slice(0, 20).map((activity) => (
                                                                <ListItem key={activity.id} divider>
                                                                    <ListItemIcon><History /></ListItemIcon>
                                                                    <ListItemText primary={activity.description} secondary={`${activity.type} • ${new Date(activity.timestamp).toLocaleString()}`} />
                                                                </ListItem>
                                                            ))}
                                                        </List>
                                                    )}
                                                </Box>
                                            )}
                                            {tabValue === 2 && (
                                                <Box>
                                                    <Typography variant="h6" gutterBottom>Alerts</Typography>
                                                    {alerts.length === 0 ? (
                                                        <Typography variant="body2" color="text.secondary" sx={{ py: 2 }}>No alerts found</Typography>
                                                    ) : (
                                                        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                                                            {alerts.slice(0, 20).map((alert) => {
                                                                const severity = alert.level === 'critical' || alert.level === 'error' ? 'error' : alert.level === 'warning' ? 'warning' : 'info';
                                                                return (
                                                                    <Alert key={alert.id} severity={severity} icon={severity === 'error' ? <Error /> : severity === 'warning' ? <Warning /> : <Info />}>
                                                                        <Typography variant="body2" fontWeight={500}>{alert.message}</Typography>
                                                                        <Typography variant="caption" color="text.secondary">{new Date(alert.timestamp).toLocaleString()}</Typography>
                                                                    </Alert>
                                                                );
                                                            })}
                                                        </Box>
                                                    )}
                                                </Box>
                                            )}
                                            {tabValue === 3 && (
                                                <Box>
                                                    <Typography variant="h6" gutterBottom>Screenshots</Typography>
                                                    <Grid container spacing={2}>
                                                        {screenshots.length > 0 ? (
                                                            screenshots.slice(0, 12).map((screenshot) => (
                                                                <Grid item xs={12} sm={6} md={4} key={screenshot.id}>
                                                                    <Card>
                                                                        <Box component="img" src={screenshot.screenshot_url || screenshot.url} alt={`Screenshot ${screenshot.id}`} sx={{ width: '100%', height: 200, objectFit: 'cover', cursor: 'pointer' }} onClick={() => window.open(screenshot.screenshot_url || screenshot.url, '_blank')} onError={(e) => { e.target.style.display = 'none'; }} />
                                                                        <CardContent>
                                                                            <Typography variant="caption" color="text.secondary">{new Date(screenshot.timestamp || screenshot.created_at).toLocaleString()}</Typography>
                                                                        </CardContent>
                                                                    </Card>
                                                                </Grid>
                                                            ))
                                                        ) : (
                                                            <Grid item xs={12}>
                                                                <Typography variant="body2" color="text.secondary" align="center" sx={{ py: 4 }}>No screenshots available</Typography>
                                                            </Grid>
                                                        )}
                                                    </Grid>
                                                </Box>
                                            )}
                                            {tabValue === 4 && (
                                                <Box>
                                                    <Typography variant="h6" gutterBottom>Commands</Typography>
                                                    <Box sx={{ display: 'flex', gap: 2, mt: 2, mb: 3 }}>
                                                        <TextField 
                                                            fullWidth 
                                                            placeholder="Enter command (e.g., get_info, status, restart)..." 
                                                            value={command} 
                                                            onChange={(e) => setCommand(e.target.value)} 
                                                            onKeyPress={(e) => { if (e.key === 'Enter') sendCommand(); }} 
                                                        />
                                                        <Button variant="contained" startIcon={<Send />} onClick={sendCommand} disabled={!command.trim()}>
                                                            Send
                                                        </Button>
                                                    </Box>
                                                    
                                                    <Typography variant="subtitle2" gutterBottom sx={{ mt: 2 }}>
                                                        Command History
                                                    </Typography>
                                                    {commands.length === 0 ? (
                                                        <Typography variant="body2" color="text.secondary" sx={{ py: 2 }}>
                                                            No commands sent yet
                                                        </Typography>
                                                    ) : (
                                                        <List>
                                                            {commands.map((cmd) => (
                                                                <ListItem key={cmd.id} divider>
                                                                    <ListItemIcon>
                                                                        {cmd.status === 'completed' ? (
                                                                            <Info color="success" />
                                                                        ) : cmd.status === 'failed' ? (
                                                                            <Error color="error" />
                                                                        ) : cmd.status === 'pending' ? (
                                                                            <CircularProgress size={20} />
                                                                        ) : (
                                                                            <History />
                                                                        )}
                                                                    </ListItemIcon>
                                                                    <ListItemText
                                                                        primary={cmd.command}
                                                                        secondary={
                                                                            <>
                                                                                <Typography component="span" variant="body2" color="text.primary">
                                                                                    Status: {cmd.status}
                                                                                </Typography>
                                                                                {' • '}
                                                                                {new Date(cmd.created_at).toLocaleString()}
                                                                                {cmd.result && cmd.status === 'completed' && (
                                                                                    <>
                                                                                        <br />
                                                                                        <Typography component="span" variant="caption" color="text.secondary">
                                                                                            Result: {cmd.result.substring(0, 100)}
                                                                                            {cmd.result.length > 100 ? '...' : ''}
                                                                                        </Typography>
                                                                                    </>
                                                                                )}
                                                                            </>
                                                                        }
                                                                    />
                                                                    <Chip
                                                                        label={cmd.status}
                                                                        size="small"
                                                                        color={
                                                                            cmd.status === 'completed' ? 'success' : 
                                                                            cmd.status === 'failed' ? 'error' : 
                                                                            cmd.status === 'pending' ? 'warning' : 'default'
                                                                        }
                                                                    />
                                                                </ListItem>
                                                            ))}
                                                        </List>
                                                    )}
                                                </Box>
                                            )}
                                        </CardContent>
                                    </Card>
                                </>
                            )}
                        </>
                    ) : (
                        <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '60vh' }}>
                            <Computer sx={{ fontSize: 80, color: 'text.secondary', mb: 2 }} />
                            <Typography variant="h5" color="text.secondary">Select a device to view details</Typography>
                        </Box>
                    )}
                </Container>
            </Box>

            <style>{`@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }`}</style>
        </Box>
    );
}
