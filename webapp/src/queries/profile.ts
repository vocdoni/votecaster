import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { isAddress } from 'viem'
import { degen, mainnet } from 'viem/chains'
import abi from '~abis/nftdegen.json'
import { useAuth } from '~components/Auth/useAuth'
import { useBlockchain } from '~components/Blockchains/BlockchainContext'
import { useHealthcheck } from '~components/Healthcheck/use-healthcheck'
import { appUrl, degenNameResolverContractAddress } from '~constants'

export const fetchUserProfile =
  (bfetch: FetchFunction, username: string | null) => async (): Promise<UserProfileResponse> => {
    const url = username ? `${appUrl}/profile/user/${username}` : `${appUrl}/profile`
    const response = await bfetch(url)
    const user = (await response.json()) as UserProfileResponse

    return user
  }

export const useUserProfile = (username?: string) => {
  const { bfetch } = useAuth()

  return useQuery<UserProfileResponse, Error>({
    queryKey: ['profile', username],
    queryFn: fetchUserProfile(bfetch, username ?? null),
    enabled: !!bfetch,
  })
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

export const fetchWarpcastAPIEnabled = (bfetch: FetchFunction) => async (): Promise<boolean> => {
  const response = await bfetch(`${appUrl}/profile`)
  const data = await response.json()
  return data.warpcastApiEnabled
}

export const useWarpcastApiEnabled = () => {
  const { bfetch } = useAuth()

  return useQuery<boolean, Error>({
    queryKey: ['apiKeyEnabled'],
    queryFn: fetchWarpcastAPIEnabled(bfetch),
  })
}

export const useFirstDegenOrEnsName = (addresses: string[] = []) => {
  const { degen: connected } = useHealthcheck()
  const { getContract } = useBlockchain(degen)
  const { client } = useBlockchain(mainnet)

  const contract = getContract(degenNameResolverContractAddress, abi)
  // Process the addresses to ensure a consistent react query function key
  const sortedAddresses = addresses.map((v) => v.toLowerCase()).sort((a, b) => a.localeCompare(b))

  return useQuery({
    queryKey: ['firstDegenOrEnsName', ...sortedAddresses],
    retry: connected,
    queryFn: async () => {
      const degenOrEnsNames = await Promise.all(
        sortedAddresses.map(async (addr) => {
          if (!isAddress(addr)) {
            return null
          }

          const degenName = await contract.read.defaultNames([addr])
          if (degenName) {
            return `${degenName}.degen`
          }

          return client.getEnsName({ address: addr })
        })
      )
      const firstValidName = degenOrEnsNames.find((v) => !!v)
      return firstValidName || null
    },
  })
}

export const getProfileAddresses = (p?: UserProfileResponse) => {
  return p?.user.addresses
}

export const useUserDegenOrEnsName = (user?: UserProfileResponse) => {
  return useFirstDegenOrEnsName(getProfileAddresses(user))
}

export const useFetchProfileMutation = () => {
  const { bfetch } = useAuth()

  return useMutation({
    mutationKey: ['profile'],
    mutationFn: async (userId: number | string) => {
      const response = await bfetch(`${appUrl}/profile/fid/${Number(userId)}`)
      if (!response.ok) {
        throw new Error('Failed to fetch delegated user')
      }
      const { user } = (await response.json()) as UserProfileResponse
      return user
    },
  })
}

export const useDelegateVote = () => {
  const queryClient = useQueryClient()
  const { bfetch } = useAuth()

  return useMutation({
    mutationKey: ['delegate'],
    mutationFn: async ({ to, communityId }: { to: string; communityId: string }) => {
      const userResponse = await bfetch(`${appUrl}/profile/user/${to}`)
      if (!userResponse.ok) {
        throw new Error('User not found')
      }
      const { user } = (await userResponse.json()) as UserProfileResponse

      const response = await bfetch(`${appUrl}/profile/delegation`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          to: user.userID,
          communityId,
        }),
      })

      if (!response.ok) {
        throw new Error('Delegation failed')
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['delegations'],
      })
    },
  })
}

export const useRevokeDelegation = () => {
  const queryClient = useQueryClient()
  const { bfetch } = useAuth()

  return useMutation({
    mutationFn: async (delegationId: string) => {
      const response = await bfetch(`${appUrl}/profile/delegation/${delegationId}`, {
        method: 'DELETE',
      })
      if (!response.ok) {
        throw new Error('Failed to revoke delegation')
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['delegations'],
      })
    },
  })
}

export const useMuteUser = () => {
  const queryClient = useQueryClient()
  const { bfetch } = useAuth()

  return useMutation({
    mutationFn: async (username: string) => {
      const response = await bfetch(`${appUrl}/profile/mutedUsers`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username }),
      })

      if (!response.ok) {
        throw new Error('Failed to mute user')
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['profile'],
      })
    },
  })
}

export const useUnmuteUser = () => {
  const queryClient = useQueryClient()
  const { bfetch } = useAuth()

  return useMutation({
    mutationFn: async (username: string) => {
      const response = await bfetch(`${appUrl}/profile/mutedUsers/${username}`, {
        method: 'DELETE',
      })

      if (!response.ok) {
        throw new Error('Failed to unmute user')
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['profile'],
      })
    },
  })
}
