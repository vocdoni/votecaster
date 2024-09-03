import { paginationItemsPerPage } from '~constants'
import { useContractPollInfo } from '~queries/polls'

export const community2CommunityForm = (data: Community) => ({
  censusType: data.censusType as CensusType,
  name: data.name,
  admins: data.admins.map((admin) => ({ label: admin.username, value: admin.fid })),
  src: data.logoURL,
  groupChat: data.groupChat,
  channel: data.censusChannel ? data.censusChannel.id : '',
  channels: data.channels ? data.channels.map((channel) => ({ label: channel, value: channel })) : [],
  enableNotifications: data.notifications,
  disabled: !data.disabled,
  addresses: data.censusAddresses || [],
})

type DataType = ReturnType<typeof useContractPollInfo>['data']

export const contractDataToObject = (data?: DataType): Partial<PollInfo> => {
  if (!data) return {}

  const date = new Date(data.date.replace(/[UTC|CEST]+ m=\+[\d.]+$/, ''))

  return {
    ...data,
    turnout: Number(data.turnout),
    finalized: true,
    endTime: date,
    lastVoteTime: date,
    createdTime: date,
    question: data.question,
    options: [...data.options],
    participants: data.participants.map((p) => parseInt(p.toString())),
    tally: data.tally.map((t) => t.map((v) => parseInt(v.toString()))),
    // see https://github.com/vocdoni/vote-frame/issues/134
    // turnout: Number(data.turnout),
    voteCount: data.participants.length,
    totalWeight: Number(data.totalVotingPower),
  }
}

export const pageToOffset = (page: number) => (page - 1) * paginationItemsPerPage
export const offsetToPage = (offset: number) => Math.floor(offset / paginationItemsPerPage) + 1

export const numberFromId = (id?: CommunityID) => {
  if (!id) return 0

  const split = id.split(':')
  if (split.length === 1) return 0

  return parseInt(split[1], 10)
}
export const chainFromId = (id?: CommunityID) => {
  if (!id) return Object.keys(import.meta.env.chains)[0] as ChainKey

  return id.split(':')[0] as ChainKey
}

export const userToProfile = (user: User): Profile => ({
  fid: user.userID ?? 0,
  username: user.username,
  displayName: user.displayName,
  bio: '',
  pfpUrl: user.avatar,
  custody: user.custodyAddress,
  verifications: user.signers,
  addresses: user.addresses,
})

export const profileToUser = (profile: Profile): User => ({
  electionCount: 0,
  castedVotes: 0,
  username: profile.username,
  displayName: profile.displayName,
  custodyAddress: profile.custody,
  addresses: profile.addresses ?? [],
  signers: profile.verifications,
  followers: 0,
  lastUpdated: new Date(),
  avatar: profile.pfpUrl,
})

export const transformDelegations = (delegations: Delegation[]): Delegated[] => {
  const delegationMap = new Map<number, Delegated>()

  delegations.forEach(({ to, toUser, fromUser }) => {
    if (!delegationMap.has(to)) {
      delegationMap.set(to, { to: toUser, list: [] })
    }
    delegationMap.get(to)!.list.push(fromUser)
  })

  return Array.from(delegationMap.values())
}
