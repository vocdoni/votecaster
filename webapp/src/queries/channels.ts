import { appUrl } from '~constants'

type Channel = {
  description: string
  followerCount: number
  id: string
  image: string
  name: string
  url: string
}

export const fetchChannel = (bfetch: FetchFunction) => async (channelId: string) => {
  const response = await bfetch(`${appUrl}/channels/${channelId}`)
  const channel = (await response.json()) as Channel

  return channel
}

export const fetchChannelQuery = (bfetch: FetchFunction) => async (inputValue: string) => {
  const response = await bfetch(`${appUrl}/channels?q=${encodeURIComponent(inputValue)}`)
  const { channels } = (await response.json()) as { channels: Channel[] }

  return channels
}
