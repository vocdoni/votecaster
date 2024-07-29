import { useContext } from 'react'
import { ReputationContext, ReputationContextType } from './ReputationContext'

export const useReputation = (): ReputationContextType => {
  const context = useContext(ReputationContext)
  if (!context) {
    throw new Error('useReputation must be used within a ReputationProvider')
  }
  return context
}
