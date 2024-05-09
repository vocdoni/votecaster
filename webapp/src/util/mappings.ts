import { IResult } from '~typechain/src/CommunityHub'

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

export const contractDataToObject = (data?: IResult.ResultStructOutput): Partial<PollInfo> => {
  if (!data) return {}

  const date = new Date(data.date.replace(/[UTC|CEST]+ m=\+[\d.]+$/, ''))

  return {
    ...data,
    finalized: true,
    endTime: date,
    lastVoteTime: date,
    createdTime: date,
    censusParticipantsCount: 0,
    question: data.question,
    options: data.options,
    participants: data.participants.map((p) => parseInt(p.toString())),
    tally: data.tally.map((t) => t.map((v) => parseInt(v.toString()))),
    turnout: Number(data.turnout),
    voteCount: data.participants.length,
    totalWeight: data.participants.reduce((acc, p) => acc + parseInt(p.toString()), 0),
  }
}
