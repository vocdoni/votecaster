import {
  Avatar,
  Box,
  Button,
  Flex,
  FlexProps,
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
import { Fragment, useMemo } from 'react'
import { BsChatDotsFill } from 'react-icons/bs'
import { FaPlay, FaRegCircleStop, FaSliders } from 'react-icons/fa6'
import { MdHowToVote } from 'react-icons/md'
import { SiFarcaster } from 'react-icons/si'
import { TbExternalLink } from 'react-icons/tb'
import { Link as RouterLink, useNavigate } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'
import { fetchPollsByCommunity } from '~queries/rankings'
import { getChain, getContractForChain } from '~util/chain'
import { humanDate, participation } from '~util/strings'
import { CensusTypeInfo } from './CensusTypeInfo'
import { Delegates } from './Delegates'
import { ManageCommunity } from './Manage'

type CommunitiesViewProps = {
  community?: Community
  chain: ChainKey
  refetch: (options?: RefetchOptions | undefined) => Promise<QueryObserverResult<Community, Error>>
}

const WhiteBox = ({ children, ...rest }: FlexProps) => (
  <Flex
    alignItems='start'
    gap={4}
    padding={6}
    bg='white'
    boxShadow='sm'
    borderRadius='md'
    flexWrap='wrap'
    h='100%'
    maxW='100vw'
    overflowX='auto'
    {...rest}
  >
    {children}
  </Flex>
)

export const CommunitiesView = ({ community, chain: chainAlias, refetch }: CommunitiesViewProps) => {
  const { bfetch, profile, isAuthenticated } = useAuth()
  const { onOpen: openManageModal, ...modalProps } = useDisclosure()
  const { data: communityPolls } = useQuery<Poll[], Error>({
    queryKey: ['communityPolls', chainAlias, community?.id],
    queryFn: fetchPollsByCommunity(bfetch, community as Community),
    enabled: !!community,
  })
  const chain = getChain(chainAlias)
  const navigate = useNavigate() // Hook to control navigation
  const imAdmin = useMemo(
    () => isAuthenticated && community?.admins.some((admin) => admin.fid == profile?.fid),
    [isAuthenticated, community, profile]
  )

  if (!community) return

  return (
    <Grid
      w='full'
      maxW='100%'
      gap={4}
      gridTemplateAreas={{ base: '"profile" "links" "polls"', md: '"profile links" "polls polls"' }}
      gridTemplateColumns={{ base: 'minmax(0, 1fr)', md: 'minmax(0, 1fr) minmax(0, 1fr)' }}
    >
      <GridItem gridArea='profile'>
        <WhiteBox>
          <Avatar src={community.logoURL} />
          <Box>
            <Heading size='md'>{community.name}</Heading>
            <Text fontSize='smaller' fontStyle='italic'>
              Managed by <CommunityAdmins community={community} />
            </Text>
            <Text fontSize='smaller' mt='6' display='flex' flexDir='row' gap={1} alignItems='center'>
              Deployed on{' '}
              <Link
                isExternal
                href={`${chain.blockExplorers?.default.url}/address/${getContractForChain(chainAlias)}`}
                variant='primary'
              >
                <Avatar src={chain.logo} width={4} height={4} mr={1} verticalAlign='middle' />
                {chain.name}
              </Link>
            </Text>
            {!!imAdmin && (
              <Flex mt={4} gap={4}>
                <Button leftIcon={<FaSliders />} onClick={openManageModal} variant={'outline'}>
                  Manage
                </Button>
                <ManageCommunity {...modalProps} community={community} refetch={refetch} />
                {!community.disabled && (
                  <RouterLink to={`/form/${community.id}`}>
                    <Button leftIcon={<MdHowToVote />}>Create vote</Button>
                  </RouterLink>
                )}
              </Flex>
            )}
          </Box>
        </WhiteBox>
      </GridItem>
      <GridItem gridArea='links'>
        <WhiteBox flexWrap={{ base: 'wrap', lg: 'nowrap' }}>
          <VStack alignItems='start' fontSize='sm' flexDir='column' flex={1}>
            <Heading size={'sm'} mb={2}>
              Community Info
            </Heading>
            {!!community.channels && (
              <VStack alignItems='start' spacing={0}>
                <HStack spacing={2} align='center'>
                  <Icon as={SiFarcaster} size={8} />
                  <Text fontWeight={'semibold'}>Farcaster channels</Text>
                </HStack>
                <Box ml={6}>
                  {community.channels.map((channel, index) => (
                    <Fragment key={index}>
                      <Link isExternal href={`https://warpcast.com/~/channel/${channel}`} variant='primary'>
                        /{channel}
                      </Link>
                      {index < community.channels.length && (
                        <Text as='span' mx={1} color='gray'>
                          {index < community.channels.length - 2
                            ? ', '
                            : index === community.channels.length - 2 && ' & '}
                        </Text>
                      )}
                    </Fragment>
                  ))}
                </Box>
              </VStack>
            )}
            {!!community.groupChat && (
              <HStack spacing={2} align='center'>
                <Icon as={BsChatDotsFill} />
                <Link
                  isExternal
                  href={community.groupChat}
                  variant='primary'
                  display='flex'
                  flexDir='row'
                  gap={1}
                  alignItems='center'
                >
                  Group chat
                  <Icon as={TbExternalLink} size={4} />
                </Link>
              </HStack>
            )}
            <CensusTypeInfo community={community} />
          </VStack>
          <Delegates community={community} />
        </WhiteBox>
      </GridItem>
      {!!communityPolls && (
        <GridItem gridArea='polls'>
          <WhiteBox gap={4} flexDir='column'>
            <Heading size={'md'}>Community Polls</Heading>
            <TableContainer w='full'>
              <Table>
                <Thead>
                  <Tr>
                    <Th>Question</Th>
                    <Th isNumeric>Votes</Th>
                    <Th isNumeric>Census size</Th>
                    <Th isNumeric>Participation</Th>
                    <Th>Last vote</Th>
                    <Th>Status</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {communityPolls?.map((poll, index) => (
                    <Tr key={index} role='link' onClick={() => navigate(`poll/${poll.electionId}`)} cursor='pointer'>
                      <Td>
                        {poll.question}
                        <Text as={'p'} fontSize={'xs'} color='gray'>
                          by {poll.createdByDisplayname}
                        </Text>
                      </Td>
                      <Td isNumeric>{poll.voteCount}</Td>
                      <Td isNumeric>{poll.censusParticipantsCount}</Td>
                      <Td isNumeric>{participation(poll)}</Td>
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
