import { Box, Button, Heading, Text } from '@chakra-ui/react'

export const Confirm = () => {
  return (
    <Box display='flex' gap={4} flexDir='column'>
      <Heading size='sm'>Create your community</Heading>
      <Text>Your community will be deployed on the Degenchain.</Text>
      <Text>
        As soon as it's created, you will be able to create and manage polls secured by the Vocdoni protocol for
        decentralized, censorship-resistant and gassless voting.
      </Text>
      <Box display='flex' justifyContent='space-between' fontWeight='500' w='full'>
        <Text>Cost</Text>
        <Text>1000 $DEGEN</Text>
      </Box>
      <Button mt={4} colorScheme='blue' type='submit'>
        Deploy your community on Degenchain
      </Button>
    </Box>
  )
}
