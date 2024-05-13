import { useContext } from 'react'
import { HealthcheckContext } from './HealthcheckProvider'

export const useHealthcheck = () => {
  const context = useContext(HealthcheckContext)
  if (!context) {
    throw new Error('useHealthcheck must be used within a HealthcheckProvider')
  }
  return context
}

export const useDegenHealthcheck = () => {
  const context = useContext(HealthcheckContext)
  if (!context) {
    throw new Error('useHealthcheck must be used within a HealthcheckProvider')
  }
  return context.degen
}
