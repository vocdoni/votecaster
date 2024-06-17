import sha256 from 'crypto-js/sha256'

export const cleanChannel = (channel: string) => channel.replace(/.*channel\//, '')

export const ucfirst = (str: string) => str.charAt(0).toUpperCase() + str.slice(1)

export const urlValidation = (val: string) => /^(https?|ipfs):\/\//.test(val)

export const humanDate = (date?: Date, default_content?: string): string => {
  if (!date) return default_content || ''
  date = new Date(date)
  const days = date.getDate().toString().padStart(2, '0')
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  const month = months[date.getMonth()]
  const year = date.getFullYear().toString()
  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')

  return `${days} ${month}, ${year} ${hours}:${minutes}`
}

export const hashString = async (str: string) => {
  const full = sha256(str).toString()
  return full.substring(0, 14)
}

export const shortHex = (hex: string) => hex.substring(0, 6) + '...' + hex.substring(hex.length - 4)

export const camelize = (str: string) => str.toLowerCase().replace(/[^a-zA-Z0-9]+(.)/g, (_, chr) => chr.toUpperCase())
