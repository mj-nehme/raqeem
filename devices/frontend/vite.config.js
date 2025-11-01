/* eslint-env node */
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

export default ({ mode }) => {
  // Load env variables based on the current mode (development, production, etc.)
  const env = loadEnv(mode, process.cwd(), '')

  const port = Number(env.VITE_DEVICES_FRONTEND_PORT || process.env.VITE_DEVICES_FRONTEND_PORT || 4000)
  return defineConfig({
    plugins: [react()],
    server: {
      port,
      // Proxy removed - frontend uses env-configured backend URL
    },
    test: {
      environment: 'jsdom',
      setupFiles: './src/setupTests.js'
    }
  })
}
