import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'
import { createHtmlPlugin } from 'vite-plugin-html'
import svgr from 'vite-plugin-svgr'

const base = process.env.BASE_URL || '/'
const outDir = process.env.BUILD_PATH || 'dist'

// https://vitejs.dev/config/
export default defineConfig({
  base,
  build: {
    outDir,
  },
  define: {
    'import.meta.env.APP_URL': JSON.stringify(process.env.APP_URL || ''),
  },
  plugins: [
    svgr(),
    react(),
    createHtmlPlugin({
      minify: true,
      inject: {
        data: {
          baseUrl: base.replace(/\/$/, ''),
        },
      },
    }),
  ],
})
