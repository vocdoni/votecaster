import { createContext, ReactNode, useState } from 'react'
import { Chain } from 'viem'
import { BlockchainContextState } from './BlockchainContextState'

interface BlockchainRegistryContextState {
  registerContext: (state: BlockchainContextState) => void
  getContext: (id: string) => BlockchainContextState | undefined
}

export const BlockchainRegistryContext = createContext<BlockchainRegistryContextState | undefined>(undefined)

export const BlockchainRegistryProvider = ({ children }: { children: ReactNode }) => {
  const [contexts, setContexts] = useState<Map<string, BlockchainContextState>>(new Map())

  const registerContext = (state: BlockchainContextState) => {
    setContexts(new Map(contexts.set((state.client.chain as Chain).name, state)))
  }

  const getContext = (id: string) => contexts.get(id)

  return (
    <BlockchainRegistryContext.Provider value={{ registerContext, getContext }}>
      {children}
    </BlockchainRegistryContext.Provider>
  )
}
