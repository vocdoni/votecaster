import { Button, Flex, Heading, Text } from '@chakra-ui/react'
import { MdHowToVote } from 'react-icons/md'
import { Link as RouterLink } from 'react-router-dom'

export const Tired = () => (
  <Flex
    w='full'
    alignItems='center'
    justifyContent='center'
    flexDir='column'
    textAlign='center'
    gap={8}
    maxW='800px'
    color='gray.600'
    fontSize='lg'
  >
    <Heading as='h2'>Tired of the clunky Web3 governance?</Heading>
    <Text>Until today the Web3 governance experience have been fragmented, centralized and opaque.</Text>
    <Text>
      Votecaster changes that by integrating governance into the Farcaster social feed through Frames, while ensuring
      transparency, end-to-end verifiability, and flexibility, thanks to the use of the Vocdoni Protocol.
    </Text>
    <Text fontWeight='bold'>
      Running a quick poll to engage with the entire Farcaster community only takes 1 minute!
    </Text>
    <RouterLink to='/form'>
      <Button fontWeight='500' leftIcon={<MdHowToVote />}>
        Create a poll now!
      </Button>
    </RouterLink>
  </Flex>
)
