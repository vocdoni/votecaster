import { Box, Flex } from '@chakra-ui/react'
import { SignInButton } from '@farcaster/auth-kit'
import { Credits } from './Credits'
import Form from './Form'

import '@farcaster/auth-kit/styles.css'
import { useLogin } from './useLogin'

export const App = () => {
  const { isAuthenticated } = useLogin()

  return (
    <Flex
      minH='100vh'
      justifyContent='center'
      alignItems='center'
      py={{ base: 5, xl: 10 }}
      px={{ base: 0, sm: 5, xl: 10 }}
    >
      <Flex maxW={{ base: '100%', lg: '1200px' }} flexDir={{ base: 'column' }}>
        {isAuthenticated ? (
          <Form mb={5} />
        ) : (
          <Box
            minW={{ base: 0, lg: 400 }}
            my={20}
            display='flex'
            justifyContent='center'
            alignItems='center'
            flexDir='column'
          >
            <SignInButton />
            to create a poll
          </Box>
        )}
        <Credits px={{ base: 5, md: 10 }} mb={5} />
      </Flex>
    </Flex>
  )
}
