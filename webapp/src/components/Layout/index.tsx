import { Flex, Image } from '@chakra-ui/react'
import { Link, Outlet } from 'react-router-dom'
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
        <Flex
          as={Link}
          mt={4}
          fontSize='.8em'
          justifyContent='center'
          to='https://warpcast.com/vocdoni'
          target='_blank'
        >
          <Image src={logo} alt='powered by vocdoni' width='50%' my={6} />
        </Flex>
      </Flex>
    </>
  )
}
