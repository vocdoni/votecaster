
import { appUrl } from '../util/constants'
import type { FetchFunction, PollInfo } from '../util/types'

export const fetchPollInfo = (bfetch: FetchFunction) => async (electionID: string): Promise<PollInfo> => {
  const response = await bfetch(`${appUrl}/poll/info/${electionID}`)
  const {poll} = (await response.json()) as { poll: PollInfo }
  return poll
}

export const fetchPollsVoters = (bfetch: FetchFunction) => async (electionId: string): Promise<string[]> => {
  const response = await bfetch(`${appUrl}/votersOf/${electionId}`)
  const {voters} = await response.json() as { voters: string[] }
  return voters
}