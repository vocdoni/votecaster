import { useQuery } from '@tanstack/react-query'
import { useContext } from 'react'
import { chainHealth } from '~queries/healthchecks'
import { getChain } from '~util/chain'
import { HealthcheckContext } from './HealthcheckProvider'

export const useHealthcheck = () => {
  const context = useContext(HealthcheckContext)
  if (!context) {
    throw new Error('useHealthcheck must be used within a HealthcheckProvider')
  }
  return context
}

const useChainHealth = (chainKey: ChainKey) => {
  const chain = getChain(chainKey)
  const { data: isConnected } = useQuery({
    queryKey: ['healthcheck', chainKey],
    queryFn: () => chainHealth(chain.rpcUrls.default.http[0], chain.id),
    refetchInterval: 30000,
    retry: true,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30 * 1000), // exponential back-off
  })

  return { chainKey, isConnected: !!isConnected }
}

export default useChainHealth
