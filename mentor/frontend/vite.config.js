import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

export default ({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const port = Number(env.VITE_MENTOR_FRONTEND_PORT || process.env.VITE_MENTOR_FRONTEND_PORT || 5000)
  return defineConfig({
    plugins: [react()],
    server: {
      port,
      strictPort: false,
    },
    test: {
      environment: 'jsdom',
      setupFiles: './src/setupTests.js'
    }
  })
}
