export type FetchFunction = (input: RequestInfo, init?: RequestInit) => Promise<Response>

export type Address = {
  address: string
  blockchain: string
}

export type Profile = {
  fid: number
  username: string
  displayName: string
  bio: string
  pfpUrl: string
  custody: string
  verifications: string[]
}

export type Poll = {
  electionId: string
  title: string
  createdByUsername: string
  createdByDisplayname: string
  voteCount: number
  createdTime: Date
  lastVoteTime: Date
}

export interface HTTPErrorResponse {
  response?: {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    data?: any
  }
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const isErrorWithHTTPResponse = (error: any): error is HTTPErrorResponse =>
  error && typeof error === 'object' && 'response' in error && 'data' in error.response
