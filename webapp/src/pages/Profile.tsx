import { Avatar, Box, Link, Text, VStack } from '@chakra-ui/react'
import { useQuery } from 'react-query'
import { ReputationCard } from '../components/Auth/Reputation'
import { useAuth } from '../components/Auth/useAuth'
import { Check } from '../components/Check'
import { Poll, TopPolls } from '../components/Top'
import { fetchMutedUsers, fetchUserPolls } from '../queries/profile'

export const Profile = () => {
  const { bfetch, profile } = useAuth()
  // Utilizing React Query to fetch polls
  const { isLoading, error, data } = useQuery<Poll[], Error>('polls', fetchUserPolls(bfetch, profile))

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  return (
    <VStack direction='column' spacing={4}>
      <ReputationCard />
      <TopPolls polls={data || []} title='Your created polls' w='100%' />
      <MutedUsersList />
    </VStack>
  )
}

const MutedUsersList: React.FC = () => {
  const { bfetch } = useAuth()
  const { data, error, isLoading } = useQuery<Profile[], Error>('mutedUsers', fetchMutedUsers(bfetch))

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  return (
    <VStack spacing={4}>
      {data?.map((user) => (
        <Box key={user.fid} padding='4' boxShadow='lg' borderRadius='md'>
          <Link href={`https://warpcast.com/${user.username}`} isExternal>
            <Avatar src={user.pfpUrl} name={user.username} />
            <Text mt='2'>{user.username}</Text>
          </Link>
        </Box>
      ))}
    </VStack>
  )
}
