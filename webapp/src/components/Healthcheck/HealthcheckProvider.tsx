import { useQuery, useQueryClient } from '@tanstack/react-query'
import React, { createContext, ReactNode, useEffect } from 'react'
import { degenHealth } from '~queries/healthchecks'

interface IHealthcheckContext {
  degen: {
    connected: boolean
  }
}

export const HealthcheckContext = createContext<IHealthcheckContext | undefined>(undefined)

export const HealthcheckProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const queryClient = useQueryClient()

  const { data: isConnected } = useQuery({
    queryKey: ['healthcheck', 'degen'],
    queryFn: degenHealth,
    refetchInterval: 30000,
    retry: true,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30 * 1000), // exponential back-off
  })

  // ensure we don't overload requests
  useEffect(() => {
    return () => {
      queryClient.invalidateQueries({ queryKey: ['healthcheck', 'degen'] })
    }
  }, [queryClient])

  return (
    <HealthcheckContext.Provider value={{ degen: { connected: !!isConnected } }}>
      {children}
    </HealthcheckContext.Provider>
  )
}
