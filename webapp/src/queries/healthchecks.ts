import { degenChainRpc } from '~constants'

export const degenHealth = async (): Promise<boolean> => {
  const response = await fetch(degenChainRpc)
  if (!response.ok) {
    throw new Error('Degen RPC health check failed')
  }
  return true
}
