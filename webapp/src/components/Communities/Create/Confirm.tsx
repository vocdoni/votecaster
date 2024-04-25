import { Alert, AlertIcon, Box, Button, ButtonProps, Heading, Link, Text } from '@chakra-ui/react'
import { useConnectModal } from '@rainbow-me/rainbowkit'
import { useMemo } from 'react'
import { CiWallet } from 'react-icons/ci'
import { GiTopHat } from 'react-icons/gi'
import { MdOutlineRocketLaunch } from 'react-icons/md'
import { useAccount } from 'wagmi'

type ConfirmProps = {
  price: string
  balance: string
} & ButtonProps

export const Confirm = ({ price, balance, ...props }: ConfirmProps) => {
  const { isConnected } = useAccount()
  const { openConnectModal } = useConnectModal()
  const enoughBalance = useMemo(() => parseFloat(balance) > parseFloat(price), [balance, price])

  return (
    <Box display='flex' gap={4} flexDir='column'>
      <Heading size='sm'>Create your community</Heading>
      <Text>Your community will be deployed on the Degenchain.</Text>
      <Text>
        As soon as it's created, you will be able to create and manage polls secured by the Vocdoni protocol for
        decentralized, censorship-resistant and gassless voting.
      </Text>
      {!!price && (
        <Box display='flex' justifyContent='space-between' fontWeight='500' w='full'>
          <Text>Cost</Text>
          <Text>{price} $DEGEN</Text>
        </Box>
      )}
      {isConnected ? (
        enoughBalance ? (
          <Button
            mt={4}
            colorScheme='blue'
            type='submit'
            rightIcon={<GiTopHat />}
            leftIcon={<MdOutlineRocketLaunch />}
            {...props}
          >
            Deploy your community on Degenchain
          </Button>
        ) : (
          <>
            <Alert status='warning'>
              <AlertIcon />
              Seems that your wallet account does not have enough founds to deploy the community.
            </Alert>
            <Text fontSize={'xs'} color={'gray'}>
              You can get some $DEGEN with the{' '}
              <Link fontStyle={'italic'} isExternal href='https://bridge.degen.tips/'>
                Degen Chain Bridge
              </Link>{' '}
              🎩
            </Text>
          </>
        )
      ) : (
        <Button onClick={openConnectModal} colorScheme='blue' leftIcon={<CiWallet />} {...props}>
          Connect wallet first
        </Button>
      )}
    </Box>
  )
}
