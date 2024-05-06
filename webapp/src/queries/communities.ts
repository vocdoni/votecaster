import { appUrl, paginationItemsPerPage } from '~constants'

export const fetchCommunities =
  (bfetch: FetchFunction, { limit = paginationItemsPerPage, offset = 0 }) =>
  async () => {
    const response = await bfetch(`${appUrl}/communities?limit=${limit}&offset=${offset}`)
    const { communities, pagination } = (await response.json()) as { communities: Community[]; pagination: Pagination }

    return { communities, pagination }
  }

export const fetchCommunitiesByAdmin =
  (bfetch: FetchFunction, profile: Profile, { limit = paginationItemsPerPage, offset = 0 }) =>
  async () => {
    const response = await bfetch(`${appUrl}/communities?byAdminFID=${profile.fid}&limit=${limit}&offset=${offset}`)
    const { communities, pagination } = (await response.json()) as { communities: Community[]; pagination: Pagination }

    return { communities, pagination }
  }

export const fetchCommunity = (bfetch: FetchFunction, id: string) => async () => {
  const response = await bfetch(`${appUrl}/communities/${id}`)
  const community = (await response.json()) as Community

  return community
}

export const updateCommunity = async (bfetch: FetchFunction, community: Community) => {
  const response = await bfetch(`${appUrl}/communities/${community.id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(community),
  })
  return response.json()
}
