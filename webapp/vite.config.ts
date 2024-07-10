import react from '@vitejs/plugin-react'
import { Chain } from 'viem'
import { defineConfig, loadEnv, UserConfigFn } from 'vite'
import { createHtmlPlugin } from 'vite-plugin-html'
import svgr from 'vite-plugin-svgr'
import tsconfigPaths from 'vite-tsconfig-paths'
import chainsDefinition from '../chains_config.json'

let explorer = `https://explorer.vote`
const env = process.env.VOCDONI_ENVIRONMENT || 'dev'
if (['dev', 'stg'].includes(env)) {
  explorer = `https://${env}.explorer.vote`
}

type ChainsFile = typeof chainsDefinition
type ChainKey = keyof ChainsFile
type ChainsConfig = Partial<{ [K in ChainKey]: Chain }>

const getConfiguredChains = (chains: ChainKey[]): ChainsConfig => {
  const result: ChainsConfig = {}
  chains.forEach((chain) => {
    const chainConfig = chainsDefinition[chain]
    if (!chainConfig) {
      throw new Error(`Chain ${chain} not found in chains_config.json`)
    }
    result[chain] = chainConfig
  })
  return result
}

// https://vitejs.dev/config/
const viteconfig: UserConfigFn = ({ mode }) => {
  // load env variables from .env files
  process.env = { ...process.env, ...loadEnv(mode, process.cwd(), '') }

  const base = process.env.BASE_URL || '/'
  const outDir = process.env.BUILD_PATH || 'dist'
  const configuredChains: ChainKey[] = JSON.parse(process.env.VOCDONI_CHAINS || 'null') || ['degen-dev', 'base-sep']

  const config = defineConfig({
    base,
    build: {
      outDir,
    },
    define: {
      'import.meta.env.APP_URL': JSON.stringify(process.env.APP_URL || 'https://dev.farcaster.vote'),
      'import.meta.env.VOCDONI_ENVIRONMENT': JSON.stringify(env),
      'import.meta.env.VOCDONI_EXPLORER': JSON.stringify(explorer),
      'import.meta.env.MAINTENANCE': process.env.MAINTENANCE === 'true',
      'import.meta.env.VOCDONI_ADMINFID': parseInt(process.env.ADMINFID || '7548'),
      'import.meta.env.chains': JSON.stringify(getConfiguredChains(configuredChains)),
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
  console.info(config)
  return config
}

export default viteconfig
