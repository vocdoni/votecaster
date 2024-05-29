import { ReputationResponse } from '~components/Auth/useAuthProvider'

declare global {
  type FetchFunction = (input: RequestInfo, init?: RequestInit) => Promise<Response>

  type Address = {
    address: string
    blockchain: string
  }

  type CensusType = 'farcaster' | 'channel' | 'followers' | 'custom' | 'erc20' | 'nft' | 'community' | 'alfafrens'

  type Census = {
    censusId: string
    root: string
    electionId: string
    participants: { [key: string]: string }
    fromTotalAddresses: number
    createdBy: number
    totalWeight: number
    url: string
  }

  type Pagination = {
    limit: number
    offset: number
    total: number
  }

  type Profile = {
    fid: number
    username: string
    displayName: string
    bio: string
    pfpUrl: string
    custody: string
    verifications: string[]
    addresses?: string[]
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

  type PollResponse = {
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
    participants: number[]
    finalized: boolean
  }

  type PollInfo = {
    createdTime: Date
    electionId: string
    lastVoteTime: Date
    endTime: Date
    question: string
    voteCount: number
    censusParticipantsCount: number
    turnout: number
    createdByFID: number
    createdByUsername: string
    createdByDisplayname: string
    totalWeight: number
    options: string[]
    tally: number[][]
    participants: number[]
    finalized: boolean
    community?: Pick<Community, 'id' | 'name'>
  }

  type PollReminders = {
    remindableVoters: Profile[]
    alreadySent: number
    maxReminders: number
    votersWeight: [username: string, weight: string]
  }

  type PollReminderQueue = {
    queueId: string
  }

  type PollReminderStatus = {
    completed: boolean
    electionId: string
    alreadySent: number
    total: number
    fails: [username: string, error: string][]
  }

  type PollRanking = {
    electionId: string
    title: string
    voteCount: number
    createdByFID: number
    createdByUsername: string
    createdByDisplayname: string
  }

  type UserRanking = {
    fid: number
    username: string
    count: number
    displayName: string
  }

  type UserProfileResponse = ReputationResponse & {
    polls: Poll[]
    mutedUsers: Profile[]
    user: Profile
  }

  type Channel = {
    description: string
    followerCount: number
    id: string
    image: string
    name: string
    url: string
  }

  type Community = {
    id: number
    name: string
    logoURL: string
    admins: Profile[]
    notifications: boolean
    censusType: CensusType
    censusAddresses: Address[]
    censusChannel: Channel
    channels: string[]
    groupChat: string
    disabled: boolean
    userRef: Profile
  }

  type HTTPErrorResponse = {
    response?: {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      data?: any
    }
  }
}

export {}
