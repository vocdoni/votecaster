import { Box, BoxProps, Link, Spinner, Stack, Text } from '@chakra-ui/react'
import { useQuery } from 'react-query'
import { fetchPollsByVotes } from '../queries/tops'
import { useAuth } from './Auth/useAuth'

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
  <Box bg={'gray.800'} p={5} borderRadius='lg' {...rest}>
    <Text fontSize='xl' mb={4} fontWeight='600' color='white'>
      {title || 'Top Polls'}
    </Text>
    <Stack spacing={3}>
      {polls.map((poll, index) => (
        <Box
          key={index}
          p={3}
          bg={'gray.700'}
          borderRadius='md'
          display='flex'
          justifyContent='space-between'
          flexDir={{ base: 'column', sm: 'row' }}
          gap={{ base: 0, sm: 2 }}
        >
          <Link
            href={`https://warpcast.com/${poll.createdByUsername}`}
            isExternal
            color='teal.300'
            fontWeight='medium'
            maxW='80%'
          >
            {poll.title} — by {poll.createdByUsername}
          </Link>
          <Text color='gray.200' alignSelf='end'>
            {poll.voteCount} votes
          </Text>
        </Box>
      ))}
    </Stack>
  </Box>
)
