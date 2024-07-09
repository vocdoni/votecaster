import {
  Box,
  Button,
  Flex,
  Heading,
  Icon,
  Image,
  Link,
  Skeleton,
  Tag,
  TagLabel,
  TagLeftIcon,
  useClipboard,
  VStack,
} from '@chakra-ui/react'
import { useEffect, useState } from 'react'
import { FaCheck, FaPlay, FaRegCircleStop, FaRegCopy } from 'react-icons/fa6'
import { IoMdArrowBack } from 'react-icons/io'
import { Link as RouterLink } from 'react-router-dom'
import { appUrl } from '~constants'
import { fetchShortURL } from '~queries/polls'
import { chainFromId } from '~util/mappings'
import { useAuth } from './Auth/useAuth'
import { ChainStatus } from './Healthcheck/ChainStatus'
import { Information } from './Poll/Information'
import { ResultsSection } from './Poll/Results'
import { ParticipantTurnout, VotingPower } from './Poll/Turnout'

export type PollViewProps = {
  poll: PollInfo
  loading: boolean
}

export const PollView = ({ poll, loading }: PollViewProps) => {
  const { bfetch } = useAuth()
  const [electionURL, setElectionURL] = useState<string>(`${appUrl}/${poll.electionId}`)
  const { setValue, onCopy, hasCopied } = useClipboard(electionURL)

  // retrieve and set short URL (for copy-paste)
  useEffect(() => {
    if (loading || !poll || !poll.electionId) return

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
            <Skeleton isLoaded={!loading} display='flex' justifyContent='end'>
              {poll.community && poll.community.id ? (
                <Back link={`/communities/${poll.community.id.replace(':', '/')}`} text={poll.community.name}></Back>
              ) : (
                <Back link={`/profile/${poll.createdByUsername}`} text={poll.createdByDisplayname}></Back>
              )}
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
                {poll?.finalized && poll.community && (
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
              Copy frame link
            </Button>
          </VStack>
          <ResultsSection poll={poll} loading={loading} />
        </VStack>
      </Box>
      <Flex flex={1} direction={'column'} gap={4}>
        <Box bg='white' p={6} boxShadow='md' borderRadius='md'>
          <Heading size='sm'>Information</Heading>
          <Skeleton isLoaded={!loading}>
            <Information poll={poll} url={electionURL} />
          </Skeleton>
        </Box>
        <Flex gap={6}>
          <Box flex={1} bg='white' p={6} boxShadow='md' borderRadius='md'>
            <Skeleton isLoaded={!loading} height='100%'>
              <ParticipantTurnout mb='auto' poll={poll} />
            </Skeleton>
          </Box>
          {!!poll.turnout && (
            <Box flex={1} bg='white' p={6} boxShadow='md' borderRadius='md'>
              <Skeleton isLoaded={!loading}>
                <VotingPower poll={poll} />
              </Skeleton>
            </Box>
          )}
        </Flex>
        {poll.community && <ChainStatus alias={chainFromId(poll.community.id)} />}
      </Flex>
    </Box>
  )
}

const Back = ({ link, text }: { link: string; text: string }) => (
  <Link
    as={RouterLink}
    colorScheme='purple'
    alignSelf='start'
    size='xs'
    display='flex'
    alignItems='center'
    gap={1}
    to={link}
    mr='auto'
  >
    <Icon as={IoMdArrowBack} /> {text}
  </Link>
)
