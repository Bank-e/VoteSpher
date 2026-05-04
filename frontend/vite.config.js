import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/voter': 'http://localhost:8080',
      '/ballot': 'http://localhost:8080',
      '/candidates': 'http://localhost:8080',
      '/parties': 'http://localhost:8080',
      '/results': 'http://localhost:8080',
      '/election': 'http://localhost:8080',
    }
  }
})
