import { test, expect, vi, describe, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor, cleanup } from '@testing-library/react'
import DeviceDashboard from './DeviceDashboard'

// Mock fetch
global.fetch = vi.fn()

// Mock environment variable
vi.mock('import.meta', () => ({
    env: {
        VITE_MENTOR_API_URL: 'http://localhost:8080'
    }
}))

// Mock recharts to avoid canvas issues in tests
vi.mock('recharts', () => ({
    AreaChart: ({ children }) => <div data-testid="area-chart">{children}</div>,
    Area: () => <div data-testid="area" />,
    XAxis: () => <div data-testid="x-axis" />,
    YAxis: () => <div data-testid="y-axis" />,
    CartesianGrid: () => <div data-testid="cartesian-grid" />,
    Tooltip: () => <div data-testid="tooltip" />,
    ResponsiveContainer: ({ children }) => <div data-testid="responsive-container">{children}</div>,
    PieChart: ({ children }) => <div data-testid="pie-chart">{children}</div>,
    Pie: () => <div data-testid="pie" />,
    Cell: () => <div data-testid="cell" />
}))

const mockDevices = [
    {
        id: 'device-1',
        deviceid: 'device-1',
        device_name: 'Test Laptop',
        device_type: 'laptop',
        os: 'Windows',
        is_online: true,
        last_seen: '2024-01-01T12:00:00Z',
        current_user: 'john.doe',
        device_location: 'Office',
        ip_address: '192.168.1.100',
        mac_address: 'aa:bb:cc:dd:ee:ff'
    }
]

const mockMetrics = [
    {
        timestamp: '2024-01-01T12:00:00Z',
        cpu_usage: 45.2,
        memory_used: 8589934592,
        memory_total: 17179869184,
        disk_used: 549755813888,
        disk_total: 1099511627776,
        net_bytes_in: 1024000,
        net_bytes_out: 512000
    }
]

const mockProcesses = [
    {
        pid: 1234,
        process_name: 'chrome.exe',
        cpu: 25.5,
        memory: 536870912
    },
    {
        pid: 5678,
        process_name: 'vscode.exe',
        cpu: 15.2,
        memory: 268435456
    }
]

const mockActivities = [
    {
        activityid: 1,
        description: 'User logged in',
        activity_type: 'authentication',
        timestamp: '2024-01-01T12:00:00Z'
    },
    {
        activityid: 2,
        description: 'File accessed: document.pdf',
        activity_type: 'file_access',
        timestamp: '2024-01-01T11:30:00Z'
    }
]

const mockScreenshots = [
    {
        screenshotid: 1,
        screenshot_url: 'http://example.com/screenshot1.jpg',
        timestamp: '2024-01-01T12:00:00Z'
    },
    {
        screenshotid: 2,
        url: 'http://example.com/screenshot2.jpg',
        created_at: '2024-01-01T11:30:00Z'
    }
]

const mockCommands = [
    {
        commandid: 1,
        command_text: 'get_info',
        status: 'completed',
        result: 'System info retrieved successfully',
        created_at: '2024-01-01T12:00:00Z'
    },
    {
        commandid: 2,
        command_text: 'restart',
        status: 'pending',
        created_at: '2024-01-01T11:30:00Z'
    },
    {
        commandid: 3,
        command_text: 'update',
        status: 'failed',
        created_at: '2024-01-01T11:00:00Z'
    }
]

