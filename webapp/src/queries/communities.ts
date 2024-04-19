import {appUrl} from '../util/constants'
import {Address, FetchFunction, Profile} from '../util/types'

export type Community = {
  id: number
  name: string
  logoURL: string
  admins: Profile[]
  notifications: boolean
  censusType: string
  censusAddresses: Address[]
  channels: string[]
  groupChat: string
}

export const fetchCommunities = (bfetch: FetchFunction) => async () => {
  const response = await bfetch(`${appUrl}/communities`)
  const {communities} = (await response.json()) as { communities: Community[] }

  return communities
}

export const fetchCommunitiesByAdmin = async (bfetch: FetchFunction, profile: Profile) => {
  const response = await bfetch(`${appUrl}/communities?byAdminFID=${profile.fid}`)
  const {communities} = (await response.json()) as { communities: Community[] }

  return communities
}

export const fetchCommunity = (bfetch: FetchFunction, id: string) => async () => {
  const response = await bfetch(`${appUrl}/communities/${id}`)
  const community = (await response.json()) as Community

  return community
}
