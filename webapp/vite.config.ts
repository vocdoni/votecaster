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
      'import.meta.env.VOCDONI_DEGENCHAINRPC': JSON.stringify(process.env.VOCDONI_DEGENCHAINRPC || 'https://rpc.degen.tips'),
      'import.meta.env.VOCDONI_COMMUNITYHUBADDRESS': JSON.stringify(process.env.VOCDONI_COMMUNITYHUBADDRESS || '0xC9F73EA7D95174175490b61a5c5eBB024c1f1a88'),
      'import.meta.env.VOCDONI_COMMUNITYRESULTSADDRESS': JSON.stringify(
        process.env.VOCDONI_COMMUNITYRESULTSADDRESS || '0xAcB88BBDc760cA256a8E9254345f633819ba4355'
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
