import { Poll } from '../components/Top'

export const fetchPollsByVotes = (bfetch) => async (): Promise<Poll[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/rankings/pollsByVotes`)
  const { polls } = (await response.json()) as { polls: Poll[] }
  return polls
}

export type UserRanking = {
  fid: number
  username: string
  count: number
}

export const fetchTopVoters = (bfetch) => async (): Promise<Poll[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/rankings/usersByCastedVotes`)
  const { users } = (await response.json()) as { users: UserRanking[] }
  return users
}

export const fetchTopCreators = (bfetch) => async (): Promise<Poll[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/rankings/usersByCreatedPolls`)
  const { users } = (await response.json()) as { users: UserRanking[] }
  return users
}
