/* eslint-env node */
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

export default ({ mode }) => {
  // Load env variables based on the current mode (development, production, etc.)
  const env = loadEnv(mode, process.cwd(), '')

  // Use stable port with fallback - no dynamic conflicts
  const port = Number(env.VITE_DEVICES_FRONTEND_PORT || process.env.VITE_DEVICES_FRONTEND_PORT || 4000)
  return defineConfig({
    plugins: [react()],
    server: {
      port,
      strictPort: true, // Fail if port is in use instead of auto-increment
      host: true, // Listen on all interfaces
      // Use Kubernetes NodePort service - no port-forwarding needed
    },
    test: {
      environment: 'jsdom',
      setupFiles: './src/setupTests.js',
      globals: true,
      coverage: {
        provider: 'v8',
        reporter: ['text', 'json', 'html', 'lcov'],
        reportsDirectory: './coverage',
        exclude: [
          'node_modules/',
          'src/setupTests.js',
          '**/*.{test,spec}.{js,jsx,ts,tsx}',
          '**/coverage/**'
        ]
      }
    }
  })
}
