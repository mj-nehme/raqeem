import { test, expect, vi, describe, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
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
        name: 'Test Laptop',
        type: 'laptop',
        os: 'Windows',
        is_online: true,
        last_seen: '2024-01-01T12:00:00Z',
        current_user: 'john.doe',
        location: 'Office',
        ip_address: '192.168.1.100'
    },
    {
        id: 'device-2',
        name: 'Test Desktop',
        type: 'desktop',
        os: 'macOS',
        is_online: false,
        last_seen: '2024-01-01T10:00:00Z',
        current_user: 'jane.smith',
        location: 'Home Office',
        ip_address: '192.168.1.101'
    }
]

const mockMetrics = [
    {
        timestamp: '2024-01-01T12:00:00Z',
        cpu_usage: 45.2,
        memory_used: 8589934592,
        memory_total: 17179869184,
        disk_used: 549755813888,
        disk_total: 1099511627776
    },
    {
        timestamp: '2024-01-01T12:01:00Z',
        cpu_usage: 52.1,
        memory_used: 9663676416,
        memory_total: 17179869184,
        disk_used: 549755813888,
        disk_total: 1099511627776
    }
]

const mockAlerts = [
    {
        id: 1,
        level: 'warning',
        type: 'cpu',
        message: 'High CPU usage detected',
        timestamp: '2024-01-01T12:00:00Z',
        value: 85.5,
        threshold: 80.0
    },
    {
        id: 2,
        level: 'error',
        type: 'memory',
        message: 'Memory usage critical',
        timestamp: '2024-01-01T11:30:00Z',
        value: 95.2,
        threshold: 90.0
    }
]

