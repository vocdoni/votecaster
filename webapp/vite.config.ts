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

type ContractAddresses = {
  degen: string
  base: string
  [key: string]: string
}

const alias = (str: string) => {
  if (str === 'basesep') {
    return 'baseSepolia'
  }
  return str.toLowerCase().replace(/[^a-zA-Z0-9]+(.)/g, (_, chr) => chr.toUpperCase())
}

const parseEnvVars = (envVar: string): ContractAddresses => {
  const result: { [key: string]: string } = {}
  const pairs = envVar.split(',')

  pairs.forEach((pair) => {
    const [chain, value] = pair.split(':')
    if (!chain || !value) {
      throw new Error(`Invalid format for pair: ${pair}`)
    }
    result[alias(chain)] = value
  })

  // Ensure 'degen' and 'base' are present
  if (!result.degen || !(result.base || result.baseSepolia)) {
    throw new Error('Both "degen" and "base" contract addresses must be provided')
  }

  // Return the result as ContractAddresses type
  return result as ContractAddresses
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
      'import.meta.env.COMMUNITY_HUB_ADDRESSES': JSON.stringify(
        parseEnvVars(
          process.env.VOCDONI_COMMUNITY_HUB_ADDRESSES ||
            'degen:0x1Be05fD83B43D3d5Eb930Ab44f326Fe69d63bd63,basesep:0xdB5a0d05788A7D94026286951301545082C2A088'
        )
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
