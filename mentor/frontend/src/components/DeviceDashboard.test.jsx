import { test, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import React from 'react'
import DeviceDashboard from './DeviceDashboard'

// Mock fetch
global.fetch = vi.fn(() =>
    Promise.resolve({
        ok: true,
        json: () => Promise.resolve([]),
    })
)

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

test('renders Devices list heading', () => {
    render(<DeviceDashboard />)
    expect(screen.getByText('Devices')).toBeInTheDocument()
})
