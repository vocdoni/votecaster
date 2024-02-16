import { Flex } from '@chakra-ui/react'
import { Credits } from './Credits'
import Form from './Form'

import '@farcaster/auth-kit/styles.css'

export const App = () => {
  return (
    <Flex
      minH='100vh'
      justifyContent='center'
      alignItems='center'
      py={{ base: 5, xl: 10 }}
      px={{ base: 0, sm: 5, xl: 10 }}
    >
      <Flex maxW={{ base: '100%', lg: '1200px' }} flexDir={{ base: 'column' }}>
        <Form mb={5} />
        <Credits px={{ base: 5, md: 10 }} mb={5} />
      </Flex>
    </Flex>
  )
}
