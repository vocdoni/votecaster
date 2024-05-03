import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'
import { Check } from '~components/Check'
import { PollView } from '~components/Poll'
import { fetchPollInfo } from '~queries/polls'

const Poll = () => {
  const { pid: electionId } = useParams()
  const { bfetch } = useAuth()

  const { data, isLoading, error } = useQuery<PollResponse, Error, PollInfo>({
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

  if (error || isLoading) {
    return <Check error={error} isLoading={isLoading} />
  }

  if (!data && !isLoading) {
    return <Check error={new Error('No results found')} isLoading={false} />
  }

  return <PollView onChain={false} loading={isLoading} poll={data as PollInfo} />
}

export default Poll
