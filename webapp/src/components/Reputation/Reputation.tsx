import {
  Alert,
  Box,
  CircularProgress,
  CircularProgressLabel,
  CircularProgressProps,
  Flex,
  Icon,
  IconButton,
  Link,
  Popover,
  PopoverArrow,
  PopoverBody,
  PopoverCloseButton,
  PopoverContent,
  PopoverTrigger,
  Progress,
  SimpleGrid,
  Stat,
  StatLabel,
  StatNumber,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useBreakpointValue,
  useColorModeValue,
} from '@chakra-ui/react'
import { PropsWithChildren, useMemo } from 'react'
import { FaHeart, FaInfo, FaRegFaceGrinStars, FaUserGroup } from 'react-icons/fa6'
import { ImStatsDots } from 'react-icons/im'
import { MdOutlineHowToVote } from 'react-icons/md'
import { SlPencil } from 'react-icons/sl'
import { Reputation } from '~components/Reputation/ReputationContext'
import { useReputation } from '~queries/profile'

export const ReputationProgress = ({ reputation, ...props }: CircularProgressProps & { reputation?: Reputation }) => {
  if (!reputation) return

  return (
    <CircularProgress value={reputation.totalReputation} max={100} color='purple.600' thickness='12px' {...props}>
      <CircularProgressLabel>{reputation.totalReputation}%</CircularProgressLabel>
    </CircularProgress>
  )
}

