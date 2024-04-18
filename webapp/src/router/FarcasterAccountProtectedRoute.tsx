import { Text, VStack } from '@chakra-ui/react'
import { Outlet } from 'react-router-dom'
import { SignInButton } from '../components/Auth/SignInButton'
import { useAuth } from '../components/Auth/useAuth'

const FarcasterAccountProtectedRoute = () => {
  const { isAuthenticated } = useAuth()

  if (isAuthenticated) {
    return <Outlet />
  }

  return (
    <VStack spacing={4}>
      <Text>You need to sign in first</Text>
      {!isAuthenticated && <SignInButton />}
    </VStack>
  )
}

export default FarcasterAccountProtectedRoute
