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
  Skeleton,
  Progress,
  VStack,
  Tag,
  TagLeftIcon,
  TagLabel,
  Text,
} from '@chakra-ui/react'
import { useEffect, useMemo, useState } from 'react'
import { FaDownload, FaCheck, FaRegCircleStop, FaPlay } from 'react-icons/fa6'

import { useAuth } from './Auth/useAuth'
import { fetchShortURL } from '../queries/common'
import type { PollResult } from '../util/types'
import { humanDate } from '../util/strings'
import { CsvGenerator } from '../generator'
import { appUrl, degenContractAddress } from '../util/constants'


export type PollViewProps = {
  electionId: string | undefined,
  poll: PollResult | null,
  loading: boolean | false,
  loaded: boolean | false,
  errorMessage: string | null
}

export const PollView = ({poll, electionId, loading, loaded, errorMessage}: PollViewProps) => {
  const { bfetch } = useAuth()
  const [voters, setVoters] = useState([])
  const [electionURL, setElectionURL] = useState<string>(`${appUrl}/${electionId}`)

  useEffect(() => {
    if (loaded || loading || !poll || !electionId ) return
      ; (async () => {
        // get the short url
        try {
          const url = await fetchShortURL(bfetch)(electionURL)
          setElectionURL(url)
        } catch (e) {
          console.log("error getting short url, using default", e)
        }
        // get the voters if there are any
        if (poll.voteCount > 0) {
          try {
            const response = await fetch(`${import.meta.env.APP_URL}/votersOf/${electionId}`)
            const data = await response.json()
            setVoters(data.voters)
          } catch (e) {
            console.error("error geting election voters", e)
          }
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
      navigator.clipboard.writeText(input).catch(console.error);
    } else console.error('clipboard API not available');
  };

  const participationPercentage = useMemo(() => {
    if (!poll) return 0
    return (poll.voteCount / poll.censusParticipantsCount * 100).toFixed(1)
  }, [poll])

  if (errorMessage) return <Text>{errorMessage}</Text>

  return (
    <Box
      gap={4}
      display='flex'
      flexDir={['column', 'column', 'row']}
      alignItems='start'>
      <Box flex={1} bg='white' p={6} pb={12} boxShadow='md' borderRadius='md'>
        <VStack spacing={8} alignItems='left'>
          <VStack spacing={4} alignItems='left'>
            <Skeleton isLoaded={!loading}>
              <Flex gap={4}>
                {poll?.finalized ?
                  <Tag>
                    <TagLeftIcon as={FaRegCircleStop}></TagLeftIcon>
                    <TagLabel>Ended</TagLabel>
                  </Tag> :
                  <Tag colorScheme='green'>
                    <TagLeftIcon as={FaPlay}></TagLeftIcon>
                    <TagLabel>Ongoing</TagLabel>
                  </Tag>
                }
                {poll?.finalized && <Tag colorScheme='cyan'>
                  <TagLeftIcon as={FaCheck}></TagLeftIcon>
                  <TagLabel>Results settled on-chain</TagLabel>
                </Tag>}
              </Flex>
            </Skeleton>
            <Image src={`${import.meta.env.APP_URL}/preview/${electionId}`} fallback={<Skeleton height={200} />} />
            <Link fontSize={'sm'} color={'gray'} onClick={() => copyToClipboard(electionURL)}>Copy link to the frame</Link>
          </VStack>
          <VStack spacing={4} alignItems='left'>
            <Heading size='md'>Results</Heading>
            <Skeleton isLoaded={!loading}>
              <VStack px={4} alignItems='left'>
                <Heading size='sm' fontWeight={'semibold'}>{poll?.question}</Heading>
                {poll?.finalized && <Alert status='success' variant='left-accent' rounded={4}>
                  <Box>
                    <AlertTitle fontSize={'sm'}>Results verifiable on Degenchain</AlertTitle>
                    <AlertDescription fontSize={'sm'}>
                      <Text>This poll has ended. The results are definitive and have been settled on the 🎩 Degenchain.</Text>
                      <Link fontSize={'xs'} color='gray' textAlign={'right'} isExternal href={`https://explorer.degen.tips/address/${degenContractAddress}`}>View contract</Link>
                    </AlertDescription>
                  </Box>
                </Alert>}
                <VStack spacing={6} alignItems='left'>
                  {poll?.options.map((option, index) => (
                    <Box key={index} w='full'>
                      <Flex justifyContent='space-between' w='full'>
                        <Text>{option}</Text>
                        {!!poll?.tally[0] && <Text>{poll?.tally[0][index]} votes</Text>}
                      </Flex>
                      {!!poll?.tally[0] && <Progress size='sm' rounded={50} value={poll?.tally[0][index] / poll?.voteCount * 100} />}
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
                This poll {poll?.finalized ? 'has ended' : 'ends'} on {`${humanDate(poll?.endTime)}`}. Check the Vocdoni blockchain explorer for <Link textDecoration={'underline'} isExternal href={`https://stg.explorer.vote/processes/show/#/${electionId}`}>more information</Link>.
              </Text>
              {voters.length > 0 && <>
                <Text>You can download the list of users who casted their votes.</Text>
                <Link href={usersfile.url} download={'voters-list.csv'}>
                  <Button colorScheme='blue' size='sm' rightIcon={<FaDownload />}>Download voters</Button>
                </Link>
              </>}
            </VStack>
          </Skeleton>
        </Box>
        <Flex gap={6}>
          <Box flex={1} bg='white' p={6} boxShadow='md' borderRadius='md'>
            <Skeleton isLoaded={!loading}>
              <Box pb={4}>
                <Heading size='sm'>Participant Turnout</Heading>
                <Text fontSize={'sm'} color={'gray'}>Members who voted vs. total members in the census.</Text>
              </Box>
              <Flex alignItems={'end'} gap={2}>
                <Text fontSize={'xx-large'} lineHeight={1} fontWeight={'semibold'}>{poll?.voteCount}</Text>
                <Text>/{poll?.censusParticipantsCount}</Text>
                <Text fontSize={'xl'}>{participationPercentage}%</Text>
              </Flex>
            </Skeleton>
          </Box>
          <Box flex={1} bg='white' p={6} boxShadow='md' borderRadius='md'>
            <Skeleton isLoaded={!loading}>
              <Box pb={4}>
                <Heading size='sm'>Voting Power Turnout</Heading>
                <Text fontSize={'sm'} color={'gray'}>Total votes cast versus total vote weight.</Text>
              </Box>
              <Flex alignItems={'end'} gap={2}>
                <Text fontSize={'xx-large'} lineHeight={1} fontWeight={'semibold'}>{poll?.turnout}</Text>
                <Text>%</Text>
              </Flex>
            </Skeleton>
          </Box>
        </Flex>
      </Flex>
    </Box>
  )
}
