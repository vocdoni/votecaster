import {
  Alert,
  AlertDescription,
  AlertTitle,
  Box,
  Button,
  Flex,
  Heading,
  Image,
  Link,
  Progress,
  Skeleton,
  Tag,
  TagLabel,
  TagLeftIcon,
  Text,
  VStack,
} from '@chakra-ui/react'
import { useEffect, useMemo, useState } from 'react'
import { FaCheck, FaDownload, FaPlay, FaRegCircleStop } from 'react-icons/fa6'
import { appUrl, degenContractAddress } from '~constants'
import { fetchShortURL } from '~queries/common'
import { humanDate } from '~util/strings'
import { CsvGenerator } from '../generator'
import { useAuth } from './Auth/useAuth'

export type PollViewProps = {
  electionId: string | undefined
  onChain: boolean
  poll: PollResult | null
  loading: boolean
  voters: string[]
  errorMessage: string | null
}

export const PollView = ({ poll, voters, electionId, loading, errorMessage, onChain }: PollViewProps) => {
  const { bfetch } = useAuth()
  const [electionURL, setElectionURL] = useState<string>(`${appUrl}/${electionId}`)

  useEffect(() => {
    if (loading || !poll || !electionId) return
    ;(async () => {
      // get the short url
      try {
        const url = await fetchShortURL(bfetch)(electionURL)
        setElectionURL(url)
      } catch (e) {
        console.info('error getting short url, using long version', e)
      }
    })()
  }, [])

  const usersfile = useMemo(() => {
    if (!voters.length) return { url: '', filename: '' }
    return new CsvGenerator(
      ['Username'],
      voters.map((username) => [username])
    )
  }, [voters])

  const copyToClipboard = (input: string) => {
    if (navigator && navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(input).catch(console.error)
    } else console.error('clipboard API not available')
  }

  const participationPercentage = useMemo(() => {
    if (!poll || !poll.censusParticipantsCount) return 0

    return ((poll.voteCount / poll.censusParticipantsCount) * 100).toFixed(1)
  }, [poll])

  if (errorMessage) return <Text>{errorMessage}</Text>

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
            <Image src={`${appUrl}/preview/${electionId}`} fallback={<Skeleton height={200} />} />
            <Link fontSize={'sm'} color={'gray'} onClick={() => copyToClipboard(electionURL)}>
              Copy link to the frame
            </Link>
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
                  {poll?.options.map((option, index) => (
                    <Box key={index} w='full'>
                      <Flex justifyContent='space-between' w='full'>
                        <Text>{option}</Text>
                        {!!poll.voteCount && !!poll?.tally[0] && <Text>{poll?.tally[0][index]} votes</Text>}
                      </Flex>
                      {!!poll.voteCount && !!poll?.tally[0] && (
                        <Progress size='sm' rounded={50} value={(poll?.tally[0][index] / poll?.voteCount) * 100} />
                      )}
                    </Box>
                  ))}
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
                <Link variant='primary' isExternal href={`https://stg.explorer.vote/processes/show/#/${electionId}`}>
                  Check the Vocdoni blockchain explorer
                </Link>
                {` `}for more information.
              </Text>
              {!!voters.length && (
                <>
                  <Text>You can download the list of users who casted their votes.</Text>
                  <Link href={usersfile.url} download={'voters-list.csv'}>
                    <Button colorScheme='blue' size='sm' rightIcon={<FaDownload />}>
                      Download voters
                    </Button>
                  </Link>
                </>
              )}
            </VStack>
          </Skeleton>
        </Box>
        <Flex gap={6}>
          <Box flex={1} bg='white' p={6} boxShadow='md' borderRadius='md'>
            <Skeleton isLoaded={!loading}>
              <Box pb={4}>
                <Heading size='sm'>Participant Turnout</Heading>
                <Text fontSize={'sm'} color={'gray'}>
                  Ratio of unique voters to total elegible participants.
                </Text>
              </Box>
              <Flex alignItems={'end'} gap={2}>
                <Text fontSize={'xx-large'} lineHeight={1} fontWeight={'semibold'}>
                  {poll?.voteCount}
                </Text>
                <Text>/{poll?.censusParticipantsCount}</Text>
                {!!participationPercentage && <Text fontSize='xl'>{participationPercentage}%</Text>}
              </Flex>
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
