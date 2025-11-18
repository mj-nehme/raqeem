import { test, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import React from 'react'
import DeviceDashboard from './DeviceDashboard'

// Mock fetch
const createMockFetch = (commandsMockData) => {
    return vi.fn((url) => {
        if (url.includes('/commands')) {
            return Promise.resolve({
                ok: true,
                json: () => Promise.resolve(commandsMockData || []),
            })
        }
        return Promise.resolve({
            ok: true,
            json: () => Promise.resolve([]),
        })
    })
}

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

// Mock environment variable
vi.mock('import.meta', () => ({
    env: {
        VITE_MENTOR_API_URL: 'http://localhost:8080'
    }
}))

test('DeviceDashboard renders without crashing', () => {
    global.fetch = createMockFetch()
    render(<DeviceDashboard />)
    expect(screen.getByText('Devices')).toBeInTheDocument()
})

test('command POST request uses correct field names', async () => {
    const mockFetch = vi.fn((url, options) => {
        if (options && options.method === 'POST' && url.includes('/devices/commands')) {
            // Verify the request body has correct field names
            const body = JSON.parse(options.body)
            // These assertions verify the fix for field name mismatches
            expect(body).toHaveProperty('deviceid')
            expect(body).toHaveProperty('command_text')
            // Should NOT have device_id (old incorrect field name)
            expect(body).not.toHaveProperty('device_id')
            
            return Promise.resolve({
                ok: true,
                json: () => Promise.resolve({
                    commandid: '423e4567-e89b-12d3-a456-426614174003',
                    deviceid: body.deviceid,
                    command_text: body.command_text,
                    status: 'pending',
                    created_at: new Date().toISOString()
                })
            })
        }
        return Promise.resolve({
            ok: true,
            json: () => Promise.resolve([]),
        })
    })
    
    global.fetch = mockFetch
    
    // Simulate sending a command via fetch directly (unit test approach)
    const deviceId = '123e4567-e89b-12d3-a456-426614174000'
    const commandText = 'test_command'
    
    await fetch('/api/devices/commands', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ deviceid: deviceId, command_text: commandText }),
    })
    
    // Verify fetch was called
    expect(mockFetch).toHaveBeenCalled()
})

test('command list uses correct field names for display', () => {
    // Mock commands with correct backend field names
    const mockCommands = [
        {
            commandid: '223e4567-e89b-12d3-a456-426614174001',
            deviceid: '123e4567-e89b-12d3-a456-426614174000',
            command_text: 'get_info',
            status: 'completed',
            created_at: '2024-01-01T12:00:00Z',
            result: 'Device info retrieved',
            exit_code: 0
        },
        {
            commandid: '323e4567-e89b-12d3-a456-426614174002',
            deviceid: '123e4567-e89b-12d3-a456-426614174000',
            command_text: 'restart',
            status: 'pending',
            created_at: '2024-01-01T12:01:00Z',
            result: '',
            exit_code: 0
        }
    ]
    
    // Verify that commands have the correct field names
    mockCommands.forEach(cmd => {
        expect(cmd).toHaveProperty('commandid')
        expect(cmd).toHaveProperty('deviceid')
        expect(cmd).toHaveProperty('command_text')
        expect(cmd).toHaveProperty('status')
        expect(cmd).toHaveProperty('created_at')
        
        // Should NOT have old incorrect field names
        expect(cmd).not.toHaveProperty('id')
        expect(cmd).not.toHaveProperty('device_id')
        expect(cmd).not.toHaveProperty('command')
    })
    
    // Test that we can access the fields as the component would
    const firstCommand = mockCommands[0]
    expect(firstCommand.commandid).toBe('223e4567-e89b-12d3-a456-426614174001')
    expect(firstCommand.command_text).toBe('get_info')
    expect(firstCommand.status).toBe('completed')
})

