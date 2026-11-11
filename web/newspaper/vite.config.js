import fs from 'fs'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

const pkg = JSON.parse(fs.readFileSync(path.resolve(__dirname, 'package.json'), 'utf-8'))
const appVersion = process.env.APP_VERSION || pkg.version || '0.2.0'

export default defineConfig({
  base: '/',
  plugins: [vue()],
  define: {
    __APP_VERSION__: JSON.stringify(appVersion)
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    emptyOutDir: true
  },
  server: {
    port: 5174,
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true
      }
    }
  }
})
