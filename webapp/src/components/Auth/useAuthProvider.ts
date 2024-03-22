import { useCallback, useEffect, useState } from 'react'
import { appUrl } from '../../util/constants'
import { Profile } from '../types/Profile'

interface AuthState {
  isAuthenticated: boolean
  bearer: string | null
  bfetch: (input: RequestInfo, init?: RequestInit) => Promise<Response>
  login: (params: LoginParams) => void
  logout: () => void
  profile: Profile | null
  reputation: Reputation | null
}

const baseRep = {
  reputation: 0,
  data: {
    castedVotes: 0,
    electionsCreated: 0,
    followersCount: 0,
    participationAchievement: 0,
  },
}

export type Reputation = typeof baseRep
type ReputationResponse = {
  reputation: number
  reputationData: (typeof baseRep)['data']
}

type LoginParams = {
  profile: Profile
  bearer: string
  reputation: Reputation
}

export const useAuthProvider = (): AuthState => {
  const [bearer, setBearer] = useState<string | null>(localStorage.getItem('bearer'))
  const [profile, setProfile] = useState<Profile | null>(JSON.parse(localStorage.getItem('profile') || 'null'))
  const [reputation, setReputation] = useState<Reputation | null>(
    JSON.parse(localStorage.getItem('reputation') || 'null')
  )

  const bearedFetch = useCallback(
    async (input: RequestInfo, init: RequestInit = {}) => {
      const headers = new Headers(init.headers || {})
      if (bearer) {
        headers.append('Authorization', `Bearer ${bearer}`)
      }
      const updatedInit = { ...init, headers }
      return fetch(input, updatedInit)
    },
    [bearer]
  )

  const storeReputation = (rep: ReputationResponse) => {
    const reputation = {
      reputation: rep.reputation,
      data: {
        ...rep.reputationData,
      },
    }
    setReputation(reputation)
    localStorage.setItem('reputation', JSON.stringify(reputation))
  }

  // if no bearer but profile, logout
  useEffect(() => {
    if (!bearer && !!profile) {
      logout()
    }
  }, [bearer, profile])

  // check if the token is still valid
  useEffect(() => {
    if (!bearer) return

    bearedFetch(`${appUrl}/auth/check`)
      // Invalid token or expired, so logout
      .then(async (response) => {
        if (response.status !== 200) {
          logout()
        }

        return response.json()
      })
      // update reputation
      .then(storeReputation)
      // network errors or other issues
      .catch(() => {
        logout()
      })
  }, [])

  const login = useCallback(({ profile, bearer, reputation }: LoginParams) => {
    localStorage.setItem('bearer', bearer)
    localStorage.setItem('profile', JSON.stringify(profile))
    localStorage.setItem('reputation', JSON.stringify(reputation))
    setBearer(bearer)
    setProfile(profile)
    setReputation(reputation)
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem('bearer')
    localStorage.removeItem('profile')
    localStorage.removeItem('reputation')
    setBearer(null)
    setProfile(null)
    setReputation(null)
  }, [])

  return {
    isAuthenticated: !!bearer && !!profile && !!reputation,
    profile,
    reputation,
    login,
    logout,
    bearer,
    bfetch: bearedFetch,
  }
}
