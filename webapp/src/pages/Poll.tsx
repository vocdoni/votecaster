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
  } = useQuery({
    queryKey: ['poll', electionId],
    queryFn: fetchPollInfo(bfetch, electionId),
    select: (data) => ({
      ...data,
      censusRoot: '',
      censusURI: '',
      endTime: new Date(data.endTime),
      tally: data.tally ? [data.tally.map((t) => parseInt(t))] : [[]],
    }),
  })

  const {
    data: voters,
    isLoading: vLoading,
    error: vError,
  } = useQuery({
    queryKey: ['voters', electionId],
    queryFn: fetchPollsVoters(bfetch, electionId),
    enabled: !!results && results?.voteCount > 0,
  })

  return (
    <PollView
      onChain={false}
      loading={rLoading || vLoading}
      poll={results}
      voters={voters || []}
      errorMessage={error || vError}
      electionId={electionId}
    />
  )
}

export default Poll
