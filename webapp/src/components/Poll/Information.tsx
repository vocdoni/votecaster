import { useEffect } from 'react'
import { HStack, Link, Text, VStack, useToast } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { vocdoniExplorer, adminFID } from '~constants'
import { fetchCommunity } from '~queries/communities'
import { fetchCensus } from '~queries/census'
import { humanDate } from '~util/strings'
import { useAuth } from '../Auth/useAuth'
import { ParticipantsTableModal } from './ParticipantsTableModal'
import { PollRemindersModal } from './PollRemindersModal'
import { RemainingVotersTableModal } from './RemainingVotersTableModal'
import { VotersTableModal } from './VotersTableModal'

export const Information = ({ poll, url }: { poll: PollInfo, url?: string }) => {
  const { profile, bfetch } = useAuth()
  const toast = useToast()
  const {data: community} = useQuery({
    queryKey: ['community', poll?.community?.id],
    queryFn: fetchCommunity(bfetch, poll?.community?.id.toString() || ''),
    enabled: !!poll?.community?.id.toString(),
  })

    const { data: census, error: errorCensus } = useQuery({
      queryKey: ['census', poll.electionId],
      queryFn: fetchCensus(bfetch, poll.electionId),
      enabled: !!poll.electionId,
      refetchOnWindowFocus: false,
      retry: (count, error: any) => {
        if (error.status !== 200) {
          return count < 1
        }
        return false
      },
    })

  useEffect(() => {
    if (!errorCensus) return

    toast({
      title: 'Error',
      description: errorCensus?.message || 'Failed to retrieve remaining voters list',
      status: 'error',
      duration: 5000,
      isClosable: true,
    })
  }, [errorCensus])

  if (!poll) return;

  const isAdmin = () => {
    if (!profile || !community) return false;
    return community.admins.some((admin) => admin.fid === profile.fid) || profile.fid === adminFID
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
            {!!census && <>
              <VotersTableModal poll={poll} census={census} />
              <RemainingVotersTableModal poll={poll} census={census} />
              <ParticipantsTableModal poll={poll} census={census} />
            </>}
              {!!poll.community && !poll?.finalized && isAdmin() && <PollRemindersModal poll={poll} frameURL={url} />}
          </HStack>
        </>
      )}
    </VStack>
  )
}