describe('DeviceDashboard Component', () => {
    beforeEach(() => {
        vi.clearAllMocks()
        fetch.mockClear()
    })

    afterEach(() => {
        vi.restoreAllMocks()
    })

    test('renders Devices list heading', () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => mockDevices
        })

        render(<DeviceDashboard />)
        expect(screen.getByText('Devices')).toBeInTheDocument()
    })

    test('displays loading state initially', () => {
        fetch.mockImplementation(() => new Promise(() => { })) // Never resolves

        render(<DeviceDashboard />)
        expect(screen.getByRole('progressbar')).toBeInTheDocument()
    })

    test('loads and displays devices', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => mockDevices
        })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
            expect(screen.getByText('Test Desktop')).toBeInTheDocument()
        })
    })

    test('shows online/offline status', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => mockDevices
        })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Online')).toBeInTheDocument()
            expect(screen.getByText('Offline')).toBeInTheDocument()
        })
    })

    test('displays device details when device is selected', async () => {
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
                json: async () => mockAlerts
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

        const laptopItem = screen.getByText('Test Laptop')
        fireEvent.click(laptopItem)

        await waitFor(() => {
            expect(screen.getByText('Device Details')).toBeInTheDocument()
            expect(screen.getByText('john.doe')).toBeInTheDocument()
            expect(screen.getByText('Office')).toBeInTheDocument()
        })
    })

    test('shows device type icons correctly', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => mockDevices
        })

        render(<DeviceDashboard />)

        await waitFor(() => {
            // Should render icons for laptop and desktop types
            expect(screen.getByTestId('LaptopIcon') || screen.getByTestId('ComputerIcon')).toBeInTheDocument()
        })
    })

    test('displays metrics charts when device selected', async () => {
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
                json: async () => mockAlerts
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

        const laptopItem = screen.getByText('Test Laptop')
        fireEvent.click(laptopItem)

        await waitFor(() => {
            expect(screen.getByTestId('area-chart')).toBeInTheDocument()
        })
    })

    test('shows alerts for selected device', async () => {
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
                json: async () => mockAlerts
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

        const laptopItem = screen.getByText('Test Laptop')
        fireEvent.click(laptopItem)

        await waitFor(() => {
            expect(screen.getByText('High CPU usage detected')).toBeInTheDocument()
            expect(screen.getByText('Memory usage critical')).toBeInTheDocument()
        })
    })

    test('handles API error for devices', async () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => { })
        fetch.mockRejectedValueOnce(new Error('Failed to fetch'))

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(consoleError).toHaveBeenCalledWith('Failed to fetch devices:', expect.any(Error))
        })

        consoleError.mockRestore()
    })

    test('handles API error for metrics', async () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => { })
        fetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockDevices
            })
            .mockRejectedValueOnce(new Error('Failed to fetch metrics'))

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const laptopItem = screen.getByText('Test Laptop')
        fireEvent.click(laptopItem)

        await waitFor(() => {
            expect(consoleError).toHaveBeenCalledWith('Failed to fetch metrics:', expect.any(Error))
        })

        consoleError.mockRestore()
    })

    test('refreshes devices when refresh button clicked', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => mockDevices
        })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const refreshButton = screen.getByLabelText(/refresh/i) || screen.getByTestId('RefreshIcon').closest('button')
        fireEvent.click(refreshButton)

        expect(fetch).toHaveBeenCalledTimes(2) // Initial load + refresh
    })

    test('sends remote command to device', async () => {
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
                json: async () => mockAlerts
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

        const laptopItem = screen.getByText('Test Laptop')
        fireEvent.click(laptopItem)

        await waitFor(() => {
            expect(screen.getByLabelText(/command/i)).toBeInTheDocument()
        })

        const commandInput = screen.getByLabelText(/command/i)
        const sendButton = screen.getByRole('button', { name: /send/i })

        fireEvent.change(commandInput, { target: { value: 'ls -la' } })
        fireEvent.click(sendButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                'http://localhost:8080/devices/device-1/commands',
                expect.objectContaining({
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ command: 'ls -la' })
                })
            )
        })
    })

    test('filters devices by search term', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => mockDevices
        })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
            expect(screen.getByText('Test Desktop')).toBeInTheDocument()
        })

        const searchInput = screen.getByPlaceholderText(/search devices/i)
        fireEvent.change(searchInput, { target: { value: 'laptop' } })

        expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        expect(screen.queryByText('Test Desktop')).not.toBeInTheDocument()
    })

    test('displays memory and disk usage percentages', async () => {
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
                json: async () => mockAlerts
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

        const laptopItem = screen.getByText('Test Laptop')
        fireEvent.click(laptopItem)

        await waitFor(() => {
            // Memory: 8GB / 16GB = 50%
            // Disk: 512GB / 1TB = 50%
            expect(screen.getByText(/50%/)).toBeInTheDocument()
        })
    })

    test('switches between dashboard tabs', async () => {
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
                json: async () => mockAlerts
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

        const laptopItem = screen.getByText('Test Laptop')
        fireEvent.click(laptopItem)

        await waitFor(() => {
            const alertsTab = screen.getByRole('tab', { name: /alerts/i })
            fireEvent.click(alertsTab)
            expect(screen.getByText('High CPU usage detected')).toBeInTheDocument()
        })
    })

    test('formats timestamps correctly', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => mockDevices
        })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        // Should format the last_seen timestamp
        expect(screen.getByText(/Jan 1, 2024/)).toBeInTheDocument()
    })

    test('shows alert severity indicators', async () => {
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
                json: async () => mockAlerts
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

        const laptopItem = screen.getByText('Test Laptop')
        fireEvent.click(laptopItem)

        await waitFor(() => {
            // Should show warning and error chips
            expect(screen.getByText('warning')).toBeInTheDocument()
            expect(screen.getByText('error')).toBeInTheDocument()
        })
    })

    test('handles empty device list', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => []
        })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('No devices found')).toBeInTheDocument()
        })
    })

    test('auto-refreshes device list', async () => {
        vi.useFakeTimers()

        fetch.mockResolvedValue({
            ok: true,
            json: async () => mockDevices
        })

        render(<DeviceDashboard />)

        // Initial load
        expect(fetch).toHaveBeenCalledTimes(1)

        // Fast-forward 30 seconds
        vi.advanceTimersByTime(30000)

        expect(fetch).toHaveBeenCalledTimes(2)

        vi.useRealTimers()
    })

    test('cleans up intervals on unmount', () => {
        vi.useFakeTimers()
        const clearIntervalSpy = vi.spyOn(global, 'clearInterval')

        fetch.mockResolvedValue({
            ok: true,
            json: async () => mockDevices
        })

        const { unmount } = render(<DeviceDashboard />)
        unmount()

        expect(clearIntervalSpy).toHaveBeenCalled()

        vi.useRealTimers()
        clearIntervalSpy.mockRestore()
    })
})