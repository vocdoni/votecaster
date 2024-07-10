import { Chain, getDefaultConfig } from '@rainbow-me/rainbowkit'
import { BrowserProvider } from 'ethers'
import { UseWalletClientReturnType } from 'wagmi'

export const config = getDefaultConfig({
  appName: 'farcaster.vote',
  projectId: '735ab19f8bdb36d6ab32328218ded4ac',
  chains: Object.values(import.meta.env.chains) as unknown as [Chain, ...Chain[]],
})

export const walletClientToSigner = async (walletClient: UseWalletClientReturnType['data']) => {
  const { account, chain, transport } = walletClient!
  const network = {
    chainId: chain.id,
    name: chain.name,
    ensAddress: chain.contracts?.ensRegistry?.address,
  }
  const provider = new BrowserProvider(transport, network)
  const signer = await provider.getSigner(account.address)
  return signer
}
