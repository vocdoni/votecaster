import { Box, Button, Heading, Text } from '@chakra-ui/react'
import { useConnectModal } from '@rainbow-me/rainbowkit'
import { CiWallet } from 'react-icons/ci'
import { GiTopHat } from 'react-icons/gi'
import { MdOutlineRocketLaunch } from 'react-icons/md'
import { useAccount } from 'wagmi'

export const Confirm = () => {
  const { isConnected } = useAccount()
  const { openConnectModal } = useConnectModal()

  return (
    <Box display='flex' gap={4} flexDir='column'>
      <Heading size='sm'>Create your community</Heading>
      <Text>Your community will be deployed on the Degenchain.</Text>
      <Text>
        As soon as it's created, you will be able to create and manage polls secured by the Vocdoni protocol for
        decentralized, censorship-resistant and gassless voting.
      </Text>
      <Box display='flex' justifyContent='space-between' fontWeight='500' w='full'>
        <Text>Cost</Text>
        <Text>1000 $DEGEN</Text>
      </Box>
      {isConnected ? (
        <Button mt={4} colorScheme='blue' type='submit' rightIcon={<GiTopHat />} leftIcon={<MdOutlineRocketLaunch />}>
          Deploy your community on Degenchain
        </Button>
      ) : (
        <Button onClick={openConnectModal} colorScheme='blue' leftIcon={<CiWallet />}>
          Connect wallet first
        </Button>
      )}
    </Box>
  )
}
