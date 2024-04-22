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


  const config = defineConfig({
    base,
    build: {
      outDir,
    },
    define: {
      'import.meta.env.APP_URL': JSON.stringify(process.env.APP_URL || 'https://dev.farcaster.vote'),
      'import.meta.env.VOCDONI_DEGENCHAINRPC': JSON.stringify(process.env.VOCDONI_DEGENCHAINRPC || 'https://rpc.degen.tips'),
      'import.meta.env.VOCDONI_COMMUNITYHUBADDRESS': JSON.stringify(process.env.VOCDONI_COMMUNITYHUBADDRESS || '0x013320A5E5052E05CEE868A78b9C5C4708d26217'),
      'import.meta.env.VOCDONI_COMMUNITYRESULTSADDRESS': JSON.stringify(
        process.env.VOCDONI_COMMUNITYRESULTSADDRESS || '0x904FeEd341410bfA457023d737CC9928DB95f90b'
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
  console.log(config)
  return config
}

export default viteconfig
