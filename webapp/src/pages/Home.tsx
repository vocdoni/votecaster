import { Box, Button, Heading, Spacer, VStack } from '@chakra-ui/react'
import { FaMagnifyingGlass } from 'react-icons/fa6'
import { Link as RouterLink } from 'react-router-dom'
import { FeaturedCommunities } from '~components/Communities/Featured'
import { Features } from '~components/Home/Features'
import { Jumbotron } from '~components/Home/Jumbotron'
import { Tired } from '~components/Home/Tired'
import { LatestPollsSimplified } from '~components/Top'

const Home = () => (
  <VStack w='full' spacing={8}>
    <Jumbotron />
    <HomeSpacer />
    <Box display='flex' flexDir='column' gap={8} alignItems='center' w='full'>
      <FeaturedCommunities />
      <RouterLink to='/communities'>
        <Button size='sm' leftIcon={<FaMagnifyingGlass />} variant='outline'>
          Explore more communities
        </Button>
      </RouterLink>
    </Box>
    <HomeSpacer />
    <Tired />
    <HomeSpacer />
    <VStack spacing={8}>
      <Heading size='lg' textAlign='center' fontWeight={500}>
        Latest 5 polls created
      </Heading>
      <LatestPollsSimplified />
      <RouterLink to='/leaderboards'>
        <Button variant='outline' size='sm' leftIcon={<FaMagnifyingGlass />}>
          Check the leaderboard
        </Button>
      </RouterLink>
    </VStack>
    <HomeSpacer />
    <Features id='features' />
  </VStack>
)

const HomeSpacer = () => <Spacer borderTop='1px solid' borderTopColor='gray.300' my={8} />

export default Home
