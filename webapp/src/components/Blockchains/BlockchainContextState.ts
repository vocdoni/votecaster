import { Abi, Address, Chain, GetContractReturnType, PublicClient, Transport } from 'viem'

export interface BlockchainContextState {
  client: PublicClient<Transport, Chain>
  getContract: <TAbi extends Abi | readonly unknown[]>(
    address: Address,
    abi: TAbi
  ) => GetContractReturnType<TAbi, PublicClient<Transport, Chain>, Address>
}
