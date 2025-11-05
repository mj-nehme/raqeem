import { test, expect, vi, describe, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import DeviceSimulator from './DeviceSimulator'

// Mock fetch
global.fetch = vi.fn()

// Mock environment variable
vi.mock('import.meta', () => ({
    env: {
        VITE_DEVICES_API_URL: 'http://localhost:3000/api'
    }
}))

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

        const deviceIdInput = screen.getByDisplayValue(/device-/)
        expect(deviceIdInput.value).toMatch(/^device-[a-z0-9]{9}$/)
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
                'http://localhost:3000/api/v1/devices/register',
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

        expect(requestBody.name).toBe('My Test Device')
        expect(requestBody.type).toBe('tablet')
        expect(requestBody.os).toBe('Android')
        expect(requestBody.current_user).toBe('testuser123')
    })

    test('handles registration error', async () => {
        fetch.mockRejectedValueOnce(new Error('Network error'))
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => { })

        render(<DeviceSimulator />)

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(consoleError).toHaveBeenCalled()
        })

        consoleError.mockRestore()
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
            expect(screen.getByRole('button', { name: /start simulation/i })).toBeInTheDocument()
        })

        // Start simulation
        const startButton = screen.getByRole('button', { name: /start simulation/i })
        fireEvent.click(startButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /stop simulation/i })).toBeInTheDocument()
        })

        // Stop simulation
        const stopButton = screen.getByRole('button', { name: /stop simulation/i })
        fireEvent.click(stopButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /start simulation/i })).toBeInTheDocument()
        })
    })

    test('sends alert manually', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

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
            expect(screen.getByText(/device registered/i)).toBeInTheDocument()
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

        // Mock HTMLCanvasElement.prototype.toBlob
        HTMLCanvasElement.prototype.toBlob = function(callback) {
            callback(new Blob(['fake-image-data'], { type: 'image/png' }))
        }

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /registered/i })).toBeInTheDocument()
        })

        const sendScreenshotButton = screen.getByRole('button', { name: /send screenshot/i })
        fireEvent.click(sendScreenshotButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                expect.stringContaining('screenshots'),
                expect.objectContaining({
                    method: 'POST'
                })
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
            expect(screen.getByText(/device registered/i)).toBeInTheDocument()
        })
    })

    test('generates different device IDs on multiple renders', () => {
        const { unmount } = render(<DeviceSimulator />)
        const firstDeviceId = screen.getByDisplayValue(/device-/).value

        unmount()

        render(<DeviceSimulator />)
        const secondDeviceId = screen.getByDisplayValue(/device-/).value

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

        expect(requestBody.name).toMatch(/laptop-/) // Should use default pattern
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
})
