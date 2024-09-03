import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useAuth } from '~components/Auth/useAuth'
import { useReputation } from '~queries/profile'
import { Reputation, ReputationContextType } from './ReputationContext'

export const useReputationProvider = (): ReputationContextType => {
  const { isAuthenticated } = useAuth()
  const queryClient = useQueryClient()
  const [reputation, setReputation] = useState<Reputation | undefined>(
    JSON.parse(localStorage.getItem('reputation') || '{}')
  )

  const { data, error, status } = useReputation()

  useEffect(() => {
    if (isAuthenticated && data) {
      setReputation(data)
      localStorage.setItem('reputation', JSON.stringify(data))
    }
    if (!isAuthenticated) {
      localStorage.removeItem('reputation')
      setReputation(undefined)
      queryClient.removeQueries({ queryKey: ['reputation'] }) // Clear the query if not authenticated
    }
  }, [isAuthenticated, data, queryClient])

  useEffect(() => {
    if (error) {
      console.error('Failed to fetch reputation:', error)
    }
  }, [error])

  return {
    reputation,
    fetchReputation: () => queryClient.invalidateQueries({ queryKey: ['reputation'] }), // Invalidate the query key for refetching
    status,
  }
}
