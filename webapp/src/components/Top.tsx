import { Box, BoxProps, Link, Stack, Tag, Text } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { PropsWithChildren } from 'react'
import { Link as RouterLink } from 'react-router-dom'
import { fetchLatestPolls, fetchPollsByVotes, fetchTopCreators, fetchTopVoters } from '~queries/tops'
import { useAuth } from './Auth/useAuth'
import { Check } from './Check'

export const TopTenPolls = (props: BoxProps) => {
  const { bfetch } = useAuth()
  const { data, error, isLoading } = useQuery<Poll[], Error>({
    queryKey: ['topTenPolls'],
    queryFn: fetchPollsByVotes(bfetch),
  })

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  if (!data || !data.length) return null

  return <TopPolls polls={data} title='Top 10 polls (by votes)' {...props} />
}

export const LatestPolls = (props: BoxProps) => {
  const { bfetch } = useAuth()
  const { data, error, isLoading } = useQuery<Poll[], Error>({
    queryKey: ['latestPolls'],
    queryFn: fetchLatestPolls(bfetch),
  })

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  if (!data || !data.length) return null

  return <TopPolls polls={data} title='Latest polls' {...props} />
}

export const TopCreators = (props: BoxProps) => {
  const { bfetch } = useAuth()
  const { data, error, isLoading } = useQuery<UserRanking[], Error>({
    queryKey: ['topCreators'],
    queryFn: fetchTopCreators(bfetch),
  })

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  if (!data || !data.length) return null

  return <TopUsers users={data} title='Top poll creators' {...props} />
}

export const TopVoters = (props: BoxProps) => {
  const { bfetch } = useAuth()
  const { data, error, isLoading } = useQuery<UserRanking[], Error>({
    queryKey: ['topVoters'],
    queryFn: fetchTopVoters(bfetch),
  })

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  if (!data || !data.length) return null

  return <TopUsers users={data} title='Top voters' {...props} />
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
          as={RouterLink}
          to={`/poll/${poll.electionId}`}
          _hover={{
            textDecoration: 'none',
          }}
          style={{ textDecoration: 'none' }}
        >
          <TopCard>
            <Text color='purple.500' fontWeight='medium'>
              {poll.title} — by {poll.createdByDisplayname || poll.createdByUsername}
            </Text>
            <Text color='gray.500' alignSelf={{ base: 'start', sm: 'end' }}>
              {poll.voteCount} votes
            </Text>
          </TopCard>
        </Link>
      ))}
    </Stack>
  </Box>
)

export const UserPolls = ({ polls, title, ...rest }: { polls: Poll[]; title: string } & BoxProps) => (
  <Box bg={'purple.100'} p={5} borderRadius='lg' boxShadow='md' {...rest}>
    <Text fontSize='xl' mb={4} fontWeight='600' color='purple.800'>
      {title || 'Top Polls'}
    </Text>
    <Stack spacing={3}>
      {polls.map((poll, index) => (
        <Link
          key={index}
          as={RouterLink}
          to={`/poll/${poll.electionId}`}
          _hover={{
            textDecoration: 'none',
          }}
          style={{ textDecoration: 'none' }}
        >
          <TopCard>
            <Text color='purple.500' fontWeight='medium' maxW='80%'>
              {poll.title}
            </Text>
            <Text color='gray.500' alignSelf={{ base: 'start', sm: 'end' }}>
              {poll.voteCount} votes
            </Text>
          </TopCard>
        </Link>
      ))}
    </Stack>
  </Box>
)

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
          <TopCard>
            <Text color='purple.500' fontWeight='medium' maxW='80%'>
              {user.displayName || user.username}
            </Text>
            <Tag colorScheme='purple' alignSelf={{ base: 'start', sm: 'end' }} borderRadius='full'>
              {user.count}
            </Tag>
          </TopCard>
        </Link>
      ))}
    </Stack>
  </Box>
)

const TopCard = ({ children }: PropsWithChildren) => {
  return (
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
      {children}
    </Box>
  )
}
