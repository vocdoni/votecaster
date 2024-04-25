import { appUrl } from '~constants'

export const fetchAirstackBlockchains = (bfetch: FetchFunction) => async () => {
  const response = await bfetch(`${appUrl}/census/airstack/blockchains`)
  const { blockchains } = (await response.json()) as { blockchains: string[] }
  return blockchains.sort()
}
