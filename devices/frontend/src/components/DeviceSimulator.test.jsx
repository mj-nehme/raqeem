import { test, expect, vi, describe, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import DeviceSimulator from './DeviceSimulator'

// Mock fetch
// eslint-disable-next-line no-undef
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

    test('renders Device Simulator and Send Alert button', () => {
        render(<DeviceSimulator />)
        expect(screen.getByText('Device Simulator')).toBeInTheDocument()
        expect(screen.getByRole('button', { name: /send alert/i })).toBeInTheDocument()
    })

    test('renders device registration form fields', () => {
        render(<DeviceSimulator />)

        expect(screen.getByLabelText(/device name/i)).toBeInTheDocument()
        expect(screen.getByLabelText(/device type/i)).toBeInTheDocument()
        expect(screen.getByLabelText(/operating system/i)).toBeInTheDocument()
        expect(screen.getByLabelText(/current user/i)).toBeInTheDocument()
        expect(screen.getByRole('button', { name: /register device/i })).toBeInTheDocument()
    })

    test('generates device ID automatically', () => {
        render(<DeviceSimulator />)

        const deviceIdInput = screen.getByDisplayValue(/device-/)
        expect(deviceIdInput.value).toMatch(/^device-[a-z0-9]{9}$/)
    })

    test('updates device name input', () => {
        render(<DeviceSimulator />)

        const nameInput = screen.getByLabelText(/device name/i)
        fireEvent.change(nameInput, { target: { value: 'Test Device' } })

        expect(nameInput.value).toBe('Test Device')
    })

    test('updates device type selection', () => {
        render(<DeviceSimulator />)

        const typeSelect = screen.getByLabelText(/device type/i)
        fireEvent.change(typeSelect, { target: { value: 'desktop' } })

        expect(typeSelect.value).toBe('desktop')
    })

    test('updates operating system selection', () => {
        render(<DeviceSimulator />)

        const osSelect = screen.getByLabelText(/operating system/i)
        fireEvent.change(osSelect, { target: { value: 'Linux' } })

        expect(osSelect.value).toBe('Linux')
    })

    test('updates current user input', () => {
        render(<DeviceSimulator />)

        const userInput = screen.getByLabelText(/current user/i)
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

        const nameInput = screen.getByLabelText(/device name/i)
        const typeSelect = screen.getByLabelText(/device type/i)
        const osSelect = screen.getByLabelText(/operating system/i)
        const userInput = screen.getByLabelText(/current user/i)

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

        // Check initial stats
        expect(screen.getByText(/metrics: 0/i)).toBeInTheDocument()
        expect(screen.getByText(/activities: 0/i)).toBeInTheDocument()
        expect(screen.getByText(/alerts: 0/i)).toBeInTheDocument()
        expect(screen.getByText(/screenshots: 0/i)).toBeInTheDocument()
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

        render(<DeviceSimulator />)

        // Register device first
        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByRole('button', { name: /upload screenshot/i })).toBeInTheDocument()
        })

        const uploadButton = screen.getByRole('button', { name: /upload screenshot/i })
        fireEvent.click(uploadButton)

        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                expect.stringContaining('screenshots'),
                expect.objectContaining({
                    method: 'POST'
                })
            )
        })
    })

    test('displays device information after registration', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        const nameInput = screen.getByLabelText(/device name/i)
        fireEvent.change(nameInput, { target: { value: 'Test Device Name' } })

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            expect(screen.getByText(/test device name/i)).toBeInTheDocument()
        })
    })

    test('clears logs when limit reached', async () => {
        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        render(<DeviceSimulator />)

        // Trigger multiple log entries
        for (let i = 0; i < 55; i++) {
            const sendAlertButton = screen.getByRole('button', { name: /send alert/i })
            fireEvent.click(sendAlertButton)
            await new Promise(resolve => setTimeout(resolve, 10))
        }

        // Should only keep latest 50 logs
        const logEntries = screen.getAllByText(/alert sent/i)
        expect(logEntries.length).toBeLessThanOrEqual(50)
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

        const sendAlertButton = screen.getByRole('button', { name: /send alert/i })
        fireEvent.click(sendAlertButton)

        await waitFor(() => {
            const timeElements = screen.getAllByText(/\d{1,2}:\d{2}:\d{2}/)
            expect(timeElements.length).toBeGreaterThan(0)
        })
    })

    test('handles simulation interval cleanup on unmount', async () => {
        const { unmount } = render(<DeviceSimulator />)

        fetch.mockResolvedValue({
            ok: true,
            json: async () => ({ success: true })
        })

        const registerButton = screen.getByRole('button', { name: /register device/i })
        fireEvent.click(registerButton)

        await waitFor(() => {
            const startButton = screen.getByRole('button', { name: /start simulation/i })
            fireEvent.click(startButton)
        })

        // Component should cleanup intervals on unmount
        expect(() => unmount()).not.toThrow()
    })
})
