import { appUrl } from '~constants'

export const fetchPollInfo = (bfetch: FetchFunction, electionID: string) => async (): Promise<PollResponse> => {
  const response = await bfetch(`${appUrl}/poll/info/${electionID}`)
  const { poll } = (await response.json()) as { poll: PollResponse }
  return poll
}

export const fetchPollsVoters = (bfetch: FetchFunction, electionId: string) => async (): Promise<string[]> => {
  const response = await bfetch(`${appUrl}/votersOf/${electionId}`)
  const { usernames } = (await response.json()) as { usernames: string[] }
  return usernames
}

export const fetchPollsRemainingVoters = (bfetch: FetchFunction, electionId: string) => async (): Promise<string[]> => {
  const response = await bfetch(`${appUrl}/remainingVotersOf/${electionId}`)
  const { usernames } = (await response.json()) as { usernames: string[] }
  return usernames
}
