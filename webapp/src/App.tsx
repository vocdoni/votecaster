import { Flex } from '@chakra-ui/react'
import { SignInButton, useProfile } from '@farcaster/auth-kit'
import '@farcaster/auth-kit/styles.css'
import { Credits } from './Credits'
import Form from './Form'

const App = () => {
  const { isAuthenticated } = useProfile()
  return (
    <Flex minH='100vh' justifyContent='center' alignItems='center' p={{ base: 0, sm: 5, xl: 0 }}>
      <Flex maxW={{ base: '100%', lg: '1200px' }} flexDir={{ base: 'column', md: 'row' }}>
        <Credits px={10} mb={5} order={{ base: 1, md: 0 }} />
        {isAuthenticated ? <Form mb={5} order={{ base: 0, md: 1 }} /> : <SignInButton />}
      </Flex>
    </Flex>
  )
}

export default App
