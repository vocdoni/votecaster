import { createContext } from 'react'

const baseRep = {
  boosters: {
    hasVotecasterNFTPass: false,
    hasVotecasterLaunchNFT: false,
    isVotecasterAlphafrensFollower: false,
    isVotecasterFarcasterFollower: false,
    isVocdoniFarcasterFollower: false,
    votecasterAnnouncementRecasted: false,
    hasKIWI: false,
    hasDegenDAONFT: false,
    hasHaberdasheryNFT: false,
    has10kDegenAtLeast: false,
    hasTokyoDAONFT: false,
    has5ProxyAtLeast: false,
    hasProxyStudioNFT: false,
    hasNameDegen: false,
    hasFarcasterOGNFT: false,
    hasMoxiePass: false,
  },
  boostersInfo: {
    degenAtLeast10kPuntuaction: 0,
    degenDAONFTPuntuaction: 0,
    farcasterOGNFTPuntuaction: 0,
    haberdasheryNFTPuntuaction: 0,
    kiwiPuntuaction: 0,
    moxiePassPuntuaction: 0,
    nameDegenPuntuaction: 0,
    proxyAtLeast5Puntuaction: 0,
    proxyStudioNFTPuntuaction: 0,
    tokyoDAONFTPuntuaction: 0,
    vocdoniFarcasterFollowerPuntuaction: 0,
    votecasterAlphafrensFollowerPuntuaction: 0,
    votecasterAnnouncementRecastedPuntuaction: 0,
    votecasterFarcasterFollowerPuntuaction: 0,
    votecasterLaunchNFTPuntuaction: 0,
    votecasterNFTPassPuntuaction: 0,
  },
  activityPoints: {
    followersPoints: 0,
    createdElectionsPoints: 0,
    castVotesPoints: 0,
    participationsPoints: 0,
    communitiesPoints: 0,
  },
  activityCounts: {
    followersCount: 0,
    createdElectionsCount: 0,
    castVotesCount: 0,
    participationsCount: 0,
    communitiesCount: 0,
  },
  activityInfo: {
    maxCastedReputation: 0,
    maxCommunityReputation: 0,
    maxElectionsReputation: 0,
    maxFollowersReputation: 0,
    maxReputation: 0,
    maxVotesReputation: 0,
  },
  totalReputation: 0,
  totalPoints: 0,
}

export type Reputation = typeof baseRep

export interface ReputationContextType {
  reputation: Reputation | undefined
  status: string
  fetchReputation: () => void
}

export const ReputationContext = createContext<ReputationContextType | undefined>(undefined)
