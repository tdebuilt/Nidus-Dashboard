import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte(), tailwindcss()],
  build: {
    outDir: 'static',
    emptyOutDir: true,
    chunkSizeWarningLimit: 1500,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules/svelte')) {
            return 'vendor-svelte'
          }
        },
      },
    },
  },
  server: {
    proxy: {
      '/api': 'http://localhost:3777',
    },
  },
  test: {
    environment: 'jsdom',
    include: ['src/**/*.test.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      include: ['src/lib/**/*.ts', 'src/lib/**/*.svelte', 'src/pages/**/*.svelte'],
      exclude: ['src/**/*.test.ts', 'src/**/*.d.ts'],
    },
  },
  resolve: {
    conditions: ['browser'],
  },
})
