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
  plugins: [react()],
})
