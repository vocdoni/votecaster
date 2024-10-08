import { Box, Grid, GridItem, Heading, Link, Text, VStack } from '@chakra-ui/react'
import { useParams } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'
import { Check } from '~components/Check'
import { Delegations } from '~components/Delegations'
import { FarcasterLogo } from '~components/FarcasterLogo'
import { PurpleBox } from '~components/Layout/PurpleBox'
import { MutedUsersList } from '~components/MutedUsersList'
import { ReputationCard } from '~components/Reputation/Reputation'
import { UserPolls } from '~components/Top'
import { WarpcastApiKey } from '~components/WarpcastApiKey'
import { useUserDegenOrEnsName, useUserProfile } from '~queries/profile'

const Profile = () => {
  const { id } = useParams()
  const { profile } = useAuth()
  const username = id || profile?.username
  const isOwnProfile = profile?.username === username
  const { isLoading, error, data: user } = useUserProfile(id)
  const { data: degenOrEnsName } = useUserDegenOrEnsName(user)

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  return (
    <Grid
      templateColumns={{ base: '1fr', md: '1fr 400px' }} // Stacks on mobile, side by side on wider screens
      templateRows={{ base: 'repeat(3, auto)', md: 'auto 1fr' }} // Creates enough rows for the content on mobile
      gap={4}
      templateAreas={{
        base: `"profile" "muted" "muted" "warpcastapikey" "polls"`,
        md: `"polls profile" "polls muted" "polls muted" "polls delegations" "polls warpcastapikey"`,
      }}
    >
      <GridItem gridArea='profile'>
        <PurpleBox>
          <Box display='flex' flexDir='row' gap={2}>
            <VStack spacing={0} alignItems='start'>
              <Text fontWeight='500' display='flex' gap={2}>
                {user?.user.displayName || user?.user.username}{' '}
                <Link href={`https://warpcast.com/${user?.user.username}`} isExternal>
                  <FarcasterLogo fill='purple' />
                </Link>
              </Text>
              {degenOrEnsName && (
                <Link
                  isExternal
                  href={
                    degenOrEnsName.endsWith('.eth')
                      ? `https://rainbow.me/${degenOrEnsName}`
                      : `https://nftdegen.lol/profile/?id=${degenOrEnsName}`
                  }
                  fontSize='sm'
                  fontStyle='italic'
                >
                  {degenOrEnsName}
                </Link>
              )}
            </VStack>
          </Box>
          <ReputationCard reputation={user?.reputation} />
        </PurpleBox>
      </GridItem>
      <GridItem gridArea='muted'>{isOwnProfile && <MutedUsersList list={user?.mutedUsers} />}</GridItem>
      <GridItem gridArea='delegations'>{isOwnProfile && <Delegations delegations={user?.delegations} />}</GridItem>
      <GridItem gridArea='warpcastapikey'>{isOwnProfile && <ProfileWarpcastApiKey />}</GridItem>
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

const ProfileWarpcastApiKey = () => (
  <PurpleBox>
    <Heading fontSize='xl' fontWeight='600' color='purple.800' pos='relative'>
      Warpcast Api Key
    </Heading>
    <Text>Set your Warpcast API Key here to unlock awesome features like poll reminders.</Text>

    <WarpcastApiKey />
  </PurpleBox>
)

export default Profile
