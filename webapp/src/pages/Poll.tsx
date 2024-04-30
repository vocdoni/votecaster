import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'
import { PollView } from '~components/Poll'
import { fetchPollInfo, fetchPollsVoters } from '~queries/polls'

const Poll = () => {
  const { pid: electionId } = useParams()
  const { bfetch } = useAuth()

  const {
    data: results,
    isLoading: rLoading,
    error,
  } = useQuery<PollResponse, Error, PollInfo>({
    queryKey: ['poll', electionId],
    queryFn: fetchPollInfo(bfetch, electionId!),
    enabled: !!electionId && electionId.length > 0,
    select: (data: PollResponse) => ({
      ...data,
      endTime: new Date(data.endTime),
      lastVoteTime: new Date(data.lastVoteTime),
      createdTime: new Date(data.createdTime),
      tally: data.tally ? [data.tally.map((t) => parseInt(t))] : [[]],
      totalWeight: Number(data.totalWeight),
    }),
  })

  const {
    data: voters,
    isLoading: vLoading,
    error: vError,
  } = useQuery({
    queryKey: ['voters', electionId],
    queryFn: fetchPollsVoters(bfetch, electionId!),
    enabled: !!results && results?.voteCount > 0,
  })

  return (
    <PollView
      onChain={false}
      loading={rLoading || vLoading}
      poll={results || null}
      voters={voters || []}
      errorMessage={error?.toString() || vError?.toString() || null}
      electionId={electionId}
    />
  )
}

export default Poll
