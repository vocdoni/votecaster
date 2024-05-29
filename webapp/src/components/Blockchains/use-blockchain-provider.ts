import { useEffect, useState } from 'react'
import {
  Abi,
  Address,
  Chain,
  createPublicClient,
  getContract,
  GetContractReturnType,
  http,
  PublicClient,
  Transport,
} from 'viem'
import { BlockchainContextState } from './BlockchainContextState'
import { useBlockchainRegistry } from './use-blockchain-registry'

export interface BlockchainProviderProps {
  chain: Chain
}

export const useBlockchainProvider = ({ chain }: BlockchainProviderProps): BlockchainContextState => {
  const [client, setClient] = useState(() => createPublicClient({ chain, transport: http() }))
  const { registerContext } = useBlockchainRegistry()

  const getContractInstance = <TAbi extends Abi | readonly unknown[]>(
    address: Address,
    abi: TAbi
  ): GetContractReturnType<TAbi, PublicClient<Transport, Chain>, Address> =>
    getContract({
      address,
      client,
      abi,
    })

  useEffect(() => {
    setClient(createPublicClient({ chain, transport: http() }))
    registerContext({ client, getContract: getContractInstance })
  }, [chain])

  return {
    client,
    getContract: getContractInstance,
  }
}
