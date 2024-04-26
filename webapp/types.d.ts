import { Reputation } from '~components/Auth/useAuthProvider'

declare global {
  type FetchFunction = (input: RequestInfo, init?: RequestInit) => Promise<Response>

  type Address = {
    address: string
    blockchain: string
  }

  type CensusType = 'farcaster' | 'channel' | 'followers' | 'custom' | 'erc20' | 'nft' | 'community'

  type Profile = {
    fid: number
    username: string
    displayName: string
    bio: string
    pfpUrl: string
    custody: string
    verifications: string[]
  }

  type Poll = {
    censusParticipantsCount: number
    createdByDisplayname: string
    createdByUsername: string
    createdTime: Date
    electionId: string
    endTime: Date
    finalized: boolean
    lastVoteTime: Date
    question: string
    title: string
    turnout: number
    voteCount: number
  }

  type PollResult = {
    censusRoot: string
    censusURI: string
    endTime: Date
    options: string[]
    participants: number[]
    censusParticipantsCount: number
    question: string
    tally: number[][]
    turnout: number
    voteCount: number
    finalized: boolean
  }

  type PollInfo = {
    createdTime: string
    electionId: string
    lastVoteTime: string
    endTime: string
    question: string
    voteCount: number
    censusParticipantsCount: number
    turnout: number
    createdByUsername: string
    createdByDisplayname: string
    totalWeight: string
    options: string[]
    tally: string[]
    finalized: boolean
    participants: number[]
  }

  type UserRanking = {
    fid: number
    username: string
    count: number
    displayName: string
  }

  type UserProfileResponse = {
    polls: Poll[]
    mutedUsers: Profile[]
    user: Profile
    reputation: number
    reputationData: Reputation
  }

  type Community = {
    id: number
    name: string
    logoURL: string
    admins: Profile[]
    notifications: boolean
    censusType: string
    censusAddresses: Address[]
    channels: string[]
    groupChat: string
    disabled: boolean
  }

  type HTTPErrorResponse = {
    response?: {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      data?: any
    }
  }
}

export {}
