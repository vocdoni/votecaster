import { useQuery } from '@tanstack/react-query'
import { ethers } from 'ethers'
import { useAuth } from '~components/Auth/useAuth'
import { useDegenHealthcheck } from '~components/Healthcheck/use-healthcheck'
import { appUrl, degenChainRpc, degenContractAddress } from '~constants'
import { CommunityHub__factory } from '~typechain'
import { toArrayBuffer } from '~util/hex'

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

export const fetchPollsReminders = (bfetch: FetchFunction, electionId: string) => async (): Promise<PollReminders> => {
  const response = await bfetch(`${appUrl}/poll/${electionId}/reminders`)
  const data = await response.json()
  const remindableVoters: Profile[] = []
  for (const fid in data.remindableVoters) {
    remindableVoters.push({
      fid: parseInt(fid),
      username: data.remindableVoters[fid]
    } as Profile)
  }

  const votersWeight: [username: string, weight: string] = {} as [string, string]
  for (const fid in data.votersWeight) {
    votersWeight[data.remindableVoters[fid]] = data.votersWeight[fid]
  }
  return {
    remindableVoters,
    alreadySent: data.alreadySent,
    maxReminders: data.maxReminders,
    votersWeight: votersWeight,
  } as PollReminders
}

export const fetchShortURL = (bfetch: FetchFunction) => async (url: string) => {
  const response = await bfetch(`${appUrl}/short?url=${url}`)
  const { result } = (await response.json()) as { result: string }
  return result
}

export const useApiPollInfo = (electionId?: string) => {
  const { bfetch } = useAuth()

  return useQuery<PollResponse, Error, PollInfo>({
    queryKey: ['apiPollInfo', electionId],
    queryFn: fetchPollInfo(bfetch, electionId!),
    enabled: !!electionId,
    select: (data) => ({
      ...data,
      totalWeight: Number(data.totalWeight),
      createdTime: new Date(data.createdTime),
      endTime: new Date(data.endTime),
      lastVoteTime: new Date(data.lastVoteTime),
      tally: [data.tally.map((t) => Number(t))],
    }),
  })
}

export const useContractPollInfo = (communityId?: string, electionId?: string) => {
  const { connected } = useDegenHealthcheck()
  return useQuery({
    queryKey: ['contractPollInfo', communityId, electionId],
    queryFn: async () => {
      const provider = new ethers.JsonRpcProvider(degenChainRpc, undefined, { polling: false, staticNetwork: true })
      try {
        const contract = CommunityHub__factory.connect(degenContractAddress, provider)
        const contractData = await contract.getResult(communityId!, toArrayBuffer(electionId!))
        return contractData
      } catch (e) {
        provider.destroy()
        throw e
      }
    },
    retry: connected,
    enabled: !!communityId && !!electionId,
  })
}
