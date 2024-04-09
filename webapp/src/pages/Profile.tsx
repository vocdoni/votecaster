import { Avatar, Box, Link, Spinner, Text, VStack } from '@chakra-ui/react'
import { useQuery } from 'react-query'
import { ReputationCard } from '../components/Auth/Reputation'
import { useAuth } from '../components/Auth/useAuth'
import { Poll, TopPolls } from '../components/Top'
import { fetchMutedUsers, fetchUserPolls } from '../queries/profile'

export const Profile = () => {
  const { bfetch, profile } = useAuth()
  // Utilizing React Query to fetch polls
  const pollsQuery = useQuery<Poll[], Error>('polls', fetchUserPolls(bfetch, profile))

  if (pollsQuery.isLoading) return <Spinner />

  if (pollsQuery.isError) {
    console.error('Error fetching polls:', pollsQuery.error)
    // You can return an error message or any UI to reflect the error
    return <div>Error fetching polls</div>
  }

  return (
    <VStack direction='column' spacing={4}>
      <ReputationCard />
      <TopPolls polls={pollsQuery.data || []} title='Your created polls' w='100%' />
      <MutedUsersList />
    </VStack>
  )
}

const MutedUsersList: React.FC = () => {
  const { bfetch } = useAuth()
  const { data: mutedUsers, error, isLoading } = useQuery<Profile[], Error>('mutedUsers', fetchMutedUsers(bfetch))

  if (isLoading) {
    return <div>Loading muted users list...</div>
  }

  if (error) {
    return <Alert status='warning'>An error has occurred: {error.message}</Alert>
  }

  return (
    <VStack spacing={4}>
      {mutedUsers?.map((user) => (
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
