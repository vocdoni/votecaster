import { Box, Button, Heading, Image, Link, Text, VStack } from '@chakra-ui/react'
import { MdHowToVote } from 'react-icons/md'
import { Link as RouterLink } from 'react-router-dom'
import hat from '/degen-hat.png'

export const Jumbotron = () => (
  <Box boxShadow='md' bg='white' borderRadius='md' p={20} textAlign='center' w='full'>
    <VStack spacing={4} alignItems='center'>
      <Heading as='h1' size='2xl' fontWeight='800' maxW='800px' display='block'>
        The governance platform for your Farcaster community.
      </Heading>
      <Text fontWeight='500' size='2xl' color='gray.600'>
        Kickstart your community
      </Text>
      <RouterLink to='/communities/new'>
        <Button display='flex' gap={2} fontWeight='500'>
          <Box width='1.2rem' height='1.2rem' lineHeight='1'>
            <Image src={hat} />
          </Box>{' '}
          Create your community
        </Button>
      </RouterLink>
      <Text fontStyle='italic' color='gray.400'>
        Experience the farcaster-native governance with your community deployed on Degenchain
        <br />
        <Link as={RouterLink} variant='primary' to='#features'>
          Check all the Features
        </Link>
      </Text>
      <Text fontWeight='500' fontSize='xl' color='gray.600'>
        or run a quick poll within a frame
      </Text>
      <RouterLink to='/form'>
        <Button fontWeight='500' leftIcon={<MdHowToVote />}>
          Create a quick poll
        </Button>
      </RouterLink>
    </VStack>
  </Box>
)
