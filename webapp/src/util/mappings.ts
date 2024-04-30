export const community2CommunityForm = (data: Community) => ({
  censusType: data.censusType as CensusType,
  name: data.name,
  admins: data.admins.map((admin) => ({ label: admin.username, value: admin.fid })),
  src: data.logoURL,
  groupChat: data.groupChat,
  channel: data.censusChannel.id,
  channels: data.channels.map((channel) => ({ label: channel, value: channel })),
  enableNotifications: data.notifications,
  disabled: data.disabled,
})
