export const fetchUserPolls = (bfetch, profile) => async (): Promise<Poll[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/auth/check`)
  const { elections } = (await response.json()) as { elections: Poll[] }
  if (!elections) {
    throw new Error('Received no elections')
  }
  return elections.map((poll) => ({
    ...poll,
    createdByUsername: profile?.username,
  }))
}

export const fetchMutedUsers = (bfetch) => async (): Promise<Profile[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/auth/check`)
  const data = await response.json()
  return data.mutedUsers
}
