import {appUrl} from '../util/constants'
import {Address, FetchFunction, Profile} from '../util/types'

export type Channel = {
  id: string
  name: string
  description: string
  followerCount: number
  image: string
  url: string
}

export type Community = {
  id: number
  name: string
  logoURL: string
  admins: Profile[]
  notifications: boolean
  censusType: string
  censusAddresses: Address[]
  channels: Channel[]
}

export const fetchCommunities = (bfetch: FetchFunction) => async () => {
  const response = await bfetch(`${appUrl}/communities`)
  const {communities} = (await response.json()) as { communities: Community[] }

  return communities
}

export const fetchCommunity = (bfetch: FetchFunction, id: string) => async () => {
  const response = await bfetch(`${appUrl}/communities/${id}`)
  const community = (await response.json()) as Community

  return community
}
