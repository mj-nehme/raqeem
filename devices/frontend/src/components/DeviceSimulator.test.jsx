import { test, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import DeviceSimulator from './DeviceSimulator'
import React from 'react'

test('renders Device Simulator and Send Alert button', () => {
    render(<DeviceSimulator />)
    expect(screen.getByText('🖥️ Device Simulator')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /send alert/i })).toBeInTheDocument()
})
