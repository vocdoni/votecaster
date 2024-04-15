import { getDefaultConfig } from '@rainbow-me/rainbowkit'

export const config = getDefaultConfig({
  appName: 'farcaster.vote',
  projectId: '735ab19f8bdb36d6ab32328218ded4ac',
  chains: [
    {
      id: 666666666,
      name: 'Degenchain',
      rpcUrls: {
        default: {
          http: 'https://rpc.degen.tips',
        },
      },
    },
  ],
})
