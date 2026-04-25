import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/healthz': 'http://localhost:17654',
      '/videos': 'http://localhost:17654',
      '/tags': 'http://localhost:17654',
      '/sync': 'http://localhost:17654',
      '/directories': 'http://localhost:17654',
      '/jav': 'http://localhost:17654',
      '/config': 'http://localhost:17654',
    },
  },
})
