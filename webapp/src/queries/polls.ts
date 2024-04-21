
import { appUrl } from '../util/constants'
import type { FetchFunction, PollInfo } from '../util/types'

export const fetchPollInfo = (bfetch: FetchFunction) => async (electionID: string): Promise<PollInfo> => {
  const response = await bfetch(`${appUrl}/polls/info/${electionID}`)
  return await response.json()
}