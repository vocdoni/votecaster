import { Flex, Icon, Image, Link, VStack } from '@chakra-ui/react'
import { FaDiscord, FaGithub, FaXTwitter } from 'react-icons/fa6'
import { SiFarcaster } from 'react-icons/si'
import { Outlet } from 'react-router-dom'
import { Navbar } from './Navbar'
import logo from '/poweredby.svg'

export const Layout = () => {
  return (
    <>
      <Navbar />
      <Flex
        flexDir='column'
        justifyContent='center'
        alignItems='center'
        p={{ base: 0, sm: 5, xl: 10 }}
        mx='auto'
        maxW='1980px'
      >
        <Outlet />
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
        </VStack>
      </Flex>
    </>
  )
}
