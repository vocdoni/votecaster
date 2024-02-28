import { Flex } from '@chakra-ui/react'
import { Credits } from './Credits'
import Form from './Form'

import '@farcaster/auth-kit/styles.css'
import { TopTenPolls } from './Top'

export const App = () => {
  return (
    <Flex
      minH='100vh'
      justifyContent='center'
      alignItems='center'
      py={{ base: 5, xl: 10 }}
      px={{ base: 0, sm: 5, xl: 10 }}
    >
      <Flex maxW={{ base: '100%', lg: '1200px' }} flexDir={{ base: 'column' }} gap={8}>
        <Form justifyContent='center' />
        <Credits px={{ base: 5, md: 10 }} justifyContent='center' />
        <TopTenPolls mx={{ base: 0, md: 10 }} />
      </Flex>
    </Flex>
  )
}
