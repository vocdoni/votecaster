import type { FetchFunction } from '../util/types'

export const fetchAirstackBlockchains = (bfetch: FetchFunction) => async () => {
  const response = await bfetch(`${import.meta.env.APP_URL}/census/airstack/blockchains`)
  const { blockchains } = (await response.json()) as { blockchains: string[] }
  return blockchains.sort()
}
