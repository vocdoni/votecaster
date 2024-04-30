import { Text, VStack } from '@chakra-ui/react'
import { ConnectButton } from '@rainbow-me/rainbowkit'
import { Outlet } from 'react-router-dom'
import { useAccount } from 'wagmi'
import { SignInButton } from '~components/Auth/SignInButton'
import { useAuth } from '~components/Auth/useAuth'

const WalletProtectedRoute = () => {
  const { isAuthenticated } = useAuth()
  const { isConnected } = useAccount()

  if (isConnected && isAuthenticated) {
    return <Outlet />
  }

  return (
    <VStack spacing={4}>
      <Text>You need both farcaster and a wallet connected</Text>
      {!isAuthenticated && <SignInButton />}
      {!isConnected && <ConnectButton />}
    </VStack>
  )
}

export default WalletProtectedRoute
