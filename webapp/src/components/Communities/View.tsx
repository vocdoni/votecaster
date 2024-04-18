import { Avatar, Box, Flex, Grid, GridItem, Heading, Icon, Link, Text } from '@chakra-ui/react'
import { PropsWithChildren } from 'react'
import { FaExternalLinkAlt } from 'react-icons/fa'
import { Community } from '../../queries/communities'

export type CommunitiesViewProps = {
  community: Community
}

const WhiteBox = ({ children }: PropsWithChildren) => (
  <Flex alignItems='center' gap={4} padding={4} bg='white' boxShadow='sm' borderRadius='md' flexWrap='wrap' h='100%'>
    {children}
  </Flex>
)

export const CommunitiesView = ({ community }: CommunitiesViewProps) => {
  if (!community) return

  return (
    <Grid
      w='full'
      gap={4}
      gridTemplateAreas={{ base: '"profile" "links"', md: '"profile links"' }}
      gridTemplateColumns={{ base: 'full', md: '50%' }}
    >
      <GridItem gridArea='profile'>
        <WhiteBox>
          <Avatar src={community.logoURL} />
          <Box>
            <Heading size='md'>{community.name}</Heading>
            <Text fontSize='smaller' fontStyle='italic'>
              Managed by <CommunityAdmins community={community} />
            </Text>
          </Box>
        </WhiteBox>
      </GridItem>
      <GridItem gridArea='links'>
        <WhiteBox>
          {community.channels.map((channel) => (
            <Link isExternal href={channel.url}>
              {channel.name} <Icon as={FaExternalLinkAlt} w={3} />
            </Link>
          ))}
        </WhiteBox>
      </GridItem>
    </Grid>
  )
}

export const CommunityAdmins = ({ community }: CommunitiesViewProps) => {
  return community.admins.map((admin, k) => (
    <>
      <Link isExternal href={`https://warpcast.com/${admin.username}`}>
        {admin.displayName || admin.username}
      </Link>
      {k === community.admins.length - 2 ? ' & ' : k < community.admins.lengths - 2 ? ', ' : ''}
    </>
  ))
}
