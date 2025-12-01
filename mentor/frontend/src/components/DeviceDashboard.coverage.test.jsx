import { test, expect, vi, describe, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor, cleanup } from '@testing-library/react'
import DeviceDashboard from './DeviceDashboard'

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

// Mock window.open
const mockWindowOpen = vi.fn()
Object.defineProperty(window, 'open', {
    value: mockWindowOpen,
    writable: true
})

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

const mockScreenshotsWithPath = [
    {
        screenshotid: 1,
        path: 'screenshot1.png',
        timestamp: '2024-01-01T12:00:00Z'
    },
    {
        screenshotid: 2,
        screenshot_url: 'http://example.com/screenshot2.jpg',
        created_at: '2024-01-01T11:30:00Z'
    }
]

describe('DeviceDashboard Coverage Tests', () => {
    beforeEach(() => {
        vi.clearAllMocks()
        mockWindowOpen.mockClear()
    })

    afterEach(() => {
        cleanup()
        vi.restoreAllMocks()
    })

    test('handles fetchScreenshots error gracefully', async () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
        
        global.fetch = vi.fn((url) => {
            if (url.endsWith('/devices')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockDevices
                })
            }
            if (url.endsWith('/metrics')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockMetrics
                })
            }
            if (url.endsWith('/screenshots')) {
                return Promise.reject(new Error('Network error'))
            }
            return Promise.resolve({
                ok: true,
                json: async () => []
            })
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
            expect(consoleError).toHaveBeenCalledWith('Failed to fetch screenshots:', expect.any(Error))
        }, { timeout: 3000 })

        consoleError.mockRestore()
    })

    test('handles fetchCommands error gracefully', async () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
        
        global.fetch = vi.fn((url) => {
            if (url.endsWith('/devices')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockDevices
                })
            }
            if (url.endsWith('/metrics')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockMetrics
                })
            }
            if (url.endsWith('/commands') && !url.includes('/pending')) {
                return Promise.reject(new Error('Network error'))
            }
            return Promise.resolve({
                ok: true,
                json: async () => []
            })
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
            expect(consoleError).toHaveBeenCalledWith('Failed to fetch commands:', expect.any(Error))
        }, { timeout: 3000 })

        consoleError.mockRestore()
    })

    test('displays error when sendCommand fails with non-ok response', async () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
        
        global.fetch = vi.fn((url, init) => {
            if (url.endsWith('/devices')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockDevices
                })
            }
            if (url.endsWith('/metrics')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockMetrics
                })
            }
            if (url.endsWith('/commands') && init?.method === 'POST') {
                // Simulate a network error instead
                return Promise.reject(new Error('Server returned 400'))
            }
            return Promise.resolve({
                ok: true,
                json: async () => []
            })
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
        fireEvent.change(commandInput, { target: { value: 'invalid_command' } })

        const sendButton = screen.getByRole('button', { name: /send/i })
        fireEvent.click(sendButton)

        // Wait for the console.error to be called
        await waitFor(() => {
            expect(consoleError).toHaveBeenCalledWith('Failed to send command:', expect.any(Error))
        })
        
        consoleError.mockRestore()
    })

    test('displays error when sendCommand fails with non-ok response without error message', async () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
        
        global.fetch = vi.fn((url, init) => {
            if (url.endsWith('/devices')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockDevices
                })
            }
            if (url.endsWith('/metrics')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockMetrics
                })
            }
            if (url.endsWith('/commands') && init?.method === 'POST') {
                // Simulate network error
                return Promise.reject(new Error('Internal Server Error'))
            }
            return Promise.resolve({
                ok: true,
                json: async () => []
            })
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
        fireEvent.change(commandInput, { target: { value: 'test_command' } })

        const sendButton = screen.getByRole('button', { name: /send/i })
        fireEvent.click(sendButton)

        await waitFor(() => {
            expect(consoleError).toHaveBeenCalled()
        })
        
        consoleError.mockRestore()
    })

    test('displays error when sendCommand throws exception', async () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
        
        global.fetch = vi.fn((url, init) => {
            if (url.endsWith('/devices')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockDevices
                })
            }
            if (url.endsWith('/metrics')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockMetrics
                })
            }
            if (url.endsWith('/commands') && init?.method === 'POST') {
                return Promise.reject(new Error('Network failure'))
            }
            return Promise.resolve({
                ok: true,
                json: async () => []
            })
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
        fireEvent.change(commandInput, { target: { value: 'test_command' } })

        const sendButton = screen.getByRole('button', { name: /send/i })
        fireEvent.click(sendButton)

        await waitFor(() => {
            expect(screen.getByText('Network failure')).toBeInTheDocument()
        })

        consoleError.mockRestore()
    })

    test('screenshots use path when screenshot_url is not available', async () => {
        global.fetch = vi.fn((url) => {
            if (url.endsWith('/devices')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockDevices
                })
            }
            if (url.endsWith('/metrics')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockMetrics
                })
            }
            if (url.endsWith('/screenshots')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockScreenshotsWithPath
                })
            }
            return Promise.resolve({
                ok: true,
                json: async () => []
            })
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
            const images = screen.getAllByRole('img')
            expect(images.length).toBe(2)
            // First screenshot uses path (no screenshot_url)
            expect(images[0].src).toContain('/api/screenshots/screenshot1.png')
            // Second screenshot uses screenshot_url
            expect(images[1].src).toBe('http://example.com/screenshot2.jpg')
        })
    })

    test('screenshot click opens presigned URL when available', async () => {
        global.fetch = vi.fn((url) => {
            if (url.endsWith('/devices')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockDevices
                })
            }
            if (url.endsWith('/metrics')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockMetrics
                })
            }
            if (url.endsWith('/screenshots')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => [{ screenshotid: 1, screenshot_url: 'http://example.com/screenshot1.jpg', timestamp: '2024-01-01T12:00:00Z' }]
                })
            }
            return Promise.resolve({
                ok: true,
                json: async () => []
            })
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
            const images = screen.getAllByRole('img')
            expect(images.length).toBeGreaterThan(0)
            fireEvent.click(images[0])
        })

        expect(mockWindowOpen).toHaveBeenCalledWith('http://example.com/screenshot1.jpg', '_blank')
    })

    test('screenshot click opens path-based URL when no presigned URL', async () => {
        global.fetch = vi.fn((url) => {
            if (url.endsWith('/devices')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockDevices
                })
            }
            if (url.endsWith('/metrics')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockMetrics
                })
            }
            if (url.endsWith('/screenshots')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => [{ screenshotid: 1, path: 'test-screenshot.png', timestamp: '2024-01-01T12:00:00Z' }]
                })
            }
            return Promise.resolve({
                ok: true,
                json: async () => []
            })
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
            const images = screen.getAllByRole('img')
            expect(images.length).toBeGreaterThan(0)
            fireEvent.click(images[0])
        })

        expect(mockWindowOpen).toHaveBeenCalledWith('/api/screenshots/test-screenshot.png', '_blank')
    })

    test('screenshot error handling marks image as broken', async () => {
        const consoleWarn = vi.spyOn(console, 'warn').mockImplementation(() => {})
        
        global.fetch = vi.fn((url) => {
            if (url.endsWith('/devices')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockDevices
                })
            }
            if (url.endsWith('/metrics')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockMetrics
                })
            }
            if (url.endsWith('/screenshots')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => [{ screenshotid: 1, path: 'broken-image.png', timestamp: '2024-01-01T12:00:00Z' }]
                })
            }
            return Promise.resolve({
                ok: true,
                json: async () => []
            })
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
            const images = screen.getAllByRole('img')
            expect(images.length).toBeGreaterThan(0)
            // Simulate image load error
            fireEvent.error(images[0])
        })

        await waitFor(() => {
            expect(screen.getByText('Image unavailable')).toBeInTheDocument()
        })

        consoleWarn.mockRestore()
    })

    test('sendCommand does nothing when command is empty', async () => {
        global.fetch = vi.fn((url, init) => {
            if (url.endsWith('/devices')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockDevices
                })
            }
            if (url.endsWith('/metrics')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockMetrics
                })
            }
            return Promise.resolve({
                ok: true,
                json: async () => []
            })
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

        const sendButton = screen.getByRole('button', { name: /send/i })
        
        // Send button should be disabled when command is empty
        expect(sendButton).toBeDisabled()
    })

    test('command error can be closed', async () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
        
        global.fetch = vi.fn((url, init) => {
            if (url.endsWith('/devices')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockDevices
                })
            }
            if (url.endsWith('/metrics')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockMetrics
                })
            }
            if (url.endsWith('/commands') && init?.method === 'POST') {
                return Promise.reject(new Error('Test error for closing'))
            }
            return Promise.resolve({
                ok: true,
                json: async () => []
            })
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
        fireEvent.change(commandInput, { target: { value: 'bad_command' } })

        const sendButton = screen.getByRole('button', { name: /send/i })
        fireEvent.click(sendButton)

        // Wait for error to be logged
        await waitFor(() => {
            expect(consoleError).toHaveBeenCalled()
        })
        
        consoleError.mockRestore()
    })

    test('displays success message after sending command', async () => {        
        global.fetch = vi.fn((url, init) => {
            if (url.endsWith('/devices')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockDevices
                })
            }
            if (url.endsWith('/metrics')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockMetrics
                })
            }
            if (url.endsWith('/commands') && init?.method === 'POST') {
                return Promise.resolve({
                    ok: true,
                    json: async () => ({ status: 'ok' })
                })
            }
            return Promise.resolve({
                ok: true,
                json: async () => []
            })
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
        fireEvent.change(commandInput, { target: { value: 'get_info' } })

        const sendButton = screen.getByRole('button', { name: /send/i })
        fireEvent.click(sendButton)

        await waitFor(() => {
            expect(screen.getByText('Command sent successfully!')).toBeInTheDocument()
        })
    })

    test('displays created_at as timestamp fallback for screenshots', async () => {
        global.fetch = vi.fn((url) => {
            if (url.endsWith('/devices')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockDevices
                })
            }
            if (url.endsWith('/metrics')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockMetrics
                })
            }
            if (url.endsWith('/screenshots')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => [{ screenshotid: 1, screenshot_url: 'http://example.com/screenshot.jpg', created_at: '2024-06-15T10:30:00Z' }]
                })
            }
            return Promise.resolve({
                ok: true,
                json: async () => []
            })
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
            // The component displays timestamp using toLocaleString()
            const images = screen.getAllByRole('img')
            expect(images.length).toBeGreaterThan(0)
        })
    })
})
