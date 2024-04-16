import type { FetchFunction, Poll, Profile } from '../util/types'

export const fetchUserPolls = (bfetch: FetchFunction, profile: Profile) => async (): Promise<Poll[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/profile`)
  const { polls } = (await response.json()) as { polls: Poll[] }
  if (!polls) {
    throw new Error('Received no elections')
  }
  return polls.map((poll) => ({
    ...poll,
    createdByUsername: profile?.username,
  }))
}

export const fetchMutedUsers = (bfetch: FetchFunction) => async (): Promise<Profile[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/profile`)
  const data = await response.json()
  return data.mutedUsers
}
