import { HStack, Link, Text, VStack } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { vocdoniExplorer } from '~constants'
import { fetchCommunity } from '~queries/communities'
import { humanDate } from '~util/strings'
import { useAuth } from '../Auth/useAuth'
import { ParticipantsTableModal } from './ParticipantsTableModal'
import { PollRemindersModal } from './PollRemindersModal'
import { RemainingVotersTableModal } from './RemainingVotersTableModal'
import { VotersTableModal } from './VotersTableModal'

export const Information = ({ poll }: { poll?: PollInfo }) => {
  const { profile, bfetch } = useAuth()
  const {data: community} = useQuery({
    queryKey: ['community', poll?.community?.id],
    queryFn: fetchCommunity(bfetch, poll?.community?.id.toString() || ''),
    enabled: !!poll?.community?.id.toString(),
  })

  const isAdmin = () =>{
    if (!profile || !community) return false
    return community.admins.some((admin) => admin.fid === profile.fid)
  }

  if (!poll) return

  return (
    <VStack spacing={6} alignItems='left' fontSize={'sm'}>
      <Text>
        This poll {poll?.finalized ? 'has ended' : 'ends'} on {`${humanDate(poll?.endTime)}`}.{` `}
        <Link variant='primary' isExternal href={`${vocdoniExplorer}/processes/show/#/${poll.electionId}`}>
          Check the Vocdoni blockchain explorer
        </Link>
        {` `}for more information.
      </Text>
      {!!poll.censusParticipantsCount && (
        <>
          <Text>
            Download the list of members who have already cast their votes, the list of remaining members who still need
            to vote, and the total census of eligible voters.
          </Text>
          <HStack spacing={2} flexWrap='wrap'>
            <VotersTableModal poll={poll} />
            <RemainingVotersTableModal poll={poll} />
            <ParticipantsTableModal poll={poll} />
            {!!poll.community && isAdmin() && <PollRemindersModal poll={poll}/>}
          </HStack>
        </>
      )}
    </VStack>
  )
}
