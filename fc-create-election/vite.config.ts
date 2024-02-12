import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

const base = process.env.BASE_URL || '/'
const outDir = process.env.BUILD_PATH || 'dist'

// https://vitejs.dev/config/
export default defineConfig({
  base,
  build: {
    outDir,
  },
  define: {
    'import.meta.env.BACKEND_URL': process.env.BACKEND_URL || '',
  },
  plugins: [react()],
})
