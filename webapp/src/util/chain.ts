import { ChainContract } from 'viem'
import { Chain } from 'wagmi/chains'
import { explorers } from '~constants'

export const chainAlias = (chain: Chain | string | undefined) => {
  if (!chain) {
    return false
  }

  if (typeof chain === 'string') {
    return chain
  }

  for (const key of Object.keys(import.meta.env.chains)) {
    if (chain.name === key) {
      return key
    }
    if (import.meta.env.chains[key] && chain.id === import.meta.env.chains[key].id) {
      return key
    }
  }

  return false
}

export const supportedChains = Object.values(import.meta.env.chains) as Chain[]

export const isSupportedChain = (chain: Chain | undefined) => {
  if (!chain) {
    return false
  }

  const alias = chainAlias(chain)
  if (!alias) {
    return false
  }

  return typeof import.meta.env.chains[alias] !== 'undefined'
}

export const getChain = (chain: Chain | undefined | string): Chain => {
  const alias = chainAlias(chain)
  if (!chain || !alias) {
    const first = Object.keys(import.meta.env.chains)[0]
    return import.meta.env.chains[first]
  }

  return import.meta.env.chains[alias]
}

export const chainExplorer = (chain: Chain | undefined) => {
  if (!chain) {
    return explorers.degen
  }

  return chain.blockExplorers?.default.url
}

export const getContractForChain = (chain: Chain | undefined | string, contractName = 'communityHub') => {
  if (!chain) {
    return '0x000000000000000000000000000000000000dead'
  }

  if (typeof chain === 'string') {
    chain = getChain(chain)
  }

  if (!chain.contracts || typeof chain.contracts[contractName] === 'undefined') {
    throw new Error(`Contract ${contractName} not found for chain ${chain.name}`)
  }

  return (chain.contracts[contractName] as ChainContract).address
}