import { Box, Flex } from '@chakra-ui/react'
import { Outlet, ScrollRestoration } from 'react-router-dom'
import { Footer } from './Footer'
import { Navbar } from './Navbar'

export const Layout = () => (
  <Box maxW={1920} margin='0 auto'>
    <ScrollRestoration />
    <Navbar />
    <Flex
      flexDir='column'
      justifyContent='center'
      alignItems='center'
      p={{ base: 2, sm: 5, xl: 10 }}
      mx='auto'
      maxW='1980px'
    >
      <Outlet />
      <Footer />
    </Flex>
  </Box>
)
