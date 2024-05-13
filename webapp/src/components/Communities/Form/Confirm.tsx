import {
  Alert,
  AlertDescription,
  AlertIcon,
  Box,
  Button,
  ButtonProps,
  Heading,
  Link,
  Progress,
  Text,
} from '@chakra-ui/react'
import { useConnectModal } from '@rainbow-me/rainbowkit'
import { CiWallet } from 'react-icons/ci'
import { GiTopHat } from 'react-icons/gi'
import { MdOutlineRocketLaunch } from 'react-icons/md'
import { useAccount, useBalance, useSwitchChain } from 'wagmi'
import { degen } from 'wagmi/chains'
import { useDegenHealthcheck } from '~components/Healthcheck/use-healthcheck'

type ConfirmProps = {
  price: bigint | null | undefined
  balance: string
} & ButtonProps

export const Confirm = ({ price, ...props }: ConfirmProps) => {
  const { connected } = useDegenHealthcheck()

  if (!connected) {
    return (
      <Alert status='warning'>
        <AlertIcon />
        <AlertDescription>Degenchain is currently down. Please try again later.</AlertDescription>
      </Alert>
    )
  }

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
          <Text>{(Number(price) / 10 ** 18).toString()} $DEGEN</Text>
        </Box>
      )}
      <ConfirmDegenTransactionButton price={price} {...props} />
    </Box>
  )
}

type ConfirmDegenTransactionButtonProps = {
  price: bigint | null | undefined
} & ButtonProps

export const ConfirmDegenTransactionButton = ({ price, ...props }: ConfirmDegenTransactionButtonProps) => {
  const { switchChain } = useSwitchChain()
  const { address, chainId, isConnected } = useAccount()
  const { openConnectModal } = useConnectModal()
  const { data, isLoading, error } = useBalance({ address, chainId: degen.id })

  if (!isConnected) {
    return (
      <Button onClick={openConnectModal} leftIcon={<CiWallet />} {...props}>
        Connect wallet first
      </Button>
    )
  }

  if (chainId !== degen.id) {
    return (
      <Button onClick={() => switchChain({ chainId: degen.id })} leftIcon={<GiTopHat />} {...props}>
        Switch to Degenchain
      </Button>
    )
  }

  if (isLoading) {
    return <Progress isIndeterminate w='full' colorScheme='purple' size='xs' />
  }

  if (error) {
    return (
      <Alert status='error'>
        <AlertDescription>{error.message.toString()}</AlertDescription>
      </Alert>
    )
  }

  if (!price || !data) {
    return
  }

  if (data.value < price) {
    return (
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
  }

  return (
    <Button mt={4} type='submit' rightIcon={<GiTopHat />} leftIcon={<MdOutlineRocketLaunch />} {...props}>
      Deploy your community on Degenchain
    </Button>
  )
}
