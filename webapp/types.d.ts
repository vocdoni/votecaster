import { Chain } from 'viem'
import { Reputation } from '~components/Reputation/ReputationContext'
import chainsDefinition from '../chains_config.json'

declare global {
  type FetchFunction = (input: RequestInfo, init?: RequestInit) => Promise<Response>

  type Address = {
    address: string
    blockchain: string
  }

  type CensusType = 'farcaster' | 'channel' | 'followers' | 'erc20' | 'nft' | 'community' | 'alfafrens'

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

  type CensusResponse = {
    root: string
    size: number
    uri: string
  }

  type CensusResponseWithUsernames = CensusResponse & {
    usernames: string[]
    fromTotalAddresses: number
  }

  type CID = {
    censusId: string
  }

  type ConfChain = Chain & {
    logo: string
  }

  type ChainsFile = typeof chainsDefinition
  type ChainKey = keyof ChainsFile
  type ChainsConfig = Partial<{ [K in ChainKey]: ConfChain }>

  type CommunityPollParams = {
    chain: ChainKey
    community: string
    poll: string
  }

  type CommunityID = `${ChainKey}:${number}`

  type Delegation = {
    id: string
    from: number
    to: number
    communityId: CommunityID
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

  type PointsLeaderboard = {
    communityCreator: number
    communityID: string
    communityName: string
    totalPoints: number
    userDisplayname: string
    userID: number
    username: string
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
    community?: Community
  }

  type PollRequest = {
    profile: Profile
    question: string
    duration: number
    options: string[]
    notifyUsers: boolean
    notificationText?: string
    census?: CensusResponse
    community?: CommunityID
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

  type PollInfo = Omit<PollResponse, 'createdTime' | 'lastVoteTime' | 'endTime' | 'tally' | 'totalWeight'> & {
    createdTime: Date
    lastVoteTime: Date
    endTime: Date
    tally: number[][]
    totalWeight: number
    community?: Pick<Community, 'id' | 'name'>
  }

  type PollReminders = {
    remindableVoters: Profile[]
    alreadySent: number
    maxReminders: number
    votersWeight: { [key: string]: string }
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
    community?: Community
  }

  type RecursivePartial<T> = {
    [P in keyof T]?: T[P] extends (infer U)[]
      ? RecursivePartial<U>[]
      : T[P] extends object | undefined
        ? RecursivePartial<T[P]>
        : T[P]
  }

  type User = {
    userID?: number
    electionCount: number
    castedVotes: number
    username: string
    displayName: string
    custodyAddress: string
    addresses: string[]
    signers: string[]
    followers: number
    lastUpdated: Date
    avatar: string
  }

  type RecursivePartial<T> = {
    [P in keyof T]?: T[P] extends (infer U)[]
      ? RecursivePartial<U>[]
      : T[P] extends object | undefined
        ? RecursivePartial<T[P]>
        : T[P]
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
    user: User
    reputation?: Reputation
    delegations?: Delegation[]
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
    id: CommunityID
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
    ready: boolean
  }

  type CommunityCreate = Omit<Community, 'userRef'>

  type HTTPErrorResponse = {
    response?: {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      data?: any
    }
  }
}

export {}
