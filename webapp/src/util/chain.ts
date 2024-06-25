import { base, baseSepolia, Chain, degen } from 'wagmi/chains'
import { explorers } from '~constants'
import { camelize } from './strings'

export const allChains: { [key: string]: Chain } = {
  degen,
  baseSepolia,
  base,
}

export const chainAlias = (chain: Chain | undefined | string) => {
  if (!chain) {
    return 'degen'
  }

  if (typeof chain === 'string') {
    if (chain === 'basesep') {
      return 'baseSepolia'
    }
    return chain
  }

  return camelize(chain.name)
}

export const supportedChains = Object.keys(import.meta.env.COMMUNITY_HUB_ADDRESSES).map((chain) => allChains[chain])

export const isSupportedChain = (chain: Chain | undefined) => {
  if (!chain) {
    return false
  }

  return typeof import.meta.env.COMMUNITY_HUB_ADDRESSES[chainAlias(chain)] !== 'undefined'
}

export const getChain = (chain: Chain | string | undefined) => {
  if (!chain) {
    return degen
  }

  return allChains[chainAlias(chain)]
}

export const chainExplorer = (chain: Chain | undefined) => {
  if (!chain) {
    return explorers.degen
  }

  return chain.blockExplorers?.default.url
}

export const getContractForChain = (chain: Chain | undefined | string) => {
  if (!chain) {
    return '0x000000000000000000000000000000000000dead'
  }

  return import.meta.env.COMMUNITY_HUB_ADDRESSES[chainAlias(chain)]
}
