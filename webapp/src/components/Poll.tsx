import {
  Alert,
  AlertDescription,
  AlertTitle,
  Box,
  Button,
  Flex,
  Heading,
  HStack,
  Icon,
  Image,
  Link,
  Progress,
  Skeleton,
  Tag,
  TagLabel,
  TagLeftIcon,
  Text,
  useClipboard,
  VStack,
} from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { FaCheck, FaDownload, FaPlay, FaRegCircleStop, FaRegCopy } from 'react-icons/fa6'
import { appUrl, degenContractAddress } from '~constants'
import { fetchShortURL } from '~queries/common'
import { fetchPollsRemainingVoters, fetchPollsVoters } from '~queries/polls'
import { downloadFile } from '~util/files'
import { humanDate } from '~util/strings'
import { CsvGenerator } from '../generator'
import { useAuth } from './Auth/useAuth'

export type PollViewProps = {
  onChain: boolean
  poll: PollInfo
  loading: boolean
}

export const PollView = ({ poll, loading, onChain }: PollViewProps) => {
  const { bfetch } = useAuth()
  const [electionURL, setElectionURL] = useState<string>(`${appUrl}/${poll.electionId}`)
  const { setValue, onCopy, hasCopied } = useClipboard(electionURL)

  // retrieve and set short URL (for copy-paste)
  useEffect(() => {
    if (loading || !poll || !poll.electionId) return // if ()
    const re = new RegExp(poll.electionId)
    if (!re.test(electionURL)) return
    ;(async () => {
      try {
        const url = await fetchShortURL(bfetch)(electionURL)
        setElectionURL(url)
        setValue(url)
      } catch (e) {
        console.info('error getting short url, using long version', e)
      }
    })()
  }, [loading, poll])

  return (
    <Box gap={4} display='flex' flexDir={['column', 'column', 'row']} alignItems='start'>
      <Box flex={1} bg='white' p={6} pb={12} boxShadow='md' borderRadius='md'>
        <VStack spacing={8} alignItems='left'>
          <VStack spacing={4} alignItems='left'>
            <Skeleton isLoaded={!loading}>
              <Flex gap={4}>
                {poll?.finalized ? (
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
                {poll?.finalized && onChain && (
                  <Tag colorScheme='cyan'>
                    <TagLeftIcon as={FaCheck}></TagLeftIcon>
                    <TagLabel>Results settled on-chain</TagLabel>
                  </Tag>
                )}
              </Flex>
            </Skeleton>
            <Image src={`${appUrl}/preview/${poll.electionId}`} fallback={<Skeleton height={200} />} />
            <Button
              fontSize={'sm'}
              onClick={onCopy}
              colorScheme='purple'
              alignSelf='start'
              bg={hasCopied ? 'purple.600' : 'purple.300'}
              color='white'
              size='xs'
              rightIcon={<Icon as={hasCopied ? FaCheck : FaRegCopy} />}
            >
              Copy link to the frame
            </Button>
          </VStack>
          <VStack spacing={4} alignItems='left'>
            <Heading size='md'>Results</Heading>
            <Skeleton isLoaded={!loading}>
              <VStack px={4} alignItems='left'>
                <Heading size='sm' fontWeight={'semibold'}>
                  {poll?.question}
                </Heading>
                {poll?.finalized && onChain && (
                  <Alert status='success' variant='left-accent' rounded={4}>
                    <Box>
                      <AlertTitle fontSize={'sm'}>Results verifiable on Degenchain</AlertTitle>
                      <AlertDescription fontSize={'sm'}>
                        <Text>
                          This poll has ended. The results are definitive and have been settled on the 🎩 Degenchain.
                        </Text>
                        <Link
                          fontSize={'xs'}
                          color='gray'
                          textAlign={'right'}
                          isExternal
                          href={`https://explorer.degen.tips/address/${degenContractAddress}`}
                        >
                          View contract
                        </Link>
                      </AlertDescription>
                    </Box>
                  </Alert>
                )}
                <VStack spacing={6} alignItems='left'>
                  {poll?.options.map((option, index) => {
                    const [tally] = poll!.tally
                    const weight = tally.reduce((acc, curr) => acc + curr, 0)

                    return (
                      <Box key={index} w='full'>
                        <Flex justifyContent='space-between' w='full'>
                          <Text>{option}</Text>
                          {!!poll.voteCount && !!tally && <Text>{tally[index]} votes</Text>}
                        </Flex>
                        {!!poll.voteCount && !!tally && (
                          <Progress size='sm' rounded={50} value={(tally[index] / weight) * 100} />
                        )}
                      </Box>
                    )
                  })}
                </VStack>
              </VStack>
            </Skeleton>
          </VStack>
        </VStack>
      </Box>
      <Flex flex={1} direction={'column'} gap={4}>
        <Box bg='white' p={6} boxShadow='md' borderRadius='md'>
          <Heading size='sm'>Information</Heading>
          <Skeleton isLoaded={!loading}>
            <VStack spacing={6} alignItems='left' fontSize={'sm'}>
              <Text>
                This poll {poll?.finalized ? 'has ended' : 'ends'} on {`${humanDate(poll?.endTime)}`}.{` `}
                <Link
                  variant='primary'
                  isExternal
                  href={`https://stg.explorer.vote/processes/show/#/${poll.electionId}`}
                >
                  Check the Vocdoni blockchain explorer
                </Link>
                {` `}for more information.
              </Text>
              <Text>You can download multiple lists of voters.</Text>
              <HStack spacing={2} flexWrap='wrap'>
                {!!poll.participants.length && <DownloadVotersButton electionId={poll.electionId} />}
                <DownloadRemainingVotersButton electionId={poll.electionId} />
              </HStack>
            </VStack>
          </Skeleton>
        </Box>
        <Flex gap={6}>
          <Box flex={1} bg='white' p={6} boxShadow='md' borderRadius='md'>
            <Skeleton isLoaded={!loading}>
              <ParticipantTurnout poll={poll} />
            </Skeleton>
          </Box>
          <Box flex={1} bg='white' p={6} boxShadow='md' borderRadius='md'>
            <Skeleton isLoaded={!loading}>
              <Box pb={4}>
                <Heading size='sm'>Voting Power Turnout</Heading>
                <Text fontSize={'sm'} color={'gray'}>
                  Proportion of voting power used relative to the total available.
                </Text>
              </Box>
              <Flex alignItems={'end'} gap={2}>
                <Text fontSize={'xx-large'} lineHeight={1} fontWeight={'semibold'}>
                  {poll?.turnout}
                </Text>
                <Text>%</Text>
              </Flex>
            </Skeleton>
          </Box>
        </Flex>
      </Flex>
    </Box>
  )
}

type DownloadUsersListButtonProps = {
  electionId: string
  filename: string
  text: string
  queryFn: () => Promise<string[]>
}

const DownloadUsersListButton = ({ electionId, filename, text, queryFn }: DownloadUsersListButtonProps) => {
  const {
    data: voters,
    refetch,
    isFetching,
  } = useQuery({
    queryKey: [text, electionId],
    queryFn,
    enabled: false,
  })
  const [downloaded, setDownloaded] = useState<string>('')

  useEffect(() => {
    if (voters?.length && downloaded !== JSON.stringify(voters)) {
      const csv = new CsvGenerator(
        ['Username'],
        voters.map((username) => [username]),
        filename
      )
      setDownloaded(JSON.stringify(voters))
      downloadFile(csv.url, csv.filename)
    }
  }, [voters])

  return (
    <Button
      isLoading={isFetching}
      loadingText='Preparing download...'
      onClick={() => refetch()}
      colorScheme='blue'
      size='sm'
      rightIcon={<FaDownload />}
      disabled={isFetching}
    >
      {text}
    </Button>
  )
}

const DownloadVotersButton = ({ electionId }: { electionId: string }) => {
  const { bfetch } = useAuth()

  return (
    <DownloadUsersListButton
      electionId={electionId}
      filename='voters.csv'
      text='Download voters list'
      queryFn={fetchPollsVoters(bfetch, electionId)}
    />
  )
}

const DownloadRemainingVotersButton = ({ electionId }: { electionId: string }) => {
  const { bfetch } = useAuth()

  return (
    <DownloadUsersListButton
      electionId={electionId}
      filename='remaining-voters.csv'
      text='Download remaining voters list'
      queryFn={fetchPollsRemainingVoters(bfetch, electionId)}
    />
  )
}

const participationPercentage = (poll: PollInfo) => {
  if (!poll || !poll.censusParticipantsCount) return 0

  return ((poll.voteCount / poll.censusParticipantsCount) * 100).toFixed(1)
}

const ParticipantTurnout = ({ poll }: { poll: PollInfo | null }) => {
  if (!poll) return

  const pc = poll?.censusParticipantsCount || 0
  const pp = participationPercentage(poll)

  return (
    <>
      <Box pb={4}>
        <Heading size='sm'>{pc ? `Participant Turnout` : `Participants`}</Heading>
        <Text fontSize={'sm'} color={'gray'}>
          {poll.censusParticipantsCount
            ? `Ratio of unique voters to total elegible participants.`
            : `Number of unique voters.`}
        </Text>
      </Box>
      <Flex alignItems={'end'} gap={2}>
        <Text fontSize={'xx-large'} lineHeight={1} fontWeight={'semibold'}>
          {poll?.voteCount}
        </Text>
        {!!poll?.censusParticipantsCount && <Text>/{poll?.censusParticipantsCount}</Text>}
        {!!pp && <Text fontSize='xl'>{pp}%</Text>}
      </Flex>
    </>
  )
}
