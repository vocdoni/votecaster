import { Box, BoxProps, Link, Spinner, Stack, Tag, Text } from '@chakra-ui/react'
import { useQuery } from 'react-query'
import { fetchPollsByVotes, fetchTopCreators, fetchTopVoters, UserRanking } from '../queries/tops'
import { useAuth } from './Auth/useAuth'
import { Check } from './Check'

const appUrl = import.meta.env.APP_URL

export type Poll = {
  title: string
  createdByUsername: string
  voteCount: number
}

export const TopTenPolls = (props: BoxProps) => {
  const { bfetch } = useAuth()
  const { data: polls, error, isLoading } = useQuery<Poll[], Error>('topTenPolls', fetchPollsByVotes(bfetch))

  if (isLoading) return <Spinner />

  if (error) {
    console.error('Error fetching top 10 polls:', error)
    return <div>Error fetching polls</div>
  }

  if (!polls || !polls.length) return null

  return <TopPolls polls={polls} title='Top 10 polls (by votes)' {...props} />
}

export const TopPolls = ({ polls, title, ...rest }: { polls: Poll[]; title: string } & BoxProps) => (
  <Box bg={'purple.100'} p={5} borderRadius='lg' boxShadow='md' {...rest}>
    <Text fontSize='xl' mb={4} fontWeight='600' color='purple.800'>
      {title || 'Top Polls'}
    </Text>
    <Stack spacing={3}>
      {polls.map((poll, index) => (
        <Link
          key={index}
          href={`https://warpcast.com/polls/${poll.id}`}
          isExternal
          _hover={{
            textDecoration: 'none',
          }}
          style={{ textDecoration: 'none' }}
        >
          <Box
            p={3}
            bg={'white'}
            borderRadius='md'
            display='flex'
            justifyContent='space-between'
            flexDir={{ base: 'column', sm: 'row' }}
            gap={{ base: 0, sm: 2 }}
            boxShadow='sm'
            border='1px'
            borderColor='purple.200'
            _hover={{
              bg: 'purple.50',
            }}
          >
            <Text color='purple.500' fontWeight='medium' maxW='80%'>
              {poll.title} — by {poll.createdByUsername}
            </Text>
            <Text color='gray.500' alignSelf={{ base: 'start', sm: 'end' }}>
              {poll.voteCount} votes
            </Text>
          </Box>
        </Link>
      ))}
    </Stack>
  </Box>
)

export const TopCreators = (props: BoxProps) => {
  const { bfetch } = useAuth()
  const { data, error, isLoading } = useQuery<Poll[], Error>('topCreators', fetchTopCreators(bfetch))

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  if (!data || !data.length) return null

  return <TopUsers users={data} title='Top poll creators' {...props} />
}

export const TopVoters = (props: BoxProps) => {
  const { bfetch } = useAuth()
  const { data, error, isLoading } = useQuery<Poll[], Error>('topVoters', fetchTopVoters(bfetch))

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  if (!data || !data.length) return null

  return <TopUsers users={data} title='Top voters' {...props} />
}

export const TopUsers = ({ users, title, ...rest }: { users: UserRanking[]; title: string } & BoxProps) => (
  <Box bg={'purple.100'} p={5} borderRadius='lg' boxShadow='md' {...rest}>
    <Text fontSize='xl' mb={4} fontWeight='600' color='purple.800'>
      {title || 'Top Users'}
    </Text>
    <Stack spacing={3}>
      {users.map((user, index) => (
        <Link
          key={index}
          href={`https://warpcast.com/${user.username}`}
          isExternal
          _hover={{
            textDecoration: 'none', // Prevents the default underline on hover
          }}
          style={{ textDecoration: 'none' }} // Ensures that the Link does not have an underline
        >
          <Box
            p={3}
            bg={'white'}
            borderRadius='md'
            display='flex'
            justifyContent='space-between'
            gap={{ base: 0, sm: 2 }}
            boxShadow='sm'
            border='1px'
            borderColor='purple.200'
            _hover={{
              bg: 'purple.50', // Light hover effect for the box
            }}
          >
            <Text
              color='purple.500'
              fontWeight='medium'
              maxW='80%'
              isTruncated // Ensures text does not overflow
            >
              {user.displayName || user.username}
            </Text>
            <Tag colorScheme='purple' alignSelf={{ base: 'start', sm: 'end' }} borderRadius='full'>
              {user.count}
            </Tag>
          </Box>
        </Link>
      ))}
    </Stack>
  </Box>
)
