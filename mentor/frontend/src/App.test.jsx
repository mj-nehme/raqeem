import { test, expect, describe, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import App from './App'

// Mock fetch
global.fetch = vi.fn()

// Mock recharts to avoid canvas issues
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

describe('App Component', () => {
    beforeEach(() => {
        fetch.mockClear()
        fetch.mockResolvedValue({
            ok: true,
            json: async () => []
        })
    })

    test('renders without crashing', () => {
        render(<App />)
        expect(screen.getByText('Devices')).toBeInTheDocument()
    })

    test('renders with Material-UI theme', () => {
        const { container } = render(<App />)
        // Check that MUI components are rendered
        expect(container.querySelector('.MuiBox-root')).toBeInTheDocument()
    })

    test('contains DeviceDashboard component', () => {
        render(<App />)
        // Check for key elements from DeviceDashboard
        const devicesHeadings = screen.getAllByText('Devices')
        expect(devicesHeadings.length).toBeGreaterThan(0)
    })
})
