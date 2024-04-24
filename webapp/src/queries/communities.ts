import {appUrl} from '../util/constants'
import {Community, FetchFunction, Profile} from '../util/types'

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
