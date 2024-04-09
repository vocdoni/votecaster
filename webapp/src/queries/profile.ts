export const fetchUserPolls = (bfetch, profile) => async (): Promise<Poll[]> => {
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

export const fetchMutedUsers = (bfetch) => async (): Promise<Profile[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/profile`)
  const data = await response.json()
  return data.mutedUsers
}