export const ReputationCard = ({ reputation }: { reputation?: Reputation }) => {
  const bg = useColorModeValue('whiteAlpha.500', 'whiteAlpha.200')
  const boxShadow = useColorModeValue('0px 4px 6px rgba(0, 0, 0, 0.1)', '0px 4px 6px rgba(0, 0, 0, 0.3)')
  const isMobile = useBreakpointValue({ base: true, md: false })

  if (!reputation) return null

  return (
    <Popover placement='auto' trigger='hover' closeOnBlur>
      <PopoverTrigger>
        <Box
          p={4}
          bg={bg}
          boxShadow={boxShadow}
          borderRadius='lg'
          bgGradient='linear(to-r, purple.700, purple.400)'
          color='white'
          pos='relative'
        >
          <Flex justifyContent='space-around' alignItems='center'>
            <Stat>
              <StatLabel fontSize='lg'>
                <Icon as={ImStatsDots} boxSize={4} /> Votecaster reputation
              </StatLabel>
              <StatNumber fontSize='2xl'>{reputation.totalReputation}</StatNumber>
            </Stat>
            <ReputationProgress reputation={reputation} />
          </Flex>
          {isMobile && (
            <PopoverTrigger>
              <IconButton
                aria-label='Open info'
                icon={<Icon as={FaInfo} />}
                variant='text'
                color='white'
                pos='absolute'
                top={0}
                right={0}
              />
            </PopoverTrigger>
          )}
          <SimpleGrid columns={2} spacing={3} mt={4}>
            <Stat>
              <StatLabel fontSize='x-small'>
                Manager of {reputation.activityCounts.communitiesCount} communities
              </StatLabel>
              <FlexStatNumber>
                {reputation.activityPoints.communitiesPoints}/{reputation.activityInfo.maxCommunityReputation}
                {` `}
                <Icon as={FaUserGroup} boxSize={3} />
              </FlexStatNumber>
            </Stat>{' '}
            <Stat>
              <StatLabel fontSize='x-small'>Casted {reputation.activityCounts.castVotesCount} votes</StatLabel>
              <FlexStatNumber>
                {reputation.activityPoints.castVotesPoints}/{reputation.activityInfo.maxVotesReputation}
                {` `}
                <Icon as={MdOutlineHowToVote} boxSize={3.5} />
              </FlexStatNumber>
            </Stat>
            <Stat>
              <StatLabel fontSize='x-small'>Created {reputation.activityCounts.createdElectionsCount} polls</StatLabel>
              <FlexStatNumber>
                {reputation.activityPoints.createdElectionsPoints}/{reputation.activityInfo.maxElectionsReputation}
                {` `}
                <Icon as={SlPencil} boxSize={3} />
              </FlexStatNumber>
            </Stat>
            <Stat>
              <StatLabel fontSize='x-small'>{reputation.activityCounts.followersCount} followers</StatLabel>
              <FlexStatNumber>
                {reputation.activityPoints.followersPoints}/{reputation.activityInfo.maxFollowersReputation}
                {` `}
                <Icon as={FaHeart} boxSize={3} />
              </FlexStatNumber>
            </Stat>
            <Stat>
              <StatLabel fontSize='x-small'>
                Participated in {reputation.activityCounts.participationsCount} polls
              </StatLabel>
              <FlexStatNumber>
                {reputation.activityPoints.participationsPoints}/{reputation.activityInfo.maxCastedReputation}
                {` `}
                <Icon as={FaRegFaceGrinStars} boxSize={3} />
              </FlexStatNumber>
            </Stat>
          </SimpleGrid>
        </Box>
      </PopoverTrigger>
      <PopoverContent bg='purple.500' border='none'>
        <PopoverArrow bg='purple.500' />
        <PopoverCloseButton color='white' />
        <PopoverBody color='white' p={5}>
          To enhance your reputation and unlock premium features, actively creating polls and encouraging participation
          is highly effective. Additionally, you can accumulate points by engaging with polls created by others.
        </PopoverBody>
      </PopoverContent>
    </Popover>
  )
}
export const ReputationTable = ({ username }: { username?: string }) => {
  const { data: reputation, error, isFetching } = useReputation(username)

  const totalReputation = reputation?.totalReputation
  const totalAvailable = useMemo(() => {
    if (!reputation) return 0

    return Object.values(reputation.activityInfo).reduce((acc, curr) => acc + curr, 0)
  }, [reputation])

  if (isFetching) {
    return <Progress size='xs' isIndeterminate colorScheme='purple' />
  }

  if (error) {
    return <Alert status='error'>{error.message}</Alert>
  }

  if (!reputation) {
    return null
  }

  return (
    <Table>
      <Thead>
        <Tr>
          <Th>Action</Th>
          <Th>Earned reputation</Th>
          <Th>Available reputation</Th>
        </Tr>
      </Thead>
      <Tbody>
        <Tr>
          <Td>Created {reputation.activityCounts.communitiesCount} communities</Td>
          <Td>{reputation.activityPoints.communitiesPoints}</Td>
          <Td>{reputation.activityInfo.maxCommunityReputation}</Td>
        </Tr>
        <Tr>
          <Td>Created {reputation.activityCounts.createdElectionsCount} polls</Td>
          <Td>{reputation.activityPoints.createdElectionsPoints}</Td>
          <Td>{reputation.activityInfo.maxElectionsReputation}</Td>
        </Tr>
        <Tr>
          <Td>Participated in {reputation.activityCounts.participationsCount} polls</Td>
          <Td>{reputation.activityPoints.participationsPoints}</Td>
          <Td>{reputation.activityInfo.maxCastedReputation}</Td>
        </Tr>
        <Tr>
          <Td>Cast {reputation.activityCounts.castVotesCount} votes</Td>
          <Td>{reputation.activityPoints.castVotesPoints}</Td>
          <Td>{reputation.activityInfo.maxVotesReputation}</Td>
        </Tr>
        <Tr>
          <Td>{reputation.activityCounts.followersCount} followers</Td>
          <Td>{reputation.activityPoints.followersPoints}</Td>
          <Td>{reputation.activityInfo.maxFollowersReputation}</Td>
        </Tr>
        <Tr>
          <Td>
            <WarpcastLink path='vocdoni'>Follow @vocdoni</WarpcastLink> on Farcaster
          </Td>
          <Td>
            {reputation.boosters.isVocdoniFarcasterFollower
              ? reputation.boostersInfo.vocdoniFarcasterFollowerPuntuaction
              : 0}
          </Td>
          <Td>{reputation.boostersInfo.vocdoniFarcasterFollowerPuntuaction}</Td>
        </Tr>
        <Tr>
          <Td>
            <WarpcastLink path='votecaster'>Follow @votecaster</WarpcastLink> on Farcaster
          </Td>
          <Td>
            {reputation.boosters.isVotecasterFarcasterFollower
              ? reputation.boostersInfo.votecasterFarcasterFollowerPuntuaction
              : 0}
          </Td>
          <Td>{reputation.boostersInfo.votecasterFarcasterFollowerPuntuaction}</Td>
        </Tr>
        <Tr>
          <Td>Own a votecaster NFT pass</Td>
          <Td>{reputation.boosters.hasVotecasterNFTPass ? reputation.boostersInfo.votecasterNFTPassPuntuaction : 0}</Td>
          <Td>{reputation.boostersInfo.votecasterNFTPassPuntuaction}</Td>
        </Tr>
        <Tr>
          <Td>Own the votecaster launch NFT</Td>
          <Td>
            {reputation.boosters.hasVotecasterLaunchNFT ? reputation.boostersInfo.votecasterLaunchNFTPuntuaction : 0}
          </Td>
          <Td>{reputation.boostersInfo.votecasterLaunchNFTPuntuaction}</Td>
        </Tr>
        <Tr>
          <Td>
            Recasted the <WarpcastLink path='vocdoni/0x7eafebeb'>votecaster launch cast</WarpcastLink>
          </Td>
          <Td>
            {reputation.boosters.votecasterAnnouncementRecasted
              ? reputation.boostersInfo.votecasterAnnouncementRecastedPuntuaction
              : 0}
          </Td>
          <Td>{reputation.boostersInfo.votecasterAnnouncementRecastedPuntuaction}</Td>
        </Tr>
        <Tr>
          <Td>Hold Farcaster OG NFT</Td>
          <Td>{reputation.boosters.hasFarcasterOGNFT ? reputation.boostersInfo.farcasterOGNFTPuntuaction : 0}</Td>
          <Td>{reputation.boostersInfo.farcasterOGNFTPuntuaction}</Td>
        </Tr>
        <Tr>
          <Td>Hold +10k $degen</Td>
          <Td>{reputation.boosters.has10kDegenAtLeast ? reputation.boostersInfo.degenAtLeast10kPuntuaction : 0}</Td>
          <Td>{reputation.boostersInfo.degenAtLeast10kPuntuaction}</Td>
        </Tr>
        <Tr>
          <Td>Hold DegenDAO NFT</Td>
          <Td>{reputation.boosters.hasDegenDAONFT ? reputation.boostersInfo.degenDAONFTPuntuaction : 0}</Td>
          <Td>{reputation.boostersInfo.degenDAONFTPuntuaction}</Td>
        </Tr>
        <Tr>
          <Td>
            Hold <OpenSeaCollection slug='degen-haberdashers'>Haberdashery NFT</OpenSeaCollection>
          </Td>
          <Td>{reputation.boosters.hasHaberdasheryNFT ? reputation.boostersInfo.haberdasheryNFTPuntuaction : 0}</Td>
          <Td>{reputation.boostersInfo.haberdasheryNFTPuntuaction}</Td>
        </Tr>
        <Tr>
          <Td>
            Hold{' '}
            <Link href='https://news.kiwistand.com/' isExternal>
              🥝Kiwi NFT
            </Link>
          </Td>
          <Td>{reputation.boosters.hasKIWI ? reputation.boostersInfo.kiwiPuntuaction : 0}</Td>
          <Td>{reputation.boostersInfo.kiwiPuntuaction}</Td>
        </Tr>
        <Tr>
          <Td>
            Hold <WarpcastLink path='betashop.eth/0xd375d45d'>Ⓜ️oxie Pass</WarpcastLink>
          </Td>
          <Td>{reputation.boosters.hasMoxiePass ? reputation.boostersInfo.moxiePassPuntuaction : 0}</Td>
          <Td>{reputation.boostersInfo.moxiePassPuntuaction}</Td>
        </Tr>
        <Tr>
          <Td>Hold .degen NFT</Td>
          <Td>{reputation.boosters.hasNameDegen ? reputation.boostersInfo.nameDegenPuntuaction : 0}</Td>
          <Td>{reputation.boostersInfo.nameDegenPuntuaction}</Td>
        </Tr>
        <Tr>
          <Td>Hold ProxyStudio NFT</Td>
          <Td>{reputation.boosters.hasProxyStudioNFT ? reputation.boostersInfo.proxyStudioNFTPuntuaction : 0}</Td>
          <Td>{reputation.boostersInfo.proxyStudioNFTPuntuaction}</Td>
        </Tr>
        <Tr>
          <Td>Hold +5 $PROXY</Td>
          <Td>{reputation.boosters.has5ProxyAtLeast ? reputation.boostersInfo.proxyAtLeast5Puntuaction : 0}</Td>
          <Td>{reputation.boostersInfo.proxyAtLeast5Puntuaction}</Td>
        </Tr>
        <Tr>
          <Td>
            Hold <OpenSeaCollection slug='tokyo-dao-1'>TokyoDAO NFT</OpenSeaCollection>
          </Td>
          <Td>{reputation.boosters.hasTokyoDAONFT ? reputation.boostersInfo.tokyoDAONFTPuntuaction : 0}</Td>
          <Td>{reputation.boostersInfo.tokyoDAONFTPuntuaction}</Td>
        </Tr>
        <Tr fontWeight='bold'>
          <Td>Totals</Td>
          <Td>{totalReputation}</Td>
          <Td>{totalAvailable}</Td>
        </Tr>
      </Tbody>
    </Table>
  )
}

const OpenSeaCollection = ({ slug, children }: PropsWithChildren<{ slug: string }>) => (
  <Link href={`https://opensea.io/collection/${slug}`} isExternal>
    {children}
  </Link>
)

const WarpcastLink = ({ path, children }: PropsWithChildren<{ path: string }>) => (
  <Link href={`https://warpcast.com/${path}`} isExternal>
    {children}
  </Link>
)

const FlexStatNumber = ({ children }: PropsWithChildren) => (
  <StatNumber fontSize='sm' display='flex' flexDir='row' alignItems='center' gap={1}>
    {children}
  </StatNumber>
)
