import { useContext } from 'react'
import { BlockchainRegistryContext } from './BlockchainRegistry'

export const useBlockchainRegistry = () => {
  const context = useContext(BlockchainRegistryContext)
  if (!context) {
    throw new Error('useBlockchainRegistry must be used within a BlockchainRegistryProvider')
  }
  return context
}
