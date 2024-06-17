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
