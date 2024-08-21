import { appUrl } from '~constants'

export const fetchTokenBasedBlockchains = (bfetch: FetchFunction) => async () => {
  const response = await bfetch(`${appUrl}/census/community/blockchains`)
  const { blockchains } = (await response.json()) as { blockchains: string[] }
  return blockchains.sort()
}

type CensusResponse = Omit<Census, 'totalWeight'> & {
  totalWeight: string
}

export const fetchCensus = (bfetch: FetchFunction, id: string) => async (): Promise<Census> => {
  const response = await bfetch(`${appUrl}/census/${id}`)
  const census = (await response.json()) as CensusResponse

  return {
    ...census,
    totalWeight: Number(census.totalWeight),
  }
}
