import { Flex } from '@chakra-ui/react'
import Form from '../components/Form'

export const App = () => {
  return (
    <Flex maxW={{ base: '100%', lg: '1200px' }} flexDir={{ base: 'column' }} gap={8}>
      <Form justifyContent='center' />
    </Flex>
  )
}
