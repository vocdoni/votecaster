import { useCallback, useEffect, useState } from 'react'
import { useAuth } from '~components/Auth/useAuth'
import { appUrl } from '~constants'
import { Reputation, ReputationContextType } from './ReputationContext'

export const useReputationProvider = (): ReputationContextType => {
  const { bearer, bfetch, isAuthenticated } = useAuth()
  const [reputation, setReputation] = useState<Reputation | undefined>(
    JSON.parse(localStorage.getItem('reputation') || '{}')
  )

  const storeReputation = ({ reputation }: { reputation: Reputation }) => {
    setReputation(reputation)
    localStorage.setItem('reputation', JSON.stringify(reputation))
  }

  const fetchReputation = useCallback(async () => {
    if (!bearer) return

    try {
      const response = await bfetch(`${appUrl}/profile`)
      if (response.ok) {
        const data = await response.json()
        storeReputation(data)
      } else {
        throw new Error('Failed to fetch reputation')
      }
    } catch (error) {
      console.error(error)
    }
  }, [bearer, bfetch])

  useEffect(() => {
    if (isAuthenticated) {
      fetchReputation()
    } else {
      localStorage.removeItem('reputation')
      setReputation(undefined)
    }
  }, [isAuthenticated, fetchReputation])

  return { reputation, fetchReputation }
}
