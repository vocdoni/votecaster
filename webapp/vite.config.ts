import react from '@vitejs/plugin-react'
import {defineConfig, loadEnv} from 'vite'
import {createHtmlPlugin} from 'vite-plugin-html'
import svgr from 'vite-plugin-svgr'

// https://vitejs.dev/config/
const viteconfig = ({mode}) => {
  // load env variables from .env files
  process.env = {...process.env, ...loadEnv(mode, process.cwd(), '')}

  const base = process.env.BASE_URL || '/'
  const outDir = process.env.BUILD_PATH || 'dist'


  return defineConfig({
    base,
    build: {
      outDir,
    },
    define: {
      'import.meta.env.APP_URL': JSON.stringify(process.env.APP_URL || ''),
      'import.meta.env.DEGEN_CONTRACT_ADDRESS': JSON.stringify(process.env.DEGEN_CONTRACT_ADDRESS || '0xd4768df803c5a9eDA475159cfbBcF9c06F077c13'),
      'import.meta.env.RESULTS_CONTRACT_ADDRESS': JSON.stringify(
        process.env.RESULTS_CONTRACT_ADDRESS || '0x56c24fedad2C98C89830A2Ac74e9B70e1E2ca042'
      ),
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
}

export default viteconfig
