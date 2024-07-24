import { Alert, AlertDescription, AlertIcon, AlertTitle, Button } from '@chakra-ui/react'
import { useLocation } from 'react-router-dom'
import { useAuth } from './useAuth'

export const TokenSignin = () => {
  const { search } = useLocation()
  const { isAuthenticated, error, searchParamsTokenLogin, loading } = useAuth()

  if (isAuthenticated) return null

  return (
    <>
      {error && (
        <Alert status='error'>
          <AlertIcon />
          <AlertTitle>Error signing in via auth token:</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      <Button w='full' isLoading={loading} colorScheme='red' onClick={() => searchParamsTokenLogin(search)}>
        Retry
      </Button>
    </>
  )
}
