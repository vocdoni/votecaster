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
import { getChain, getContractForChain } from '~util/chain'
import { chainFromId } from '~util/mappings'

export const ResultsSection = ({ poll, loading }: { poll: PollInfo; loading: boolean }) => {
  return (
    <VStack spacing={4} alignItems='left'>
      <Heading size='md'>Results</Heading>
      <Skeleton isLoaded={!loading}>
        <VStack spacing={2} alignItems='left'>
          {poll.voteCount ? (
            <Results poll={poll} />
          ) : (
            <Text>{poll.finalized ? 'This poll received no votes' : 'No votes yet'}</Text>
          )}
          {poll?.finalized &&
            poll.community &&
            (() => {
              const alias = chainFromId(poll.community?.id)
              const chain = getChain(alias)
              return (
                <Alert status='success' variant='left-accent' rounded={4}>
                  <Box>
                    <AlertTitle fontSize={'sm'}>Results verifiable on {chain.name}</AlertTitle>
                    <AlertDescription fontSize={'sm'}>
                      <Text>
                        This poll has ended. The results are definitive and have been settled on the {chain.name}{' '}
                        blockchain.
                      </Text>
                      <Link
                        fontSize={'xs'}
                        color='gray'
                        textAlign={'right'}
                        isExternal
                        href={`${chain.blockExplorers?.default.url}/address/${getContractForChain(alias)}`}
                      >
                        View contract
                      </Link>
                    </AlertDescription>
                  </Box>
                </Alert>
              )
            })()}
        </VStack>
      </Skeleton>
    </VStack>
  )
}

export const Results = ({ poll }: { poll: PollInfo }) => {
  return (
    <VStack px={4} alignItems='left'>
      <Heading size='sm' fontWeight={'semibold'}>
        {poll?.question}
      </Heading>

      <VStack spacing={6} alignItems='left'>
        <ResultsOptions poll={poll} />
      </VStack>
    </VStack>
  )
}

export const ResultsOptions = ({ poll }: { poll: PollInfo }) => {
  if (!poll.options || !poll.voteCount) return

  return poll.options.map((option, index) => {
    const [tally] = poll.tally
    const weight = tally.reduce((acc, curr) => acc + curr, 0)
    const percentage = (tally[index] / weight) * 100

    return (
      <Box key={index} w='full'>
        <Flex justifyContent='space-between' w='full'>
          <Text>
            {option} {!!tally && <>({percentage.toFixed(2)} %)</>}
          </Text>
          {!!tally && <Text>{tally[index]} voting power</Text>}
        </Flex>
        {!!tally && <Progress size='sm' rounded={50} value={percentage} />}
      </Box>
    )
  })
}
