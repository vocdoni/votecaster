import { Poll } from '../components/Top'

export const fetchPollsByVotes = (bfetch) => async (): Promise<Poll[]> => {
  const response = await bfetch(`${import.meta.env.APP_URL}/rankings/pollsByVotes`)
  const { polls } = (await response.json()) as { polls: Poll[] }
  return polls
}
