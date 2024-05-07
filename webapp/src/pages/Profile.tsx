import { Box, Grid, GridItem, Link, Text } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { ReputationCard } from '~components/Auth/Reputation'
import { useAuth } from '~components/Auth/useAuth'
import { ReputationResponse, reputationResponse2Reputation } from '~components/Auth/useAuthProvider'
import { Check } from '~components/Check'
import { FarcasterLogo } from '~components/FarcasterLogo'
import { MutedUsersList } from '~components/MutedUsersList'
import { UserPolls } from '~components/Top'
import { fetchUserProfile } from '~queries/profile'

const Profile = () => {
  const { id } = useParams()
  const { bfetch, profile } = useAuth()
  const username = id || profile?.username
  const isOwnProfile = profile?.username === username
  // Utilizing React Query to fetch polls
  const {
    isLoading,
    error,
    data: user,
  } = useQuery<UserProfileResponse, Error>({
    queryKey: ['profile', username],
    queryFn: fetchUserProfile(bfetch, username as string),
  })

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  return (
    <Grid
      templateColumns={{ base: '1fr', md: '1fr 400px' }} // Stacks on mobile, side by side on wider screens
      templateRows={{ base: 'repeat(3, auto)', md: 'auto 1fr' }} // Creates enough rows for the content on mobile
      gap={4}
      templateAreas={{
        base: `"profile" "muted" "muted" "polls"`,
        md: `"polls profile" "polls muted" "polls muted"`,
      }}
    >
      <GridItem gridArea='profile'>
        <Box boxShadow='md' borderRadius='md' bg='purple.100' p={4} display='flex' flexDir='column' gap={2}>
          <Box display='flex' flexDir='row' gap={2}>
            <Text fontWeight='500'>{user?.user.displayName || user?.user.username}</Text>
            <Link href={`https://warpcast.com/${user?.user.username}`} isExternal>
              <FarcasterLogo fill='purple' />
            </Link>
          </Box>
          <ReputationCard reputation={reputationResponse2Reputation(user as ReputationResponse)} />
        </Box>
      </GridItem>
      <GridItem gridArea='muted'>{isOwnProfile && <MutedUsersList />}</GridItem>
      <GridItem gridArea='polls'>
        <UserPolls
          polls={user?.polls || []}
          title={`${user?.user.displayName || user?.user.username}'s polls`}
          w='100%'
        />
      </GridItem>
    </Grid>
  )
}

export default Profile
