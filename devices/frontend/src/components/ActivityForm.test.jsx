import { test, expect, vi, describe, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import ActivityForm from './ActivityForm'

// Mock axios
vi.mock('axios', () => ({
    default: {
        post: vi.fn()
    }
}))

describe('ActivityForm Component', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    afterEach(() => {
        vi.restoreAllMocks()
    })

    test('renders form with all required fields', () => {
        render(<ActivityForm />)

        expect(screen.getByText('User Activity Submission')).toBeInTheDocument()
        expect(screen.getByPlaceholderText('Location')).toBeInTheDocument()
        expect(screen.getByPlaceholderText('Password')).toBeInTheDocument()
        expect(screen.getByRole('button', { name: 'Add Activity' })).toBeInTheDocument()

        const fileInput = screen.getByLabelText(/file/i) || screen.getByDisplayValue('')
        expect(fileInput).toBeInTheDocument()
        expect(fileInput).toHaveAttribute('type', 'file')
        expect(fileInput).toHaveAttribute('accept', 'image/*')
    })

    test('updates location input value', () => {
        render(<ActivityForm />)

        const locationInput = screen.getByPlaceholderText('Location')
        fireEvent.change(locationInput, { target: { value: 'Test Office' } })

        expect(locationInput.value).toBe('Test Office')
    })

    test('updates password input value', () => {
        render(<ActivityForm />)

        const passwordInput = screen.getByPlaceholderText('Password')
        fireEvent.change(passwordInput, { target: { value: 'testpassword' } })

        expect(passwordInput.value).toBe('testpassword')
    })

    test('handles file selection', () => {
        render(<ActivityForm />)

        const fileInput = screen.getByLabelText(/file/i) || document.querySelector('input[type="file"]')
        const testFile = new File(['test image'], 'test.png', { type: 'image/png' })

        fireEvent.change(fileInput, { target: { files: [testFile] } })

        expect(fileInput.files[0]).toBe(testFile)
        expect(fileInput.files[0].name).toBe('test.png')
    })

    test('shows validation for required fields', () => {
        render(<ActivityForm />)

        const locationInput = screen.getByPlaceholderText('Location')
        const passwordInput = screen.getByPlaceholderText('Password')
        const fileInput = document.querySelector('input[type="file"]')

        expect(locationInput).toHaveAttribute('required')
        expect(passwordInput).toHaveAttribute('required')
        expect(fileInput).toHaveAttribute('required')
    })

    test('successful form submission', async () => {
        const axios = await import('axios')
        axios.default.post.mockResolvedValue({
            data: { status: 'success' }
        })

        render(<ActivityForm />)

        const locationInput = screen.getByPlaceholderText('Location')
        const passwordInput = screen.getByPlaceholderText('Password')
        const fileInput = document.querySelector('input[type="file"]')
        const submitButton = screen.getByRole('button', { name: 'Add Activity' })

        const testFile = new File(['test image'], 'test.png', { type: 'image/png' })

        fireEvent.change(locationInput, { target: { value: 'Test Location' } })
        fireEvent.change(passwordInput, { target: { value: 'testpass' } })
        fireEvent.change(fileInput, { target: { files: [testFile] } })

        fireEvent.click(submitButton)

        await waitFor(() => {
            expect(screen.getByText('Activity added successfully')).toBeInTheDocument()
        })

        expect(axios.default.post).toHaveBeenCalledWith(
            '/upload/',
            expect.any(FormData),
            {
                headers: {
                    'Content-Type': 'multipart/form-data',
                },
            }
        )
    })

    test('handles API error response', async () => {
        const axios = await import('axios')
        axios.default.post.mockResolvedValue({
            data: { status: 'error' }
        })

        render(<ActivityForm />)

        const locationInput = screen.getByPlaceholderText('Location')
        const passwordInput = screen.getByPlaceholderText('Password')
        const fileInput = document.querySelector('input[type="file"]')
        const submitButton = screen.getByRole('button', { name: 'Add Activity' })

        const testFile = new File(['test image'], 'test.png', { type: 'image/png' })

        fireEvent.change(locationInput, { target: { value: 'Test Location' } })
        fireEvent.change(passwordInput, { target: { value: 'testpass' } })
        fireEvent.change(fileInput, { target: { files: [testFile] } })

        fireEvent.click(submitButton)

        await waitFor(() => {
            expect(screen.getByText('Failed to add activity')).toBeInTheDocument()
        })
    })

    test('handles network error', async () => {
        const axios = await import('axios')
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => { })
        axios.default.post.mockRejectedValue(new Error('Network error'))

        render(<ActivityForm />)

        const locationInput = screen.getByPlaceholderText('Location')
        const passwordInput = screen.getByPlaceholderText('Password')
        const fileInput = document.querySelector('input[type="file"]')
        const submitButton = screen.getByRole('button', { name: 'Add Activity' })

        const testFile = new File(['test image'], 'test.png', { type: 'image/png' })

        fireEvent.change(locationInput, { target: { value: 'Test Location' } })
        fireEvent.change(passwordInput, { target: { value: 'testpass' } })
        fireEvent.change(fileInput, { target: { files: [testFile] } })

        fireEvent.click(submitButton)

        await waitFor(() => {
            expect(screen.getByText('Failed to add activity')).toBeInTheDocument()
        })

        expect(consoleError).toHaveBeenCalledWith('Error:', expect.any(Error))
        consoleError.mockRestore()
    })

    test('form data is correctly constructed', async () => {
        const axios = await import('axios')
        axios.default.post.mockResolvedValue({
            data: { status: 'success' }
        })

        render(<ActivityForm />)

        const locationInput = screen.getByPlaceholderText('Location')
        const passwordInput = screen.getByPlaceholderText('Password')
        const fileInput = document.querySelector('input[type="file"]')
        const submitButton = screen.getByRole('button', { name: 'Add Activity' })

        const testFile = new File(['test image'], 'test.png', { type: 'image/png' })

        fireEvent.change(locationInput, { target: { value: 'Office Building' } })
        fireEvent.change(passwordInput, { target: { value: 'secret123' } })
        fireEvent.change(fileInput, { target: { files: [testFile] } })

        fireEvent.click(submitButton)

        await waitFor(() => {
            expect(axios.default.post).toHaveBeenCalled()
        })

        const formData = axios.default.post.mock.calls[0][1]
        expect(formData).toBeInstanceOf(FormData)
    })

    test('clears message on new submission', async () => {
        const axios = await import('axios')
        axios.default.post.mockResolvedValue({
            data: { status: 'success' }
        })

        render(<ActivityForm />)

        const locationInput = screen.getByPlaceholderText('Location')
        const passwordInput = screen.getByPlaceholderText('Password')
        const fileInput = document.querySelector('input[type="file"]')
        const submitButton = screen.getByRole('button', { name: 'Add Activity' })

        const testFile = new File(['test image'], 'test.png', { type: 'image/png' })

        // First submission
        fireEvent.change(locationInput, { target: { value: 'Location 1' } })
        fireEvent.change(passwordInput, { target: { value: 'pass1' } })
        fireEvent.change(fileInput, { target: { files: [testFile] } })
        fireEvent.click(submitButton)

        await waitFor(() => {
            expect(screen.getByText('Activity added successfully')).toBeInTheDocument()
        })

        // Second submission should clear and update message
        fireEvent.change(locationInput, { target: { value: 'Location 2' } })
        fireEvent.change(passwordInput, { target: { value: 'pass2' } })
        fireEvent.click(submitButton)

        await waitFor(() => {
            expect(screen.getByText('Activity added successfully')).toBeInTheDocument()
        })
    })

    test('accepts only image files', () => {
        render(<ActivityForm />)

        const fileInput = document.querySelector('input[type="file"]')
        expect(fileInput).toHaveAttribute('accept', 'image/*')
    })

    test('prevents default form submission', async () => {
        const axios = await import('axios')
        axios.default.post.mockResolvedValue({
            data: { status: 'success' }
        })

        render(<ActivityForm />)

        const form = document.querySelector('form')
        const mockPreventDefault = vi.fn()

        const submitEvent = new Event('submit', { bubbles: true, cancelable: true })
        submitEvent.preventDefault = mockPreventDefault

        fireEvent(form, submitEvent)

        expect(mockPreventDefault).toHaveBeenCalled()
    })
})