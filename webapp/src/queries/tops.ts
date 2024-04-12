export type Poll = {
  electionId: string
  title: string
  createdByUsername: string
  createdByDisplayname: string
  voteCount: number
  createdTime: Date
  lastVoteTime: Date
}

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

export const fetchTopVoters = (bfetch) => async (): Promise<UserRanking[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/rankings/usersByCastedVotes`)
  const { users } = (await response.json()) as { users: UserRanking[] }
  return users
}

export const fetchTopCreators = (bfetch) => async (): Promise<UserRanking[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/rankings/usersByCreatedPolls`)
  const { users } = (await response.json()) as { users: UserRanking[] }
  return users
}

export const fetchLatestPolls = (bfetch) => async (): Promise<Poll[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/rankings/lastElections`)
  const { polls } = (await response.json()) as { polls: Poll[] }
  return polls
}
