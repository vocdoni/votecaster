import { Alert, AlertIcon, Box, Flex } from '@chakra-ui/react'
import { Outlet, ScrollRestoration } from 'react-router-dom'
import { Footer } from './Footer'
import { Navbar } from './Navbar'

export const Layout = () => (
  <Box maxW={1920} margin='0 auto'>
    <ScrollRestoration />
    <Navbar />
    {import.meta.env.MAINTENANCE === 'true' && (
      <Alert status='warning'>
        <AlertIcon />
        App is under maintenance, some features may not work as expected.
      </Alert>
    )}
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
