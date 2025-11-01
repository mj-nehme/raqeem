import { test, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import React from 'react'
import DeviceDashboard from './DeviceDashboard'

test('renders Devices list heading', () => {
    render(<DeviceDashboard />)
    expect(screen.getByText('Devices')).toBeInTheDocument()
})
