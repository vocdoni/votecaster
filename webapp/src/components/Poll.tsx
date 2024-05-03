import {
  Alert,
  AlertDescription,
  AlertTitle,
  Box,
  Button,
  Flex,
  Heading,
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
import { useEffect, useState } from 'react'
import { FaCheck, FaPlay, FaRegCircleStop, FaRegCopy } from 'react-icons/fa6'
import { appUrl, degenContractAddress } from '~constants'
import { fetchShortURL } from '~queries/common'
import { useAuth } from './Auth/useAuth'
import { CensusListModal } from './Poll/CensusListModal'
import { Information } from './Poll/Information'
import { ParticipantTurnout, VotingPower } from './Poll/Turnout'

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
            <Information poll={poll} />
          </Skeleton>
        </Box>
        <Flex gap={6}>
          <Box flex={1} bg='white' p={6} boxShadow='md' borderRadius='md'>
            <Skeleton isLoaded={!loading} height='100%'>
              <CensusListModal id={poll.electionId}>
                <ParticipantTurnout mb='auto' poll={poll} />
              </CensusListModal>
            </Skeleton>
          </Box>
          <Box flex={1} bg='white' p={6} boxShadow='md' borderRadius='md'>
            <Skeleton isLoaded={!loading}>
              <VotingPower poll={poll} />
            </Skeleton>
          </Box>
        </Flex>
      </Flex>
    </Box>
  )
}
