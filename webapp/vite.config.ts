import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv, UserConfigFn } from 'vite'
import { createHtmlPlugin } from 'vite-plugin-html'
import svgr from 'vite-plugin-svgr'
import tsconfigPaths from 'vite-tsconfig-paths'

let explorer = `https://explorer.vote`
const env = process.env.VOCDONI_ENVIRONMENT || 'dev'
if (['dev', 'stg'].includes(env)) {
  explorer = `https://${env}.explorer.vote`
}

// https://vitejs.dev/config/
const viteconfig: UserConfigFn = ({ mode }) => {
  // load env variables from .env files
  process.env = { ...process.env, ...loadEnv(mode, process.cwd(), '') }

  const base = process.env.BASE_URL || '/'
  const outDir = process.env.BUILD_PATH || 'dist'

  const config = defineConfig({
    base,
    build: {
      outDir,
    },
    define: {
      'import.meta.env.APP_URL': JSON.stringify(process.env.APP_URL || 'https://dev.farcaster.vote'),
      'import.meta.env.VOCDONI_ENVIRONMENT': JSON.stringify(env),
      'import.meta.env.VOCDONI_EXPLORER': JSON.stringify(explorer),
      'import.meta.env.MAINTENANCE': JSON.stringify(process.env.MAINTENANCE || false),
      'import.meta.env.VOCDONI_DEGENCHAINRPC': JSON.stringify(
        process.env.VOCDONI_DEGENCHAINRPC || 'https://rpc.degen.tips'
      ),
      'import.meta.env.VOCDONI_COMMUNITYHUBADDRESS': JSON.stringify(
        process.env.VOCDONI_COMMUNITYHUBADDRESS || '0xC6d3ae00a9c2322dE48B63053e989E7E2e6C2cc9'
      ),
      'import.meta.env.VOCDONI_ADMINFID': parseInt(process.env.ADMINFID || '7548'),
    },
    plugins: [
      tsconfigPaths(),
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
  console.log(config)
  return config
}

export default viteconfig
