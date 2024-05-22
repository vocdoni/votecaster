import {
  Alert,
  AlertDescription,
  AlertTitle,
  Box,
  Flex,
  Heading,
  Link,
  Progress,
  Skeleton,
  Text,
  VStack,
} from '@chakra-ui/react'
import { degenContractAddress } from '~constants'

export const ResultsSection = ({ poll, onChain, loading }: { poll: PollInfo; loading: boolean; onChain: boolean }) => (
  <VStack spacing={4} alignItems='left'>
    <Heading size='md'>Results</Heading>
    <Skeleton isLoaded={!loading}>
      {poll.voteCount ? (
        <Results poll={poll} onChain={onChain} />
      ) : (
        <Text>{poll.finalized ? 'This poll received no votes' : 'No votes yet'}</Text>
      )}
    </Skeleton>
  </VStack>
)

export const Results = ({ poll, onChain }: { poll: PollInfo; onChain: boolean }) => (
  <VStack px={4} alignItems='left'>
    <Heading size='sm' fontWeight={'semibold'}>
      {poll?.question}
    </Heading>
    {poll?.finalized && onChain && (
      <Alert status='success' variant='left-accent' rounded={4}>
        <Box>
          <AlertTitle fontSize={'sm'}>Results verifiable on Degenchain</AlertTitle>
          <AlertDescription fontSize={'sm'}>
            <Text>This poll has ended. The results are definitive and have been settled on the 🎩 Degenchain.</Text>
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
      <ResultsOptions poll={poll} />
    </VStack>
  </VStack>
)

export const ResultsOptions = ({ poll }: { poll: PollInfo }) => {
  if (!poll.options || !poll.voteCount) return

  return poll.options.map((option, index) => {
    const [tally] = poll.tally
    const weight = tally.reduce((acc, curr) => acc + curr, 0)

    return (
      <Box key={index} w='full'>
        <Flex justifyContent='space-between' w='full'>
          <Text>{option}</Text>
          {!!tally && <Text>{tally[index]} votes</Text>}
        </Flex>
        {!!tally && <Progress size='sm' rounded={50} value={(tally[index] / weight) * 100} />}
      </Box>
    )
  })
}
