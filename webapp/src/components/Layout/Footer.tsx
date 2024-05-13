import { Flex, HStack, Icon, Image, Link, Text, VStack } from '@chakra-ui/react'
import { FaDiscord, FaGithub, FaXTwitter } from 'react-icons/fa6'
import { SiFarcaster } from 'react-icons/si'
import { useDegenHealthcheck } from '~components/Healthcheck/use-healthcheck'
import { explorers } from '~constants'
import logo from '/poweredby.svg'

export const Footer = () => (
  <VStack my={24} spacing={12}>
    <Link isExternal href='https://vocdoni.io' width='80%'>
      <Image src={logo} alt='powered by vocdoni' />
    </Link>
    <Flex gap={8} justifyContent='center' color={'gray.600'}>
      <Link isExternal href='https://github.com/vocdoni'>
        <Icon as={FaGithub} />
      </Link>
      <Link isExternal href='https://warpcast.com/vocdoni'>
        <Icon as={SiFarcaster} />
      </Link>
      <Link isExternal href='https://x.com/vocdoni'>
        <Icon as={FaXTwitter} />
      </Link>
      <Link isExternal href='https://chat.vocdoni.io/'>
        <Icon as={FaDiscord} />
      </Link>
    </Flex>
    <HStack fontSize='xs' color='gray.500' fontStyle='italic'>
      <Healthchecks />
    </HStack>
  </VStack>
)

export const Healthchecks = () => {
  const { connected } = useDegenHealthcheck()

  return (
    <>
      <Link isExternal display='flex' flexDir='row' alignItems='center' gap={1} href={explorers.degen}>
        Degenchain RPC status:{' '}
        <Text fontSize='xx-small' fontStyle='normal'>
          {connected ? '🟢' : '🔴'}
        </Text>
      </Link>
    </>
  )
}
