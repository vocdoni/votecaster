import { ethers } from 'ethers'
import { mainnet } from 'viem/chains'
import { isAddress, createPublicClient, http } from 'viem'

import { appUrl, degenChainRpc } from '~constants'
import { useQuery } from '@tanstack/react-query'

const degenNameResolverAbiJson =
  '[{"inputs":[{"internalType":"string","name":"_domainName","type":"string"}],"name":"getDomainHolder","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"defaultNames","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"}]'

export const fetchUserProfile = (bfetch: FetchFunction, username: string) => async (): Promise<UserProfileResponse> => {
  const response = await bfetch(`${appUrl}/profile/user/${username}`)
  const user = (await response.json()) as UserProfileResponse

  return user
}

export const fetchUserPolls = (bfetch: FetchFunction, username: string) => async (): Promise<Poll[]> => {
  const response = await bfetch(`${appUrl}/profile/user/${username}`)
  const { polls } = (await response.json()) as { polls: Poll[] }
  if (!polls) {
    throw new Error('Received no elections')
  }
  return polls.map((poll) => ({
    ...poll,
    createdByUsername: username,
  }))
}

export const fetchMutedUsers = (bfetch: FetchFunction) => async (): Promise<Profile[]> => {
  const response = await bfetch(`${appUrl}/profile`)
  const data = await response.json()
  return data.mutedUsers
}

export const fetchWarpcastAPIEnabled = (bfetch: FetchFunction) => async (): Promise<boolean> => {
  const response = await bfetch(`${appUrl}/profile`)
  const data = await response.json()
  return data.warpcastApiEnabled
}

const publicClient = createPublicClient({
  chain: mainnet,
  transport: http(),
})

const getDegenNameContract = () => {
  const provider = new ethers.JsonRpcProvider(degenChainRpc)

  return new ethers.Contract('0x4087fb91A1fBdef05761C02714335D232a2Bf3a1', degenNameResolverAbiJson, provider)
}

const fetchDegenOrEnsName = async (addr: string): Promise<string | null> => {
  if (!isAddress(addr)) {
    return null
  }

  const degenNameContract = getDegenNameContract()
  const degenName = await degenNameContract.defaultNames(addr)
  if (degenName) {
    return `${degenName}.degen`
  }

  return publicClient.getEnsName({ address: addr })
}

export const getProfileAddresses = (p?: UserProfileResponse) => {
  return p?.user.verifications ?? p?.user.addresses ?? []
}

export const useFirstDegenOrEnsName = (addresses: string[] = []) => {
  // Process the addresses to ensure a consistent react query function key
  const sortedAddresses = addresses
    .map((v) => v.toLowerCase())
    .sort((a, b) => {
      if (a > b) {
        return 1
      } else if (a < b) {
        return -1
      } else {
        return 0
      }
    })

  return useQuery({
    queryKey: ['firstDegenOrEnsName', ...sortedAddresses],
    queryFn: async () => {
      const degenOrEnsNames = await Promise.all(sortedAddresses.map(async (addr) => fetchDegenOrEnsName(addr)))
      const firstValidName = degenOrEnsNames.find((v) => !!v)
      return firstValidName || null
    },
  })
}

export const useUserDegenOrEnsName = (user?: UserProfileResponse) => {
  return useFirstDegenOrEnsName(getProfileAddresses(user))
}
