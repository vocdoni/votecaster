import { appUrl } from '../util/constants'
import { Address, FetchFunction, Profile } from '../util/types'

export type Community = {
  id: number
  name: string
  logoURL: string
  admins: Profile[]
  notifications: boolean
  censusName: string
  censusType: string
  censusAddresses: Address[]
}

export const fetchCommunities = (bfetch: FetchFunction) => async () => {
  const response = await bfetch(`${appUrl}/communities`)
  const communities = (await response.json()) as Community[]

  return communities
}
