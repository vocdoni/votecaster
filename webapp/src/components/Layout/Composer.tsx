import { Box, ChakraProvider } from '@chakra-ui/react'
import { Outlet } from 'react-router-dom'
import { composer } from '~src/themes/composer'

export const ComposerLayout = () => (
  <ChakraProvider theme={composer}>
    <Box w='full'>
      <Outlet />
    </Box>
  </ChakraProvider>
)
