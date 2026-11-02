import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  test: {
    globals: true,
  },
  // The production image serves with `vite preview` behind a reverse proxy that routes by
  // host, so any host name reaches it.
  preview: {
    allowedHosts: true,
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/feeds': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
