import {
  Avatar,
  Box,
  Button,
  Flex,
  Grid,
  GridItem,
  Heading,
  HStack,
  Icon,
  Link,
  Table,
  TableContainer,
  Tag,
  TagLabel,
  TagLeftIcon,
  Tbody,
  Td,
  Text,
  Th,
  Thead,
  Tr,
  useDisclosure,
  VStack,
} from '@chakra-ui/react'
import { QueryObserverResult, RefetchOptions, useQuery } from '@tanstack/react-query'
import { Fragment, PropsWithChildren, ReactElement, useMemo } from 'react'
import { BsChatDotsFill } from 'react-icons/bs'
import { FaPlay, FaRegCircleStop, FaSliders } from 'react-icons/fa6'
import { MdHowToVote } from 'react-icons/md'
import { SiFarcaster } from 'react-icons/si'
import { TbExternalLink } from 'react-icons/tb'
import { Link as RouterLink, useNavigate } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'
import { degenContractAddress } from '~constants'
import { fetchPollsByCommunity } from '~queries/tops'
import { humanDate } from '~util/strings'
import { ManageCommunity } from './Manage'

type CommunitiesViewProps = {
  community?: Community
  refetch: (options?: RefetchOptions | undefined) => Promise<QueryObserverResult<Community, Error>>
}

const WhiteBox = ({ children }: PropsWithChildren) => (
  <Flex
    alignItems='start'
    gap={4}
    padding={6}
    bg='white'
    boxShadow='sm'
    borderRadius='md'
    flexWrap='wrap'
    h='100%'
    maxW={'100vw'}
    overflowX={'auto'}
  >
    {children}
  </Flex>
)

export const CommunitiesView = ({ community, refetch }: CommunitiesViewProps) => {
  const { bfetch, profile, isAuthenticated } = useAuth()
  const { onOpen: openManageModal, ...modalProps } = useDisclosure()
  const { data: communityPolls } = useQuery<Poll[], Error>({
    queryKey: ['communityPolls', community?.id],
    queryFn: fetchPollsByCommunity(bfetch, community as Community),
    enabled: !!community,
  })
  const navigate = useNavigate() // Hook to control navigation
  const imAdmin = useMemo(
    () => isAuthenticated && community?.admins.some((admin) => admin.fid == profile?.fid),
    [isAuthenticated, community, profile]
  )
  if (!community) return

  const channelLinks: ReactElement[] = []
  community.channels.forEach((channel, index: number) => {
    channelLinks.push(
      <Link
        key={`link-${channel}`}
        fontSize='sm'
        color='gray'
        isExternal
        _hover={{ textDecoration: 'underline' }}
        href={`https://warpcast.com/~/channel/${channel}`}
      >
        /{channel}
      </Link>
    )
    // Add the separator if it's not the last item
    if (index !== community.channels.length - 1) {
      channelLinks.push(
        <Text as='span' fontSize='sm' mx={1} color={'grey'} key={`separator-${index}`}>
          &amp;
        </Text>
      )
    }
  })

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
              Deployed on{' '}
              <Link isExternal href={`https://explorer.degen.tips/address/${degenContractAddress}`}>
                <Text as={'u'}>🎩 DegenChain</Text>
              </Link>
            </Text>
            {!!imAdmin && (
              <Flex mt={4} gap={4}>
                <Button leftIcon={<FaSliders />} onClick={openManageModal} variant={'outline'}>
                  Manage
                </Button>
                <ManageCommunity {...modalProps} community={community} refetch={refetch} />
                <Button onClick={() => navigate('/')} leftIcon={<MdHowToVote />}>
                  Create vote
                </Button>
              </Flex>
            )}
          </Box>
        </WhiteBox>
      </GridItem>
      <GridItem gridArea='links'>
        <WhiteBox>
          <Box>
            <Heading size={'sm'} mb={2}>
              Community Engagement
            </Heading>
            {!!channelLinks.length && (
              <>
                <HStack spacing={2} align='center'>
                  <Icon as={SiFarcaster} size={8} />
                  <Text fontWeight={'semibold'} fontSize={'sm'}>
                    Farcaster channels
                  </Text>
                </HStack>
                <Box ml={6} mb={2}>
                  {channelLinks}
                </Box>
              </>
            )}
            {!!community.groupChat && (
              <Link isExternal href={community.groupChat}>
                <HStack spacing={2} align='center'>
                  <Icon as={BsChatDotsFill} />
                  <Heading size='xs'>
                    <Text as='u'>Group chat</Text>
                  </Heading>
                  <Icon as={TbExternalLink} size={4} />
                </HStack>
              </Link>
            )}
            {!channelLinks.length && !community.groupChat && (
              <Text>There is no aditional information for this community.</Text>
            )}
          </Box>
        </WhiteBox>
      </GridItem>
      {!!communityPolls && (
        <GridItem gridArea='polls'>
          <WhiteBox>
            <VStack width={'100%'} alignItems={'start'} gap={4}>
              <Heading size={'md'}>Community Polls</Heading>
              <TableContainer width={'100%'}>
                <Table style={{ overflowX: 'auto' }} maxW='100%'>
                  <Thead>
                    <Tr>
                      <Th>Question</Th>
                      <Th isNumeric>Votes</Th>
                      <Th isNumeric>Census size</Th>
                      <Th isNumeric>Participation(%)</Th>
                      <Th>Last vote</Th>
                      <Th>Status</Th>
                    </Tr>
                  </Thead>
                  <Tbody>
                    {communityPolls?.map((poll, index) => (
                      <Tr key={index}>
                        <Td>
                          <RouterLink to={`poll/${poll.electionId}`}>{poll.question}</RouterLink>
                          <Text as={'p'} fontSize={'xs'} color='gray'>
                            by {poll.createdByDisplayname}
                          </Text>
                        </Td>
                        <Td isNumeric>{poll.voteCount}</Td>
                        <Td isNumeric>{poll.censusParticipantsCount}</Td>
                        <Td isNumeric>{`${((poll.voteCount / poll.censusParticipantsCount) * 100).toFixed(1)}%`}</Td>
                        <Td>{poll.voteCount > 0 ? humanDate(poll.lastVoteTime) : '-'}</Td>
                        <Td>
                          <VStack>
                            {poll.finalized ? (
                              <Tag>
                                <TagLeftIcon as={FaRegCircleStop}></TagLeftIcon>
                                <TagLabel>Ended</TagLabel>
                              </Tag>
                            ) : (
                              <Tag colorScheme='green'>
                                <TagLeftIcon as={FaPlay}></TagLeftIcon>
                                <TagLabel>Ongoing</TagLabel>
                              </Tag>
                            )}
                            {poll.finalized && (
                              <Text fontSize={'xs'} color={'gray'}>
                                {humanDate(poll.endTime)}
                              </Text>
                            )}
                          </VStack>
                        </Td>
                      </Tr>
                    ))}
                  </Tbody>
                </Table>
              </TableContainer>
            </VStack>
          </WhiteBox>
        </GridItem>
      )}
    </Grid>
  )
}

export const CommunityAdmins = ({ community }: { community: Community }) => {
  if (!community) return

  return community.admins.map((admin: Profile, k: number) => (
    <Fragment key={k}>
      <Link isExternal href={`https://warpcast.com/${admin.username}`}>
        {admin.displayName || admin.username}
      </Link>
      {k === community.admins.length - 2 ? ' & ' : k < community.admins.length - 2 ? ', ' : ''}
    </Fragment>
  ))
}
