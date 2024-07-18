import { PropsWithChildren } from 'react'
import { ReputationContext } from './ReputationContext'
import { useReputationProvider } from './useReputationProvider'

export const ReputationProvider = (props: PropsWithChildren) => {
  const value = useReputationProvider()

  return <ReputationContext.Provider value={value} {...props} />
}
