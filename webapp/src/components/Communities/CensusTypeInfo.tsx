import { HStack, Icon, Link, Text } from '@chakra-ui/react'
import { Fragment } from 'react'
import { FaUserGroup } from 'react-icons/fa6'
import { explorers } from '~constants'
import { shortHex } from '~util/strings'

export const CensusTypeInfo = ({ community }: { community: Community }) => {
  return (
    <HStack>
      <Icon as={FaUserGroup} />
      <Text>
        Census {community.censusType}: <CensusTypeDetail community={community} />
      </Text>
    </HStack>
  )
}

const CensusTypeDetail = ({ community }: { community: Community }) => {
  switch (community.censusType) {
    case 'erc20':
    case 'nft':
      return community.censusAddresses.map(({ address, blockchain }, index) => (
        <Fragment key={index}>
          <Link
            isExternal
            href={`${(explorers as { [key: string]: string })[blockchain]}/address/${address}`}
            key={index}
            variant='primary'
          >
            {shortHex(address)}
          </Link>
          {index < community.censusAddresses.length - 2 && ', '}
          {index === community.censusAddresses.length - 2 && ' & '}
        </Fragment>
      ))
    case 'channel':
      return (
        <Link href={community.censusChannel.url} isExternal variant='primary'>
          {community.censusChannel.name}
        </Link>
      )
    case 'followers':
      return (
        <Link isExternal href={`https://warpcast.com/${community.userRef.username}`} variant='primary'>
          {community.userRef.displayName}
        </Link>
      )
    default:
      return null
  }
}
