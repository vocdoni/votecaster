import { createContext } from 'react'
import { Chain } from 'viem'
import { BlockchainContextState } from './BlockchainContextState'
import { useBlockchainRegistry } from './use-blockchain-registry'

export const BlockchainContext = createContext<BlockchainContextState | undefined>(undefined)

export const useBlockchain = (chain: string | Chain) => {
  const { getContext } = useBlockchainRegistry()
  const context = getContext(typeof chain === 'string' ? chain : chain.name)
  if (!context) {
    throw new Error(`useBlockchain returned undefined for id "${chain}", make sure the provider is registered.`)
  }
  return context
}
