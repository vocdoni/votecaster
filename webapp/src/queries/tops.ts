import type { Community, FetchFunction, Poll } from '../util/types'

export const fetchPollsByVotes = (bfetch: FetchFunction) => async (): Promise<Poll[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/rankings/pollsByVotes`)
  const { polls } = (await response.json()) as { polls: Poll[] }
  return polls
}

export type UserRanking = {
  fid: number
  username: string
  count: number
  displayName: string
}

export const fetchTopVoters = (bfetch: FetchFunction) => async (): Promise<UserRanking[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/rankings/usersByCastedVotes`)
  const { users } = (await response.json()) as { users: UserRanking[] }
  return users
}

export const fetchTopCreators = (bfetch: FetchFunction) => async (): Promise<UserRanking[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/rankings/usersByCreatedPolls`)
  const { users } = (await response.json()) as { users: UserRanking[] }
  return users
}

export const fetchLatestPolls = (bfetch: FetchFunction) => async (): Promise<Poll[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/rankings/lastElections`)
  const { polls } = (await response.json()) as { polls: Poll[] }
  return polls
}

export const fetchPollsByCommunity = (bfetch: FetchFunction, community: Community) => async (): Promise<Poll[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/rankings/pollsByCommunity/${community.id}`)
  const { polls } = (await response.json()) as { polls: Poll[] }
  return polls
}
