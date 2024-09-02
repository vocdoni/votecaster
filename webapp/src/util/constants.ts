export const appUrl = import.meta.env.APP_URL || ''

export const vocdoniEnvironment = import.meta.env.VOCDONI_ENVIRONMENT

export const vocdoniExplorer = import.meta.env.VOCDONI_EXPLORER

export const paginationItemsPerPage = 12

export const explorers = {
  degen: 'https://explorer.degen.tips',
  base: 'https://basescan.org',
  ethereum: 'https://etherscan.io',
}

export enum RoutePath {
  Base = '/',
  About = '/about',
  CommunitiesForm = '/communities/new',
  CommunitiesPaginatedList = '/communities/page?/:page?',
  Community = '/communities/:chain/:id',
  CommunityOld = '/communities/:id',
  CommunityOldPoll = '/communities/:id/poll/:pid',
  CommunityPoll = '/communities/:chain/:community/poll/:poll',
  Composer = '/composer',
  Leaderboards = '/leaderboards',
  MyCommunitiesPaginatedList = '/communities/mine/:page?',
  Points = '/points/:username?',
  Poll = '/poll/:pid',
  PollForm = '/form/:id?',
  Profile = '/profile',
  ProfileView = '/profile/:id',
}

export const degenNameResolverContractAddress = '0x4087fb91A1fBdef05761C02714335D232a2Bf3a1'

export const adminFID = import.meta.env.VOCDONI_ADMINFID

export const pollQuestionSuggestions: string[] = [
  "What's your favorite hobby?",
  'Which cuisine do you prefer?',
  'What type of music do you enjoy the most?',
  'How do you spend your weekends?',
  "What's your preferred mode of transportation?",
  'Which season do you like the best?',
  "What's your favorite movie genre?",
  'How often do you exercise?',
  'Which social media platform do you use the most?',
  'What kind of books do you like to read?',
]

export const getRandomPollQuestion = (): string => {
  const randomIndex = Math.floor(Math.random() * pollQuestionSuggestions.length)
  return pollQuestionSuggestions[randomIndex]
}

export const pollOptionSuggestions: string[] = ['Yes', 'No', 'Maybe', 'Sure', 'Not Sure']

export const getRandomPollOption = (): string => {
  const randomIndex = Math.floor(Math.random() * pollOptionSuggestions.length)
  return pollOptionSuggestions[randomIndex]
}

export const Validations = {
  required: {
    value: true,
    message: 'This field is required',
  },
}
