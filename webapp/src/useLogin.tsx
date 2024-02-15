import { useProfile, useSignIn } from '@farcaster/auth-kit'
import { useEffect } from 'react'

export const useLogin = () => {
  const { isAuthenticated, profile } = useProfile()
  const { signOut } = useSignIn()

  useEffect(() => {
    if (!isAuthenticated) return

    sessionStorage.setItem('profile', JSON.stringify(profile))
  }, [isAuthenticated, profile])

  return {
    profile: Object.values(profile).length ? profile : JSON.parse(sessionStorage.getItem('profile') || '{}'),
    isAuthenticated: isAuthenticated || sessionStorage.getItem('profile') !== null,
    logout: () => {
      signOut()
      sessionStorage.clear()
    },
  }
}
