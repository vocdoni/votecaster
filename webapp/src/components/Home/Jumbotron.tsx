import { Box, Button, Heading, Link, Text, VStack } from '@chakra-ui/react'
import { MdHowToVote } from 'react-icons/md'
import { Link as RouterLink } from 'react-router-dom'
import { CreateFarcasterCommunityButton } from '~components/Layout/DegenButton'

export const Jumbotron = () => (
  <Box boxShadow='md' bg='white' borderRadius='md' py={20} px={2} textAlign='center' w='full'>
    <VStack spacing={4} alignItems='center'>
      <Heading as='h1' size='jumbo' fontWeight='800' maxW='800px' display='block'>
        The governance platform for your Farcaster community.
      </Heading>
      <Text fontWeight='500' fontSize='xl' color='gray.600'>
        Get started today.
      </Text>
      <CreateFarcasterCommunityButton />
      <Text fontStyle='italic' color='gray.400'>
        Create a community to unlock more census options and enhanced governance features.
        <br />
        <Link as={RouterLink} variant='primary' to='#features'>
          Check all the Features
        </Link>
      </Text>
      <Text fontWeight='500' fontSize='xl' color='gray.600'>
        or ask all Farcaster
      </Text>
      <RouterLink to='/form'>
        <Button fontWeight='500' leftIcon={<MdHowToVote />}>
          Run a 1-click poll within a Frame
        </Button>
      </RouterLink>
    </VStack>
  </Box>
)
