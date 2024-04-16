import { getDefaultConfig } from '@rainbow-me/rainbowkit'
import { Chain, localhost } from 'wagmi/chains'

export const degen: Chain = {
  id: 666666666,
  name: 'Degen 🎩',
  rpcUrls: {
    default: {
      http: ['https://rpc.degen.tips'],
    },
  },
  nativeCurrency: {
    name: 'Degen',
    decimals: 18,
    symbol: 'DEGEN',
  },
}

export const config = getDefaultConfig({
  appName: 'farcaster.vote',
  projectId: '735ab19f8bdb36d6ab32328218ded4ac',
  chains: [degen, localhost],
})
