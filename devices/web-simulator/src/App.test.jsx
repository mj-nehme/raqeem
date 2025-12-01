import { test, expect, describe } from 'vitest'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import App from './App'

describe('App Component', () => {
    test('renders without crashing', () => {
        render(<App />)
        expect(screen.getByText(/Device Simulator/i)).toBeInTheDocument()
    })

    test('contains DeviceSimulator component', () => {
        render(<App />)
        // Check for key elements from DeviceSimulator
        expect(screen.getByText(/Device Simulator/i)).toBeInTheDocument()
        expect(screen.getByRole('button', { name: /register device/i })).toBeInTheDocument()
    })

    test('applies app styling', () => {
        const { container } = render(<App />)
        const appDiv = container.querySelector('.app')
        expect(appDiv).toBeInTheDocument()
    })
})
