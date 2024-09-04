import { Box, ChakraProvider, Flex, useColorMode } from '@chakra-ui/react'
import { useEffect } from 'react'
import { Outlet, ScrollRestoration } from 'react-router-dom'
import { theme } from '~src/themes/main'
import { Footer } from './Footer'
import { MaintenanceAlert } from './MaintenanceAlert'
import { Navbar } from './Navbar'

export const Layout = () => (
  <ChakraProvider theme={theme}>
    <ForceLightTheme />
    <Box maxW={1920} margin='0 auto'>
      <ScrollRestoration />
      <Navbar />
      <MaintenanceAlert />
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

// We need to force the theme to light to avoid issues switching from the
// composer layout (auto) to the default one (light)
const ForceLightTheme = () => {
  const { setColorMode } = useColorMode()
  useEffect(() => {
    setColorMode('light')
  }, [])

  return null
}
