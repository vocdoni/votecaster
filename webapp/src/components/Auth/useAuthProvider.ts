import { useCallback, useEffect, useState } from 'react'
import { appUrl } from '../../util/constants'
import { Profile } from '../types/Profile'

interface AuthState {
  isAuthenticated: boolean
  profile: Profile | null
  login: () => void
  logout: () => void
  bearer: string | null
  bfetch: (input: RequestInfo, init?: RequestInit) => Promise<Response>
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
      return fetch(input, updatedInit)
    },
    [bearer]
  )

  // check if the token is still valid
  useEffect(() => {
    if (!bearer) return

    bearedFetch(`${appUrl}/auth/check`)
      // Invalid token or expired, so logout
      .then((response) => {
        if (response.status !== 200) {
          logout()
        }
      })
      // Handle network errors or other issues
      .catch(() => {
        logout()
      })
  }, [])

  const login = useCallback((profile: Profile, bearer: string) => {
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
