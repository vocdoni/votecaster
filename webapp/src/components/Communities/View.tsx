import { Avatar, Box, Flex, Grid, GridItem, Heading, Icon, Link, Text, HStack, Table, Thead, Tr, Td, Th, Tbody } from '@chakra-ui/react'
import { PropsWithChildren, ReactElement } from 'react'
import { TbExternalLink } from "react-icons/tb"
import { SiFarcaster } from "react-icons/si";
import { BsChatDotsFill } from "react-icons/bs";

import { useQuery } from '@tanstack/react-query'
import { Community } from '../../queries/communities'
import { fetchPollsByCommunity } from '../../queries/tops'
import { useAuth } from '../Auth/useAuth'
import { Poll } from '../../util/types';

export type CommunitiesViewProps = {
  community: Community
}

const WhiteBox = ({ children }: PropsWithChildren) => (
  <Flex alignItems='start' gap={4} padding={6} bg='white' boxShadow='sm' borderRadius='md' flexWrap='wrap' h='100%'>
    {children}
  </Flex>
)

const lastPollVote = (poll : Poll) : string => {
  if (poll.voteCount === 0) {
    return 'Never'
  }
  return (new Date(poll.lastVoteTime)).toDateString()
}

export const CommunitiesView = ({ community }: CommunitiesViewProps) => {
  const { bfetch } = useAuth()
  const { data: communityPolls } = useQuery<Poll[], Error>({
    queryKey: ['communityPolls', community?.id],
    queryFn: fetchPollsByCommunity(bfetch, community as Community),
    enabled: !!community,
  })

  if (!community) return;

  const channelLinks: ReactElement[] = [];
  community.channels.forEach((channel, index) => {
    channelLinks.push(
      <Link key={`link-${channel}`} fontSize="sm" color="gray" isExternal _hover={{ textDecoration: 'underline' }} href={`https://warpcast.com/~/channel/${channel}`}>
        /{channel}
      </Link>
    );
    // Add the separator if it's not the last item
    if (index !== community.channels.length - 1) {
      channelLinks.push(<Text as="span" fontSize="sm" mx={1} color={'grey'} key={`separator-${index}`}>&amp;</Text>);
    }
  });

  return (
    <Grid
      w='full'
      gap={4}
      gridTemplateAreas={{ base: '"profile" "links" "polls"', md: '"profile links" "polls polls"' }}
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
            <Text fontSize='smaller' mt='6'>
              Deployed on <Link isExternal href={`https://explorer.degen.tips/address/${import.meta.env.VOCDONI_COMMUNITYHUBADDRESS}`}><Text as={'u'}>🎩 DegenChain</Text></Link>
            </Text>
          </Box>
        </WhiteBox>
      </GridItem>
      <GridItem gridArea='links'>
        <WhiteBox>
          <Box>
            <Heading size={'sm'} mb={2}>Community Engagement</Heading>
            <HStack spacing={2} align='center'>
              <Icon as={SiFarcaster} size={8}/>
              <Text fontWeight={'semibold'} fontSize={'sm'}>Official Farcaster channels</Text>
            </HStack>
            <Box ml={6} mb={2}>
              { channelLinks }
            </Box>
            <Link isExternal href={community.groupChat}>
              <HStack spacing={2} align='center'>
                <Icon as={BsChatDotsFill}/> 
                <Heading size='xs'><Text as='u'>Official group chat</Text></Heading>
                <Icon as={TbExternalLink} size={4} />
              </HStack>
            </Link>
          </Box>
        </WhiteBox>
      </GridItem>
      { !!communityPolls && <GridItem gridArea='polls'>
        <WhiteBox>
          <Heading size={'md'} mb={4}>Community Polls</Heading>
          <Table>
            <Thead>
              <Tr>
                <Th>Question</Th>
                <Th isNumeric>Census size</Th>
                <Th isNumeric>Votes</Th>
                <Th isNumeric>Turnout(%)</Th>
                <Th>Last vote</Th>
              </Tr>
            </Thead>
            <Tbody>
              {communityPolls?.map((poll, index) => (
                <Tr key={index}>
                  <Td>{poll.title} <small>by {poll.createdByDisplayname}</small></Td>
                  <Td isNumeric>{poll.censusParticipantsCount}</Td>
                  <Td isNumeric>{poll.voteCount}</Td>
                  <Td isNumeric>{`${poll.turnout}%`}</Td>
                  <Td textAlign='center'>{ lastPollVote(poll) }</Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        </WhiteBox>
      </GridItem>}
    </Grid>
  )
}

export const CommunityAdmins = ({ community }: CommunitiesViewProps) => {
  return community.admins.map((admin, k) => (
    <>
      <Link isExternal href={`https://warpcast.com/${admin.username}`}>
        {admin.displayName || admin.username}
      </Link>
      {k === community.admins.length - 2 ? ' & ' : k < community.admins.length - 2 ? ', ' : ''}
    </>
  ))
}
