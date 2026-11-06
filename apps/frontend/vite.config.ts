import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  test: {
    globals: true,
  },
  // The production image serves with nginx, not `vite preview`. This is for the "runtime" stage
  // (docker-entrypoint.sh), which does use it behind a proxy that routes by host.
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
