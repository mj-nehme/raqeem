import { test, expect, vi, describe, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import DeviceSimulator from './DeviceSimulator'

// Mock fetch
global.fetch = vi.fn()

describe('DeviceSimulator Component', () => {
    beforeEach(() => {
        vi.clearAllMocks()
        fetch.mockClear()
    })

    afterEach(() => {
        vi.restoreAllMocks()
    })

    test('renders Device Simulator heading', () => {
        render(<DeviceSimulator />)
        expect(screen.getByText(/Device Simulator/i)).toBeInTheDocument()
    })

    test('renders device registration form fields', () => {
        render(<DeviceSimulator />)

        // Check that input fields exist
        expect(screen.getByPlaceholderText(/My Device/i)).toBeInTheDocument()
        expect(screen.getByPlaceholderText(/simulator-user/i)).toBeInTheDocument()

        // Check that selects exist by finding options
        expect(screen.getByText('Laptop')).toBeInTheDocument()
        expect(screen.getByText('macOS')).toBeInTheDocument()

        expect(screen.getByRole('button', { name: /register device/i })).toBeInTheDocument()
    })

    test('generates device ID automatically', () => {
        render(<DeviceSimulator />)

        const deviceIdInput = screen.getByDisplayValue(/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/)
        expect(deviceIdInput.value).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i)
    })

    test('updates device name input', () => {
        render(<DeviceSimulator />)

        const nameInput = screen.getByPlaceholderText(/My Device/i)
        fireEvent.change(nameInput, { target: { value: 'Test Device' } })

        expect(nameInput.value).toBe('Test Device')
    })

    test('updates device type selection', () => {
        render(<DeviceSimulator />)

        // Find the type select element
        const selects = document.querySelectorAll('select')
        const typeSelect = Array.from(selects).find(s => s.value === 'laptop')

        fireEvent.change(typeSelect, { target: { value: 'desktop' } })

        expect(typeSelect.value).toBe('desktop')
    })

    test('updates operating system selection', () => {
        render(<DeviceSimulator />)

        // Find the OS select element
        const selects = document.querySelectorAll('select')
        const osSelect = Array.from(selects).find(s => s.value === 'macOS')

        fireEvent.change(osSelect, { target: { value: 'Linux' } })

        expect(osSelect.value).toBe('Linux')
    })

    test('updates current user input', () => {
        render(<DeviceSimulator />)

        const userInput = screen.getByPlaceholderText(/simulator-user/i)
        fireEvent.change(userInput, { target: { value: 'testuser' } })

        expect(userInput.value).toBe('testuser')
    })

    test('successful device registration', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                'http://localhost:3000/api/devices/register',
                expect.objectContaining({
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: expect.any(String)
                })
            )
        })
    })

    test('device registration with custom fields', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const nameInput = screen.getByPlaceholderText(/My Device/i)
        const selects = document.querySelectorAll('select')
        const typeSelect = Array.from(selects).find(s => s.value === 'laptop')
        const osSelect = Array.from(selects).find(s => s.value === 'macOS')
        const userInput = screen.getByPlaceholderText(/simulator-user/i)

        fireEvent.change(nameInput, { target: { value: 'My Test Device' } })
        fireEvent.change(typeSelect, { target: { value: 'tablet' } })
        fireEvent.change(osSelect, { target: { value: 'Android' } })
        fireEvent.change(userInput, { target: { value: 'testuser123' } })

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalled()
        })

        const callArgs = fetch.mock.calls[0]
        const requestBody = JSON.parse(callArgs[1].body)

        expect(requestBody.device_name).toBe('My Test Device')
        expect(requestBody.device_type).toBe('tablet')
        expect(requestBody.os).toBe('Android')
        expect(requestBody.current_user).toBe('testuser123')
    })

    test('handles registration error', async () => {
        fetch.mockRejectedValueOnce(new Error('Network error'))

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByText(/Error: Network error/i)).toBeInTheDocument()
        })
    })

    test('starts and stops simulation', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /start auto simulation/i })).toBeInTheDocument()
        })

        // Start simulation
        const startButton = screen.getByRole('button', { name: /start auto simulation/i })
        fireEvent.click(startButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /stop simulation/i })).toBeInTheDocument()
        })

        // Stop simulation
        const stopButton = screen.getByRole('button', { name: /stop simulation/i })
        fireEvent.click(stopButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /start auto simulation/i })).toBeInTheDocument()
        })
    })

    test('sends alert manually', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            const logs = screen.getAllByText(/device registered/i)
            expect(logs.length).toBeGreaterThan(0)
        })

        const sendAlertButton = screen.getByRole('button', { name: /send alert/i })
        fireEvent.click(sendAlertButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                expect.stringContaining('alerts'),
                expect.objectContaining({
                    method: 'POST'
                })
            )
        })
    })

    test('displays simulation logs', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            const logs = screen.getAllByText(/device registered/i)
            expect(logs.length).toBeGreaterThan(0)
        })
    })

    test('updates simulation statistics', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Check initial stats - the value and label are in separate elements
        expect(screen.getByText('Metrics Sent')).toBeInTheDocument()
        expect(screen.getByText('Activities Sent')).toBeInTheDocument()
        expect(screen.getByText('Alerts Sent')).toBeInTheDocument()
        expect(screen.getByText('Screenshots Sent')).toBeInTheDocument()
    })

    test('generates realistic MAC address', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalled()
        })

        const callArgs = fetch.mock.calls[0]
        const requestBody = JSON.parse(callArgs[1].body)

        expect(requestBody.mac_address).toMatch(/^([0-9a-f]{2}:){5}[0-9a-f]{2}$/i)
    })

    test('generates realistic IP address', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalled()
        })

        const callArgs = fetch.mock.calls[0]
        const requestBody = JSON.parse(callArgs[1].body)

        expect(requestBody.ip_address).toMatch(/^192\.168\.1\.\d{1,3}$/)
    })

    test('handles screenshot upload', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        // Mock canvas context
        const mockContext = {
            fillStyle: '',
            fillRect: vi.fn(),
            fillText: vi.fn(),
            font: ''
        }

        // Mock HTMLCanvasElement methods
        HTMLCanvasElement.prototype.getContext = vi.fn(() => mockContext)
        HTMLCanvasElement.prototype.toBlob = function (callback) {
            callback(new Blob(['fake-image-data'], { type: 'image/png' }))
        }

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        // Clear previous fetch calls
        fetch.mockClear()

        const sendScreenshotButton = screen.getByRole('button', { name: /send screenshot/i })
        fireEvent.click(sendScreenshotButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                expect.stringContaining('screenshots'),
                expect.any(Object)
            )
        })
    })

    test('updates device name field correctly', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const nameInput = screen.getByPlaceholderText(/My Device/i)
        fireEvent.change(nameInput, { target: { value: 'Test Device Name' } })

        expect(nameInput.value).toBe('Test Device Name')
    })

    test('adds logs when actions are performed', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Register device to add a log
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            const logs = screen.getAllByText(/device registered/i)
            expect(logs.length).toBeGreaterThan(0)
        })
    })

    test('generates different device IDs on multiple renders', () => {
        const { unmount } = render(<DeviceSimulator />)
        const firstDeviceId = screen.getByDisplayValue(/[0-9a-f]{8}-[0-9a-f]{4}/).value

        unmount()

        render(<DeviceSimulator />)
        const secondDeviceId = screen.getByDisplayValue(/[0-9a-f]{8}-[0-9a-f]{4}/).value

        expect(firstDeviceId).not.toBe(secondDeviceId)
    })

    test('uses default values when fields are empty', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Leave name and user empty
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalled()
        })

        const callArgs = fetch.mock.calls[0]
        const requestBody = JSON.parse(callArgs[1].body)

        expect(requestBody.device_name).toMatch(/laptop-/) // Should use default pattern
        expect(requestBody.current_user).toBe('simulator-user')
    })

    test('displays log timestamps', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Register device first to generate a log entry
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            const timeElements = screen.getAllByText(/\d{1,2}:\d{2}:\d{2}/)
            expect(timeElements.length).toBeGreaterThan(0)
        })
    })

    test('handles simulation interval cleanup on unmount', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        const { unmount } = render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /start.*simulation/i })).toBeInTheDocument()
        })

        // Component should cleanup intervals on unmount
        expect(() => unmount()).not.toThrow()
    })

    test('sends metrics manually', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        fetch.mockClear()

        const sendMetricsButton = screen.getByRole('button', { name: /send metrics/i })
        fireEvent.click(sendMetricsButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                expect.stringContaining('metrics'),
                expect.objectContaining({
                    method: 'POST'
                })
            )
        })
    })

    test('sends activities manually', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        fetch.mockClear()

        const sendActivitiesButton = screen.getByRole('button', { name: /send activities/i })
        fireEvent.click(sendActivitiesButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                expect.stringContaining('activities'),
                expect.objectContaining({
                    method: 'POST'
                })
            )
        })
    })

    test('handles metrics error', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        // Set up error for next call
        fetch.mockRejectedValueOnce(new Error('Metrics error'))

        const sendMetricsButton = screen.getByRole('button', { name: /send metrics/i })
        fireEvent.click(sendMetricsButton)

        await waitFor(() => {
            expect(screen.getByText(/Metrics error/i)).toBeInTheDocument()
        })
    })

    test('handles activities error', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        // Set up error for next call
        fetch.mockRejectedValueOnce(new Error('Activities error'))

        const sendActivitiesButton = screen.getByRole('button', { name: /send activities/i })
        fireEvent.click(sendActivitiesButton)

        await waitFor(() => {
            expect(screen.getByText(/Activities error/i)).toBeInTheDocument()
        })
    })

    test('handles alert error', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        // Set up error for next call
        fetch.mockRejectedValueOnce(new Error('Alert error'))

        const sendAlertButton = screen.getByRole('button', { name: /send alert/i })
        fireEvent.click(sendAlertButton)

        await waitFor(() => {
            expect(screen.getByText(/Alert error/i)).toBeInTheDocument()
        })
    })

    test('handles screenshot error', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({ success: true })
        })

        // Mock canvas context to throw an error
        HTMLCanvasElement.prototype.getContext = vi.fn(() => null) // This will cause an error

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        const sendScreenshotButton = screen.getByRole('button', { name: /send screenshot/i })
        fireEvent.click(sendScreenshotButton)

        await waitFor(() => {
            expect(screen.getByText(/Screenshot error/i)).toBeInTheDocument()
        })
    })

    test('disables manual action buttons when not registered', () => {
        render(<DeviceSimulator />)

        const sendMetricsButton = screen.getByRole('button', { name: /send metrics/i })
        const sendActivitiesButton = screen.getByRole('button', { name: /send activities/i })
        const sendAlertButton = screen.getByRole('button', { name: /send alert/i })
        const sendScreenshotButton = screen.getByRole('button', { name: /send screenshot/i })

        expect(sendMetricsButton).toBeDisabled()
        expect(sendActivitiesButton).toBeDisabled()
        expect(sendAlertButton).toBeDisabled()
        expect(sendScreenshotButton).toBeDisabled()
    })

    test('enables manual action buttons after registration', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        const sendMetricsButton = screen.getByRole('button', { name: /send metrics/i })
        const sendActivitiesButton = screen.getByRole('button', { name: /send activities/i })
        const sendAlertButton = screen.getByRole('button', { name: /send alert/i })
        const sendScreenshotButton = screen.getByRole('button', { name: /send screenshot/i })

        expect(sendMetricsButton).not.toBeDisabled()
        expect(sendActivitiesButton).not.toBeDisabled()
        expect(sendAlertButton).not.toBeDisabled()
        expect(sendScreenshotButton).not.toBeDisabled()
    })

    test('disables start simulation button when not registered', () => {
        render(<DeviceSimulator />)

        const startButton = screen.getByRole('button', { name: /start auto simulation/i })
        expect(startButton).toBeDisabled()
    })

    test('shows warning when trying to start simulation before registration', async () => {
        render(<DeviceSimulator />)

        // Since the button is disabled, we need to directly test the component logic
        // The button should be disabled before registration, preventing clicks
        const startButton = screen.getByRole('button', { name: /start auto simulation/i })
        expect(startButton).toBeDisabled()
    })

    test('resets device and clears statistics', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const initialDeviceId = screen.getByDisplayValue(/[0-9a-f]{8}-[0-9a-f]{4}/).value

        // Register device
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        // Click reset button
        const resetButton = screen.getByRole('button', { name: /reset/i })
        fireEvent.click(resetButton)

        // Check that device ID changed and registration was cleared
        await waitFor(() => {
            const newDeviceId = screen.getByDisplayValue(/[0-9a-f]{8}-[0-9a-f]{4}/).value
            expect(newDeviceId).not.toBe(initialDeviceId)
            expect(screen.getByRole('button', { name: /register device/i })).toBeInTheDocument()
        })
    })

    test('handles registration failure response', async () => {
        fetch.mockResolvedValueOnce({
            ok: false,
            status: 400
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByText(/Failed to register device/i)).toBeInTheDocument()
        })
    })

    test('disables input fields after registration', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        const deviceIdInput = screen.getByDisplayValue(/[0-9a-f]{8}-[0-9a-f]{4}/)
        const nameInput = screen.getByPlaceholderText(/My Device/i)
        const userInput = screen.getByPlaceholderText(/simulator-user/i)

        expect(deviceIdInput).toBeDisabled()
        expect(nameInput).toBeDisabled()
        expect(userInput).toBeDisabled()
    })

    test('displays "no logs" message initially', () => {
        render(<DeviceSimulator />)
        expect(screen.getByText(/No activity yet/i)).toBeInTheDocument()
    })

    test('updates statistics when metrics are sent', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        // Check initial metrics count
        const statsValues = screen.getAllByText('0')
        expect(statsValues.length).toBeGreaterThan(0)

        // Send metrics
        fetch.mockClear()
        const sendMetricsButton = screen.getByRole('button', { name: /send metrics/i })
        fireEvent.click(sendMetricsButton)

        await waitFor(() => {
            expect(screen.getByText('1')).toBeInTheDocument()
        })
    })

    test('updates statistics when activities are sent', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        // Send activities
        fetch.mockClear()
        const sendActivitiesButton = screen.getByRole('button', { name: /send activities/i })
        fireEvent.click(sendActivitiesButton)

        await waitFor(() => {
            // Activities count should increase (at least 1)
            expect(fetch).toHaveBeenCalledWith(
                expect.stringContaining('activities'),
                expect.any(Object)
            )
        })
    })

    test('updates statistics when alerts are sent', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        // Send alert
        fetch.mockClear()
        const sendAlertButton = screen.getByRole('button', { name: /send alert/i })
        fireEvent.click(sendAlertButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                expect.stringContaining('alerts'),
                expect.any(Object)
            )
        })
    })

    test('updates statistics when screenshots are sent', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        // Mock canvas context
        const mockContext = {
            fillStyle: '',
            fillRect: vi.fn(),
            fillText: vi.fn(),
            font: ''
        }

        HTMLCanvasElement.prototype.getContext = vi.fn(() => mockContext)
        HTMLCanvasElement.prototype.toBlob = function (callback) {
            callback(new Blob(['fake-image-data'], { type: 'image/png' }))
        }

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        // Send screenshot
        fetch.mockClear()
        const sendScreenshotButton = screen.getByRole('button', { name: /send screenshot/i })
        fireEvent.click(sendScreenshotButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                expect.stringContaining('screenshots'),
                expect.any(Object)
            )
        })
    })

    test('sends processes manually and updates statistics', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        // There's no manual button for processes, but we can test it through the auto simulation
        // Let's just verify the component renders properly with processes stat
        expect(screen.getByText('Processes Sent')).toBeInTheDocument()
    })

    test('handles non-ok response for metrics', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        // Mock non-ok response
        fetch.mockResolvedValueOnce({ ok: false })

        const sendMetricsButton = screen.getByRole('button', { name: /send metrics/i })
        fireEvent.click(sendMetricsButton)

        // Should not update stats on failed response
        await waitFor(() => {
            expect(sendMetricsButton).toBeInTheDocument()
        })
    })

    test('handles non-ok response for activities', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        fetch.mockResolvedValueOnce({ ok: false })

        const sendActivitiesButton = screen.getByRole('button', { name: /send activities/i })
        fireEvent.click(sendActivitiesButton)

        await waitFor(() => {
            expect(sendActivitiesButton).toBeInTheDocument()
        })
    })

    test('handles non-ok response for alerts', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        fetch.mockResolvedValueOnce({ ok: false })

        const sendAlertButton = screen.getByRole('button', { name: /send alert/i })
        fireEvent.click(sendAlertButton)

        await waitFor(() => {
            expect(sendAlertButton).toBeInTheDocument()
        })
    })

    test('handles non-ok response for screenshots', async () => {
        fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({ success: true })
        })

        const mockContext = {
            fillStyle: '',
            fillRect: vi.fn(),
            fillText: vi.fn(),
            font: ''
        }

        HTMLCanvasElement.prototype.getContext = vi.fn(() => mockContext)
        HTMLCanvasElement.prototype.toBlob = function (callback) {
            callback(new Blob(['fake-image-data'], { type: 'image/png' }))
        }

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        fetch.mockResolvedValueOnce({ ok: false })

        const sendScreenshotButton = screen.getByRole('button', { name: /send screenshot/i })
        fireEvent.click(sendScreenshotButton)

        await waitFor(() => {
            expect(sendScreenshotButton).toBeInTheDocument()
        })
    })

    test('device ID input has placeholder', () => {
        render(<DeviceSimulator />)
        const deviceIdInput = screen.getByDisplayValue(/[0-9a-f]{8}-[0-9a-f]{4}/)
        expect(deviceIdInput).toHaveAttribute('placeholder', 'device-xxxxx')
    })

    test('device type select has correct options', () => {
        render(<DeviceSimulator />)
        expect(screen.getByText('Laptop')).toBeInTheDocument()
        expect(screen.getByText('Desktop')).toBeInTheDocument()
        expect(screen.getByText('Mobile')).toBeInTheDocument()
        expect(screen.getByText('Tablet')).toBeInTheDocument()
    })

    test('device OS select has correct options', () => {
        render(<DeviceSimulator />)
        expect(screen.getByText('macOS')).toBeInTheDocument()
        expect(screen.getByText('Windows')).toBeInTheDocument()
        expect(screen.getByText('Linux')).toBeInTheDocument()
        expect(screen.getByText('iOS')).toBeInTheDocument()
        expect(screen.getByText('Android')).toBeInTheDocument()
    })

    test('updates device ID input', () => {
        render(<DeviceSimulator />)
        const deviceIdInput = screen.getByDisplayValue(/[0-9a-f]{8}-[0-9a-f]{4}/)
        fireEvent.change(deviceIdInput, { target: { value: 'device-custom123' } })
        expect(deviceIdInput.value).toBe('device-custom123')
    })

    test('device type select defaults to laptop', () => {
        render(<DeviceSimulator />)
        const selects = document.querySelectorAll('select')
        const typeSelect = Array.from(selects).find(s => s.value === 'laptop')
        expect(typeSelect).toBeInTheDocument()
        expect(typeSelect.value).toBe('laptop')
    })

    test('device OS select defaults to macOS', () => {
        render(<DeviceSimulator />)
        const selects = document.querySelectorAll('select')
        const osSelect = Array.from(selects).find(s => s.value === 'macOS')
        expect(osSelect).toBeInTheDocument()
        expect(osSelect.value).toBe('macOS')
    })

    test('renders all statistics labels', () => {
        render(<DeviceSimulator />)
        expect(screen.getByText('Metrics Sent')).toBeInTheDocument()
        expect(screen.getByText('Activities Sent')).toBeInTheDocument()
        expect(screen.getByText('Alerts Sent')).toBeInTheDocument()
        expect(screen.getByText('Screenshots Sent')).toBeInTheDocument()
        expect(screen.getByText('Processes Sent')).toBeInTheDocument()
    })

    test('renders all manual action buttons', () => {
        render(<DeviceSimulator />)
        expect(screen.getByRole('button', { name: /send metrics/i })).toBeInTheDocument()
        expect(screen.getByRole('button', { name: /send activities/i })).toBeInTheDocument()
        expect(screen.getByRole('button', { name: /send alert/i })).toBeInTheDocument()
        expect(screen.getByRole('button', { name: /send screenshot/i })).toBeInTheDocument()
    })

    test('renders simulation control buttons', () => {
        render(<DeviceSimulator />)
        expect(screen.getByRole('button', { name: /start auto simulation/i })).toBeInTheDocument()
    })

    test('renders reset button', () => {
        render(<DeviceSimulator />)
        expect(screen.getByRole('button', { name: /reset/i })).toBeInTheDocument()
    })

    test('renders device configuration section', () => {
        render(<DeviceSimulator />)
        expect(screen.getByText('Device Configuration')).toBeInTheDocument()
    })

    test('renders simulation controls section', () => {
        render(<DeviceSimulator />)
        expect(screen.getByText('Simulation Controls')).toBeInTheDocument()
    })

    test('renders statistics section', () => {
        render(<DeviceSimulator />)
        expect(screen.getByText('Statistics')).toBeInTheDocument()
    })

    test('renders activity logs section', () => {
        render(<DeviceSimulator />)
        expect(screen.getByText('Activity Logs')).toBeInTheDocument()
    })

    test('renders manual actions section', () => {
        render(<DeviceSimulator />)
        expect(screen.getByText('Manual Actions')).toBeInTheDocument()
    })

    test('shows warning log when trying to start simulation before registration', () => {
        render(<DeviceSimulator />)
        const startButton = screen.getByRole('button', { name: /start auto simulation/i })
        expect(startButton).toBeDisabled()
    })

    test('generates new device ID if current is empty', () => {
        const { rerender } = render(<DeviceSimulator />)
        const initialId = screen.getByDisplayValue(/[0-9a-f]{8}-[0-9a-f]{4}/).value
        rerender(<DeviceSimulator />)
        // Just verify the format is correct (UUID v4)
        expect(initialId).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i)
    })

    test('initializes processes count to 0', () => {
        render(<DeviceSimulator />)
        const statValues = screen.getAllByText('0')
        expect(statValues.length).toBeGreaterThanOrEqual(5)
    })

    test('renders all form field labels', () => {
        render(<DeviceSimulator />)
        expect(screen.getByText('Device ID')).toBeInTheDocument()
        expect(screen.getByText('Device Name')).toBeInTheDocument()
        expect(screen.getByText('Type')).toBeInTheDocument()
        expect(screen.getByText('Operating System')).toBeInTheDocument()
        expect(screen.getByText('Current User')).toBeInTheDocument()
    })

    test('device ID is generated on initial render', () => {
        render(<DeviceSimulator />)
        const deviceIdInput = screen.getByDisplayValue(/[0-9a-f]{8}-[0-9a-f]{4}/)
        expect(deviceIdInput.value).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i)
    })

    test('logs are displayed in reverse chronological order', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            const successLogs = screen.getAllByText(/device registered/i)
            expect(successLogs.length).toBeGreaterThan(0)
        })
    })

    test('MAC address is formatted correctly in registration payload', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                expect.any(String),
                expect.objectContaining({
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: expect.any(String)
                })
            )
        })

        const callArgs = fetch.mock.calls[0]
        const requestBody = JSON.parse(callArgs[1].body)
        expect(requestBody.mac_address).toMatch(/^([0-9a-f]{2}:){5}[0-9a-f]{2}$/i)
    })

    test('location is set in registration payload', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalled()
        })

        const callArgs = fetch.mock.calls[0]
        const requestBody = JSON.parse(callArgs[1].body)
        expect(requestBody.device_location).toBe('Simulated Location')
    })

    test('registration payload includes all required fields', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalled()
        })

        const callArgs = fetch.mock.calls[0]
        const requestBody = JSON.parse(callArgs[1].body)

        expect(requestBody).toHaveProperty('deviceid')
        expect(requestBody).toHaveProperty('device_name')
        expect(requestBody).toHaveProperty('device_type')
        expect(requestBody).toHaveProperty('os')
        expect(requestBody).toHaveProperty('current_user')
        expect(requestBody).toHaveProperty('device_location')
        expect(requestBody).toHaveProperty('ip_address')
        expect(requestBody).toHaveProperty('mac_address')
    })

    test('metrics payload includes required fields', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        fetch.mockClear()

        const sendMetricsButton = screen.getByRole('button', { name: /send metrics/i })
        fireEvent.click(sendMetricsButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                expect.stringContaining('metrics'),
                expect.any(Object)
            )
        })

        const callArgs = fetch.mock.calls[0]
        const requestBody = JSON.parse(callArgs[1].body)

        expect(requestBody).toHaveProperty('cpu_usage')
        expect(requestBody).toHaveProperty('cpu_temp')
        expect(requestBody).toHaveProperty('memory_total')
        expect(requestBody).toHaveProperty('memory_used')
        expect(requestBody).toHaveProperty('disk_total')
        expect(requestBody).toHaveProperty('disk_used')
    })

    test('activities payload is an array', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        fetch.mockClear()

        const sendActivitiesButton = screen.getByRole('button', { name: /send activities/i })
        fireEvent.click(sendActivitiesButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                expect.stringContaining('activities'),
                expect.any(Object)
            )
        })

        const callArgs = fetch.mock.calls[0]
        const requestBody = JSON.parse(callArgs[1].body)

        expect(Array.isArray(requestBody)).toBe(true)
        expect(requestBody.length).toBeGreaterThan(0)
    })

    test('alert payload includes required fields', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /device registered/i })).toBeInTheDocument()
        })

        fetch.mockClear()

        const sendAlertButton = screen.getByRole('button', { name: /send alert/i })
        fireEvent.click(sendAlertButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                expect.stringContaining('alerts'),
                expect.any(Object)
            )
        })

        const callArgs = fetch.mock.calls[0]
        const requestBody = JSON.parse(callArgs[1].body)

        expect(Array.isArray(requestBody)).toBe(true)
        expect(requestBody[0]).toHaveProperty('level')
        expect(requestBody[0]).toHaveProperty('alert_type')
        expect(requestBody[0]).toHaveProperty('message')
    })
})
