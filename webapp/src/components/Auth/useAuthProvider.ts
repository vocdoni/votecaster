import { useCallback, useEffect, useMemo, useState } from 'react'
import { appUrl } from '~constants'
import { userToProfile } from '~util/mappings'

export type AuthState = ReturnType<typeof useAuthProvider>

type LoginParams = {
  profile: Profile
  bearer: string
}

export const useAuthProvider = () => {
  const [bearer, setBearer] = useState<string | null>(localStorage.getItem('bearer'))
  const [profile, setProfile] = useState<Profile | null>(JSON.parse(localStorage.getItem('profile') || 'null'))
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)

  const isAuthenticated = useMemo(() => !!bearer && !!profile, [bearer, profile])

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

  const tokenLogin = useCallback((token: string) => {
    setError(null)
    setLoading(true)
    return bearedFetch(`${appUrl}/profile`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
      .then((resp) => resp.json())
      .then(({ user }: UserProfileResponse) =>
        login({
          profile: userToProfile(user),
          bearer: token,
        })
      )
      .catch((err) => {
        setError(err.message)
      })
      .finally(() => setLoading(false))
  }, [])

  const searchParamsTokenLogin = useCallback(
    (search: string) => {
      const params = new URLSearchParams(search.replace(/^\?/, ''))
      const token = params.get('token')

      if (!token || isAuthenticated) return

      tokenLogin(token)
    },
    [isAuthenticated]
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
    bearer,
    bfetch: bearedFetch,
    error,
    isAuthenticated,
    loading,
    login,
    logout,
    profile,
    searchParamsTokenLogin,
    tokenLogin,
  }
}
