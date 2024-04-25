import { Grid, GridItem, Show, Spacer } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { ReputationCard } from '~components/Auth/Reputation'
import { useAuth } from '~components/Auth/useAuth'
import { Check } from '~components/Check'
import { MutedUsersList } from '~components/MutedUsersList'
import { UserPolls } from '~components/Top'
import { fetchUserPolls } from '~queries/profile'

const Profile = () => {
  const { bfetch, profile } = useAuth()
  // Utilizing React Query to fetch polls
  const { isLoading, error, data } = useQuery<Poll[], Error>({
    queryKey: ['polls'],
    queryFn: fetchUserPolls(bfetch, profile as Profile),
  })

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  return (
    <Grid
      templateColumns={{ base: '1fr', md: '1fr 400px' }} // Stacks on mobile, side by side on wider screens
      templateRows={{ base: 'repeat(3, auto)', md: 'auto 1fr' }} // Creates enough rows for the content on mobile
      gap={4}
      templateAreas={{ base: `"reputation" "muted" "polls"`, md: `"polls reputation" "polls muted"` }}
    >
      <GridItem gridArea='reputation'>
        <ReputationCard />
        <Show above='md'>
          <Spacer h={4} />
          <MutedUsersList />
        </Show>
      </GridItem>
      <GridItem gridArea='polls'>
        <UserPolls polls={data || []} title='Your created polls' w='100%' />
      </GridItem>
      {/* MutedUsersList will now only appear here in the mobile view, since in md+ it's in the same GridItem as ReputationCard */}
      <Show below='md'>
        <GridItem gridArea='muted'>
          <MutedUsersList />
        </GridItem>
      </Show>
    </Grid>
  )
}

export default Profile
