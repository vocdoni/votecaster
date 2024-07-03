import { useCallback, useEffect, useState } from 'react'
import { appUrl } from '~constants'

interface AuthState {
  isAuthenticated: boolean
  bearer: string | null
  bfetch: (input: RequestInfo, init?: RequestInit) => Promise<Response>
  login: (params: LoginParams) => void
  logout: () => void
  profile: Profile | null
  reputation: Reputation | undefined
}

const baseRep = {
  totalReputation: 0,
  activity: {
    followersCount: 0,
    electionsCreated: 0,
    castedVotes: 0,
    participationAchievement: 0,
    communitiesCount: 0,
  },
  boosters: {
    hasVotecasterNFTPass: false,
    hasVotecasterLaunchNFT: false,
    isVotecasterAlphafrensFollower: false,
    isVotecasterFarcasterFollower: false,
    isVocdoniFarcasterFollower: false,
    votecasterAnnouncementRecasted: false,
    hasKIWI: false,
    hasDegenDAONFT: false,
    hasHaberdasheryNFT: false,
    has10kDegenAtLeast: false,
    hasTokyoDAONFT: false,
    hasProxy: false,
    has5ProxyAtLeast: false,
    hasProxyStudioNFT: false,
    hasNameDegen: false,
  },
}

export type Reputation = typeof baseRep

type LoginParams = {
  profile: Profile
  bearer: string
  reputation: Reputation
}

export const useAuthProvider = (): AuthState => {
  const [bearer, setBearer] = useState<string | null>(localStorage.getItem('bearer'))
  const [profile, setProfile] = useState<Profile | null>(JSON.parse(localStorage.getItem('profile') || 'null'))
  const [reputation, setReputation] = useState<Reputation | undefined>(
    JSON.parse(localStorage.getItem('reputation') || '{}')
  )

  const bearedFetch = useCallback(
    async (input: RequestInfo, init: RequestInit = {}) => {
      const headers = new Headers(init.headers || {})
      if (bearer) {
        headers.append('Authorization', `Bearer ${bearer}`)
      }
      const updatedInit = { ...init, headers }
      return fetch(input, updatedInit).then(async (response) => {
        if (!response.ok) {
          const text = await response.text()
          const sanitized = text.replace('\n', '')
          throw new Error(sanitized.length ? sanitized : response.statusText)
        }

        return response
      })
    },
    [bearer]
  )

  const storeReputation = ({ reputation }: { reputation: Reputation }) => {
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
    setReputation(undefined)
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