describe('DeviceDashboard Extended Tests', () => {
    beforeEach(() => {
        fetch.mockClear()
        fetch.mockReset()
    })

    afterEach(() => {
        cleanup()
    })

    test('displays processes tab with process list', async () => {
        fetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockDevices
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockMetrics
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockProcesses
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const laptopButton = screen.getByRole('button', { name: /test laptop/i })
        fireEvent.click(laptopButton)

        await waitFor(() => {
            const processesTab = screen.getByRole('tab', { name: /processes/i })
            expect(processesTab).toBeInTheDocument()
            fireEvent.click(processesTab)
        })

        await waitFor(() => {
            expect(screen.getByText('chrome.exe')).toBeInTheDocument()
            expect(screen.getByText('vscode.exe')).toBeInTheDocument()
            expect(screen.getByText('PID: 1234')).toBeInTheDocument()
        })
    })

    test('displays activity tab with activity list', async () => {
        fetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockDevices
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockMetrics
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockActivities
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const laptopButton = screen.getByRole('button', { name: /test laptop/i })
        fireEvent.click(laptopButton)

        await waitFor(() => {
            const activityTab = screen.getByRole('tab', { name: /activity/i })
            expect(activityTab).toBeInTheDocument()
            fireEvent.click(activityTab)
        })

        await waitFor(() => {
            expect(screen.getByText('User logged in')).toBeInTheDocument()
            expect(screen.getByText('File accessed: document.pdf')).toBeInTheDocument()
        })
    })

    test('displays screenshots tab with screenshots', async () => {
        fetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockDevices
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockMetrics
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockScreenshots
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const laptopButton = screen.getByRole('button', { name: /test laptop/i })
        fireEvent.click(laptopButton)

        await waitFor(() => {
            const screenshotsTab = screen.getByRole('tab', { name: /screenshots/i })
            expect(screenshotsTab).toBeInTheDocument()
            fireEvent.click(screenshotsTab)
        })

        await waitFor(() => {
            const images = screen.getAllByRole('img')
            expect(images.length).toBeGreaterThan(0)
        })
    })

    test('displays commands tab and allows sending commands', async () => {
        fetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockDevices
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockMetrics
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockCommands
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ success: true })
            })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const laptopButton = screen.getByRole('button', { name: /test laptop/i })
        fireEvent.click(laptopButton)

        await waitFor(() => {
            const commandsTab = screen.getByRole('tab', { name: /commands/i })
            expect(commandsTab).toBeInTheDocument()
            fireEvent.click(commandsTab)
        })

        await waitFor(() => {
            expect(screen.getByText('get_info')).toBeInTheDocument()
            expect(screen.getByText('restart')).toBeInTheDocument()
            expect(screen.getByText('update')).toBeInTheDocument()
        })

        // Test sending a command
        const commandInput = screen.getByPlaceholderText(/enter command/i)
        fireEvent.change(commandInput, { target: { value: 'test_command' } })

        const sendButton = screen.getByRole('button', { name: /send/i })
        fireEvent.click(sendButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                '/api/devices/commands',
                expect.objectContaining({
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ deviceid: 'device-1', command_text: 'test_command' })
                })
            )
        })
    })

    test('sends command on Enter key press', async () => {
        fetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockDevices
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockMetrics
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ success: true })
            })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const laptopButton = screen.getByRole('button', { name: /test laptop/i })
        fireEvent.click(laptopButton)

        await waitFor(() => {
            const commandsTab = screen.getByRole('tab', { name: /commands/i })
            fireEvent.click(commandsTab)
        })

        const commandInput = screen.getByPlaceholderText(/enter command/i)
        fireEvent.change(commandInput, { target: { value: 'test_enter_command' } })
        fireEvent.keyDown(commandInput, { key: 'Enter' })

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                '/api/devices/commands',
                expect.objectContaining({
                    method: 'POST'
                })
            )
        })
    })

    test('displays no screenshots message when screenshots array is empty', async () => {
        fetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockDevices
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockMetrics
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValue({
                ok: true,
                json: async () => []
            })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const laptopButton = screen.getByRole('button', { name: /test laptop/i })
        fireEvent.click(laptopButton)

        await waitFor(() => {
            const screenshotsTab = screen.getByRole('tab', { name: /screenshots/i })
            fireEvent.click(screenshotsTab)
        })

        await waitFor(() => {
            expect(screen.getByText('No screenshots available')).toBeInTheDocument()
        })
    })

    test('displays no commands message when commands array is empty', async () => {
        fetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockDevices
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockMetrics
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const laptopButton = screen.getByRole('button', { name: /test laptop/i })
        fireEvent.click(laptopButton)

        await waitFor(() => {
            const commandsTab = screen.getByRole('tab', { name: /commands/i })
            fireEvent.click(commandsTab)
        })

        await waitFor(() => {
            expect(screen.getByText('No commands sent yet')).toBeInTheDocument()
        })
    })

    test('refresh button triggers data refresh', async () => {
        fetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockDevices
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockMetrics
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            // Second fetch for refresh
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockMetrics
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const laptopButton = screen.getByRole('button', { name: /test laptop/i })
        fireEvent.click(laptopButton)

        await waitFor(() => {
            expect(screen.getByText('CPU Usage')).toBeInTheDocument()
        })

        const fetchCallsBeforeRefresh = fetch.mock.calls.length

        // Click refresh button
        const refreshButton = screen.getByLabelText('refresh')
        fireEvent.click(refreshButton)

        // Verify fetch was called again
        await waitFor(() => {
            expect(fetch.mock.calls.length).toBeGreaterThan(fetchCallsBeforeRefresh)
        })
    })

    test('renders different device icons based on device type', async () => {
        const devicesWithTypes = [
            { id: '1', deviceid: '1', device_name: 'Phone Device', device_type: 'phone', is_online: true },
            { id: '2', deviceid: '2', device_name: 'Tablet Device', device_type: 'tablet', is_online: true },
            { id: '3', deviceid: '3', device_name: 'Desktop Device', device_type: 'desktop', is_online: true }
        ]

        fetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => devicesWithTypes
            })
            .mockResolvedValue({
                ok: true,
                json: async () => devicesWithTypes
            })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Phone Device')).toBeInTheDocument()
            expect(screen.getByText('Tablet Device')).toBeInTheDocument()
            expect(screen.getByText('Desktop Device')).toBeInTheDocument()
        })
    })
})
