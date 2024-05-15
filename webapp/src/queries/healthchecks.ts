import { degen } from 'wagmi/chains'
import { degenChainRpc } from '~constants'

export const degenHealth = async (): Promise<boolean> => {
  try {
    const response = await fetch(degenChainRpc, {
      method: 'POST',
      body: JSON.stringify({
        method: 'eth_blockNumber',
        id: degen.id,
        jsonrpc: '2.0',
      }),
    })
    if (!response.ok) {
      throw new Error('Degen RPC health check failed')
    }
    return true
  } catch (e) {
    console.error('Degen RPC health check failed', e)
    return false
  }
}
