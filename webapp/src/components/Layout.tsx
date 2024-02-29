import { Flex } from '@chakra-ui/react'
import { Outlet } from 'react-router-dom'

import '@farcaster/auth-kit/styles.css'

export const Layout = () => {
  return (
    <Flex minH='100vh' justifyContent='center' alignItems='center' p={{ base: 0, sm: 5, xl: 10 }}>
      <Outlet />
    </Flex>
  )
}
