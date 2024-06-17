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
  Select,
  Text,
  VStack,
} from '@chakra-ui/react'
import { useConnectModal } from '@rainbow-me/rainbowkit'
import { CiWallet } from 'react-icons/ci'
import { GiTopHat } from 'react-icons/gi'
import { MdOutlineRocketLaunch } from 'react-icons/md'
import { useAccount, useBalance, useSwitchChain } from 'wagmi'
import { degen } from 'wagmi/chains'
import { useDegenHealthcheck } from '~components/Healthcheck/use-healthcheck'
import { chainAlias, isSupportedChain, supportedChains } from '~util/chain'

type ConfirmProps = {
  price: bigint | null | undefined
} & ButtonProps

export const Confirm = ({ price, ...props }: ConfirmProps) => {
  const { connected } = useDegenHealthcheck()
  const { chain, isConnected } = useAccount()
  const { openConnectModal } = useConnectModal()
  const { switchChain } = useSwitchChain()

  if (!isConnected) {
    return (
      <Button onClick={openConnectModal} leftIcon={<CiWallet />} w='full' {...props}>
        Connect wallet first
      </Button>
    )
  }

  if (!chain) {
    return (
      <VStack alignItems='start'>
        <Text>You're on an unsupported chain, please switch to a valid one</Text>
        <ChainSwitcher w='full' />
      </VStack>
    )
  }

  if (chain.id === degen.id && !connected) {
    return (
      <Alert status='warning'>
        <AlertIcon />
        <AlertDescription>Degenchain is currently down. Please try again later, or switch to base.</AlertDescription>
      </Alert>
    )
  }

  return (
    <Box display='flex' gap={4} flexDir='column'>
      <Heading size='sm'>Create your community</Heading>
      <Box display='flex' flexDir='row' alignItems='center'>
        <Text>Your community will be deployed on</Text>
        <Select
          w='auto'
          ml={3}
          value={chain.id}
          onChange={(e) => {
            switchChain({ chainId: Number(e.target.value) })
          }}
        >
          {supportedChains.map((chain) => (
            <option key={chain.id} value={chain.id}>
              {chain.name}
            </option>
          ))}
        </Select>
      </Box>
      <Text>
        As soon as it's created, you will be able to create and manage polls secured by the Vocdoni protocol for
        decentralized, censorship-resistant and gassless voting.
      </Text>
      {typeof price === 'bigint' && (
        <Box display='flex' justifyContent='space-between' fontWeight='500' w='full'>
          <Text>Cost</Text>
          <Text>
            {(Number(price) / 10 ** Number(chain?.nativeCurrency.decimals)).toString()} ${chain?.nativeCurrency.symbol}
          </Text>
        </Box>
      )}
      <ConfirmDegenTransactionButton price={price} {...props} />
    </Box>
  )
}

const ChainSwitcher = (props: ButtonProps) => {
  const { switchChain } = useSwitchChain()
  const first = supportedChains[0]

  return (
    <Button onClick={() => switchChain({ chainId: first.id })} leftIcon={<GiTopHat />} {...props}>
      Switch to {first.name}
    </Button>
  )
}

type ConfirmDegenTransactionButtonProps = {
  price: bigint | null | undefined
} & ButtonProps

export const ConfirmDegenTransactionButton = ({ price, ...props }: ConfirmDegenTransactionButtonProps) => {
  const { address, chainId, chain } = useAccount()
  const { data, isLoading, error } = useBalance({ address, chainId })

  if (!chain || !isSupportedChain(chain)) {
    return <ChainSwitcher {...props} />
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

  if (typeof price !== 'bigint' || !data) {
    return <Progress isIndeterminate size='xs' colorScheme='purple' />
  }

  if (data.value < price) {
    return (
      <>
        <Alert status='warning'>
          <AlertIcon />
          Seems that your wallet account does not have enough founds to deploy the community.
        </Alert>
        {chainId === degen.id && (
          <Text fontSize={'xs'} color={'gray'}>
            You can get some $DEGEN with the{' '}
            <Link fontStyle={'italic'} isExternal href='https://bridge.degen.tips/'>
              Degen Chain Bridge
            </Link>{' '}
            🎩
          </Text>
        )}
      </>
    )
  }

  return (
    <Button
      mt={4}
      type='submit'
      rightIcon={chainAlias(chain) === 'degen' ? <GiTopHat /> : undefined}
      leftIcon={<MdOutlineRocketLaunch />}
      {...props}
    >
      Deploy your community on {chain.name}
    </Button>
  )
}
