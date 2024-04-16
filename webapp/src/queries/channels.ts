import { appUrl } from '../util/constants'
import { FetchFunction } from '../util/types'

type Channel = {
  description: string
  followerCount: number
  id: string
  image: string
  name: string
  url: string
}

export const fetchChannelQuery = (bfetch: FetchFunction) => async (inputValue: string) => {
  const response = await bfetch(`${appUrl}/channels?q=${encodeURIComponent(inputValue)}`)
  const { channels } = (await response.json()) as { channels: Channel[] }

  return channels
}
