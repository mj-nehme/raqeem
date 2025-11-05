import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

export default ({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  // Use stable port with fallback - no dynamic conflicts  
  const port = Number(env.VITE_MENTOR_FRONTEND_PORT || process.env.VITE_MENTOR_FRONTEND_PORT || 5000)
  return defineConfig({
    plugins: [react()],
    server: {
      port,
      strictPort: true, // Fail if port is in use instead of auto-increment
      host: true, // Listen on all interfaces
      proxy: {
        // Proxy API calls to stable Kubernetes NodePort service
        '/api': {
          target: env.VITE_MENTOR_API_URL || process.env.VITE_MENTOR_API_URL || 'http://localhost:30081',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, ''),
        },
      },
    },
    test: {
      environment: 'jsdom',
      setupFiles: './src/setupTests.js',
      testTimeout: 10000,
      env: {
        VITE_MENTOR_API_URL: 'http://localhost:8080'
      },
      coverage: {
        provider: 'v8',
        reporter: ['text', 'json', 'html', 'lcov'],
        reportsDirectory: './coverage',
        reportOnFailure: true, // Generate coverage even when tests fail
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
