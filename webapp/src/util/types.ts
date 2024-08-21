// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const isErrorWithHTTPResponse = (error: any): error is HTTPErrorResponse =>
  error && typeof error === 'object' && 'response' in error && 'data' in error.response

// This enum comes from the contract repo
export enum ContractCensusType {
  FC, /// All Farcaster users.
  CHANNEL, /// Users in a specific channel.
  FOLLOWERS, /// Users following a specific account.
  CSV, /// Users in a CSV file.
  ERC20, /// Users holding a specific ERC20 token.
  NFT, /// Users holding a specific NFT.
}

export const censusTypeToEnum = (census: CensusType) => {
  switch (census) {
    case 'channel':
      return ContractCensusType.CHANNEL
    case 'followers':
      return ContractCensusType.FOLLOWERS
    case 'erc20':
      return ContractCensusType.ERC20
    case 'nft':
      return ContractCensusType.NFT
    case 'farcaster':
    default:
      return ContractCensusType.FC
  }
}
