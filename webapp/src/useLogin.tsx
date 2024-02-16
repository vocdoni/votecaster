import { useProfile, useSignIn } from '@farcaster/auth-kit'
import { useEffect } from 'react'

export const useLogin = () => {
  const { isAuthenticated, profile } = useProfile()
  const { signOut } = useSignIn()

  useEffect(() => {
    if (!isAuthenticated) return

    localStorage.setItem('profile', JSON.stringify(profile))
  }, [isAuthenticated, profile])

  return {
    profile: Object.values(profile).length ? profile : JSON.parse(localStorage.getItem('profile') || '{}'),
    isAuthenticated: isAuthenticated || localStorage.getItem('profile') !== null,
    logout: () => {
      signOut()
      localStorage.clear()
    },
  }
}
