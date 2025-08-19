import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Proxy API requests to Go server to avoid CORS during dev
export default defineConfig({
  base: '/ui/',
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/grpc': {
        target: 'http://localhost:8081',
        changeOrigin: true
      }
    }
  }
})


