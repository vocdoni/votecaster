import { Flex } from '@chakra-ui/react'
import { Credits } from './Credits'
import Form from './Form'
import { TopTenPolls } from './Top'

export const App = () => {
  return (
    <Flex maxW={{ base: '100%', lg: '1200px' }} flexDir={{ base: 'column' }} gap={8}>
      <Form justifyContent='center' />
      <Credits px={{ base: 5, md: 10 }} justifyContent='center' />
      <TopTenPolls mx={{ base: 0, md: 10 }} />
    </Flex>
  )
}
