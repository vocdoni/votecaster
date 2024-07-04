// healthcheck.ts
export const chainHealth = async (rpcUrl: string, chainId: number): Promise<boolean> => {
  try {
    const response = await fetch(rpcUrl, {
      method: 'POST',
      body: JSON.stringify({
        method: 'eth_blockNumber',
        id: chainId,
        jsonrpc: '2.0',
      }),
    })
    if (!response.ok) {
      throw new Error(`Chain ${chainId} RPC health check failed`)
    }
    return true
  } catch (e) {
    console.error(`Chain ${chainId} RPC health check failed`, e)
    return false
  }
}
