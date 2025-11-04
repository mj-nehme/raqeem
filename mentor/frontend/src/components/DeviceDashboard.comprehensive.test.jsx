import { test, expect, vi, describe, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
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
        ip_address: '192.168.1.100',
        mac_address: 'aa:bb:cc:dd:ee:ff'
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
        ip_address: '192.168.1.101',
        mac_address: '11:22:33:44:55:66'
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

const mockAlerts = [
    {
        id: 1,
        level: 'warning',
        type: 'cpu',
        message: 'High CPU usage detected',
        timestamp: '2024-01-01T12:00:00Z'
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
            json: async () => []
        })

        render(<DeviceDashboard />)
        expect(screen.getByText('Devices')).toBeInTheDocument()
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

    test('shows online/offline status chips', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => mockDevices
        })

        render(<DeviceDashboard />)

        await waitFor(() => {
            const onlineChips = screen.getAllByText('Online')
            const offlineChips = screen.getAllByText('Offline')
            expect(onlineChips.length).toBeGreaterThan(0)
            expect(offlineChips.length).toBeGreaterThan(0)
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
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockAlerts
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const laptopButton = screen.getByText('Test Laptop').closest('button')
        fireEvent.click(laptopButton)

        await waitFor(() => {
            expect(screen.getByText('john.doe')).toBeInTheDocument()
        })
    })

    test('displays metrics when device selected', async () => {
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
                json: async () => mockAlerts
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const laptopButton = screen.getByText('Test Laptop').closest('button')
        fireEvent.click(laptopButton)

        await waitFor(() => {
            expect(screen.getByText('CPU Usage')).toBeInTheDocument()
            expect(screen.getByText('45.2%')).toBeInTheDocument()
        })
    })

    test('handles API error for devices gracefully', async () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => { })
        fetch.mockRejectedValueOnce(new Error('Network error'))

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(consoleError).toHaveBeenCalled()
        })

        consoleError.mockRestore()
    })

    test('shows prompt to select device when none selected', () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => []
        })

        render(<DeviceDashboard />)

        expect(screen.getByText('Select a device to view details')).toBeInTheDocument()
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
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => mockAlerts
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const laptopButton = screen.getByText('Test Laptop').closest('button')
        fireEvent.click(laptopButton)

        await waitFor(() => {
            const alertsTab = screen.getByRole('tab', { name: /alerts/i })
            expect(alertsTab).toBeInTheDocument()
            fireEvent.click(alertsTab)
        })

        await waitFor(() => {
            expect(screen.getByText('Alerts')).toBeInTheDocument()
        })
    })

    test('displays device information section', async () => {
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
                json: async () => mockAlerts
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const laptopButton = screen.getByText('Test Laptop').closest('button')
        fireEvent.click(laptopButton)

        await waitFor(() => {
            expect(screen.getByText('Device Information')).toBeInTheDocument()
            expect(screen.getByText('192.168.1.100')).toBeInTheDocument()
            expect(screen.getByText('aa:bb:cc:dd:ee:ff')).toBeInTheDocument()
        })
    })

    test('shows alert messages when device has alerts', async () => {
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
                json: async () => mockAlerts
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => []
            })

        render(<DeviceDashboard />)

        await waitFor(() => {
            expect(screen.getByText('Test Laptop')).toBeInTheDocument()
        })

        const laptopButton = screen.getByText('Test Laptop').closest('button')
        fireEvent.click(laptopButton)

        // Switch to Alerts tab
        await waitFor(() => {
            const alertsTab = screen.getByRole('tab', { name: /alerts/i })
            fireEvent.click(alertsTab)
        })

        await waitFor(() => {
            expect(screen.getByText('High CPU usage detected')).toBeInTheDocument()
        })
    })
})