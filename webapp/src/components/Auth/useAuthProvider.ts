import { useCallback, useEffect, useState } from 'react'
import { appUrl } from '~constants'

interface AuthState {
  isAuthenticated: boolean
  bearer: string | null
  bfetch: (input: RequestInfo, init?: RequestInit) => Promise<Response>
  login: (params: LoginParams) => void
  logout: () => void
  profile: Profile | null
}

type LoginParams = {
  profile: Profile
  bearer: string
}

export const useAuthProvider = (): AuthState => {
  const [bearer, setBearer] = useState<string | null>(localStorage.getItem('bearer'))
  const [profile, setProfile] = useState<Profile | null>(JSON.parse(localStorage.getItem('profile') || 'null'))

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
      })
      // network errors or other issues
      .catch(() => {
        logout()
      })
  }, [])

  const login = useCallback(({ profile, bearer }: LoginParams) => {
    localStorage.setItem('bearer', bearer)
    localStorage.setItem('profile', JSON.stringify(profile))
    setBearer(bearer)
    setProfile(profile)
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem('bearer')
    localStorage.removeItem('profile')
    setBearer(null)
    setProfile(null)
  }, [])

  return {
    isAuthenticated: !!bearer && !!profile,
    profile,
    login,
    logout,
    bearer,
    bfetch: bearedFetch,
  }
}
