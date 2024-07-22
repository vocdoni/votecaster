import { Alert, AlertIcon, Box, ChakraProvider, Flex } from '@chakra-ui/react'
import { Outlet, ScrollRestoration } from 'react-router-dom'
import { theme } from '~src/themes/main'
import { Footer } from './Footer'
import { Navbar } from './Navbar'

export const Layout = () => (
  <ChakraProvider theme={theme}>
    <Box maxW={1920} margin='0 auto'>
      <ScrollRestoration />
      <Navbar />
      {import.meta.env.MAINTENANCE && (
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
  </ChakraProvider>
)
