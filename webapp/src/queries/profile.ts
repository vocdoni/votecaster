import { useQuery } from '@tanstack/react-query'
import { createPublicClient, getContract, http, isAddress } from 'viem'
import { degen, mainnet } from 'viem/chains'
import abi from '~abis/nftdegen.json'
import { useDegenHealthcheck } from '~components/Healthcheck/use-healthcheck'
import { appUrl, degenNameResolverContractAddress } from '~constants'

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

const mainnetClient = createPublicClient({
  chain: mainnet,
  transport: http(),
})

const degenClient = createPublicClient({
  chain: degen,
  transport: http(),
})

const getDegenNameContract = () => {
  return getContract({
    address: degenNameResolverContractAddress,
    client: {
      public: degenClient,
    },
    abi,
  })
}

const fetchDegenOrEnsName = async (addr: string): Promise<string | null> => {
  if (!isAddress(addr)) {
    return null
  }

  const degenNameContract = getDegenNameContract()
  const degenName = await degenNameContract.read.defaultNames([addr])
  if (degenName) {
    return `${degenName}.degen`
  }

  return mainnetClient.getEnsName({ address: addr })
}

export const getProfileAddresses = (p?: UserProfileResponse) => {
  return p?.user.verifications ?? p?.user.addresses ?? []
}

export const useFirstDegenOrEnsName = (addresses: string[] = []) => {
  const { connected } = useDegenHealthcheck()
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
    retry: connected,
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
