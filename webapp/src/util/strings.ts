export const cleanChannel = (channel: string) => channel.replace(/.*channel\//, '')

export const ucfirst = (str: string) => str.charAt(0).toUpperCase() + str.slice(1)
