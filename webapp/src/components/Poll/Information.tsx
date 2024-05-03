import { HStack, Link, Text, VStack } from '@chakra-ui/react'
import { humanDate } from '~util/strings'
import { DownloadRemainingVotersButton, DownloadVotersButton } from './DownloadButtons'

export const Information = ({ poll }: { poll?: PollInfo }) => {
  if (!poll) return
  return (
    <VStack spacing={6} alignItems='left' fontSize={'sm'}>
      <Text>
        This poll {poll?.finalized ? 'has ended' : 'ends'} on {`${humanDate(poll?.endTime)}`}.{` `}
        <Link variant='primary' isExternal href={`https://stg.explorer.vote/processes/show/#/${poll.electionId}`}>
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
  )
}
