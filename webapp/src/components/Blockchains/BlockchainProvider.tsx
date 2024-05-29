import { PropsWithChildren } from 'react'
import { BlockchainContext } from './BlockchainContext'
import { BlockchainProviderProps, useBlockchainProvider } from './use-blockchain-provider'

export type BlockchainState = ReturnType<typeof useBlockchainProvider>

export const BlockchainProvider = ({ children, ...rest }: PropsWithChildren<BlockchainProviderProps>) => {
  const value = useBlockchainProvider(rest)

  return <BlockchainContext.Provider value={value}>{children}</BlockchainContext.Provider>
}
BlockchainProvider.displayName = 'BlockchainProvider'
