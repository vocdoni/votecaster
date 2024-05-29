import { Box, BoxProps, Flex, Heading, Text } from '@chakra-ui/react'

export const VotingPower = ({ poll }: { poll?: PollInfo }) => {
  if (!poll) return

  return (
    <>
      <Box pb={4}>
        <Heading size='sm'>Voting Power Turnout</Heading>
        <Text fontSize={'sm'} color={'gray'}>
          Proportion of voting power used relative to the total available.
        </Text>
      </Box>
      <Flex alignItems={'end'} gap={2}>
        <Text fontSize={'xx-large'} lineHeight={1} fontWeight={'semibold'}>
          {Math.round(poll?.turnout * 100) / 100}
        </Text>
        <Text>%</Text>
      </Flex>
    </>
  )
}

export const ParticipantTurnout = ({ poll, ...props }: { poll?: PollInfo } & BoxProps) => {
  if (!poll) return

  const pc = poll?.censusParticipantsCount || 0
  const pp = participationPercentage(poll)

  return (
    <Box {...props}>
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
    </Box>
  )
}

const participationPercentage = (poll: PollInfo) => {
  if (!poll || !poll.censusParticipantsCount) return 0

  return ((poll.voteCount / poll.censusParticipantsCount) * 100).toFixed(1)
}
