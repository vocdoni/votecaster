import React, { createContext, ReactNode, useCallback, useState } from 'react'
import HealthChecker from './HealthChecker'

type IHealthcheckContext = Partial<{
  [key in ChainKey]: boolean
}>

export const HealthcheckContext = createContext<IHealthcheckContext | undefined>(undefined)

export const HealthcheckProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [statuses, setStatuses] = useState<IHealthcheckContext>(() => {
    // Initialize all chains as disconnected
    const initialStatuses: Partial<IHealthcheckContext> = {}
    for (const key of Object.keys(import.meta.env.chains) as ChainKey[]) {
      initialStatuses[key] = false
    }
    return initialStatuses as IHealthcheckContext
  })

  const updateStatus = useCallback((key: ChainKey, isConnected: boolean) => {
    setStatuses((prevStatuses) => ({
      ...prevStatuses,
      [key]: isConnected,
    }))
  }, [])

  return (
    <HealthcheckContext.Provider value={statuses}>
      {children}
      {(Object.keys(import.meta.env.chains) as ChainKey[]).map((chainKey) => (
        <HealthChecker key={chainKey} chainKey={chainKey} updateStatus={updateStatus} />
      ))}
    </HealthcheckContext.Provider>
  )
}
