import { useQuery } from '@tanstack/react-query'
import { useAuth } from '~components/Auth/useAuth'
import { appUrl } from '~constants'

export const fetchPollsByVotes = (bfetch: FetchFunction) => async (): Promise<PollRanking[]> => {
  const response = await bfetch(`${appUrl}/rankings/pollsByVotes`)
  const { polls } = (await response.json()) as { polls: PollRanking[] }
  return polls
}

export const fetchTopVoters = (bfetch: FetchFunction) => async (): Promise<UserRanking[]> => {
  const response = await bfetch(`${appUrl}/rankings/usersByCastedVotes`)
  const { users } = (await response.json()) as { users: UserRanking[] }
  return users
}

export const fetchTopCreators = (bfetch: FetchFunction) => async (): Promise<UserRanking[]> => {
  const response = await bfetch(`${appUrl}/rankings/usersByCreatedPolls`)
  const { users } = (await response.json()) as { users: UserRanking[] }
  return users
}

export const latestPolls =
  (bfetch: FetchFunction, { limit = 10 }: Partial<Pagination> = {}) =>
  async (): Promise<PollRanking[]> => {
    const response = await bfetch(`${appUrl}/rankings/latestPolls?limit=${limit}`)
    const { polls } = (await response.json()) as { polls: PollRanking[] }
    return polls
  }

export const fetchPollsByCommunity = (bfetch: FetchFunction, community: Community) => async (): Promise<Poll[]> => {
  const response = await bfetch(`${appUrl}/rankings/pollsByCommunity/${community.id}`)
  const { polls } = (await response.json()) as { polls: Poll[] }
  return polls
}

export const fetchPoints = (bfetch: FetchFunction) => async () => {
  const response = await bfetch(`${appUrl}/rankings/points`)
  const json = await response.json()
  const { points } = json as { points: PointsLeaderboard[] }
  return points
}

export const useFetchPoints = () => {
  const { bfetch } = useAuth()
  return useQuery({
    queryKey: ['rankings', 'points'],
    queryFn: fetchPoints(bfetch),
  })
}
