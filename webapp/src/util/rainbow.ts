import { getDefaultConfig } from '@rainbow-me/rainbowkit'
import { degen, localhost } from 'wagmi/chains'

export const config = getDefaultConfig({
  appName: 'farcaster.vote',
  projectId: '735ab19f8bdb36d6ab32328218ded4ac',
  chains: [degen, localhost],
})
