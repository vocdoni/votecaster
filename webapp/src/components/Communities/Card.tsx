import { Avatar, Badge, Box, Flex, HStack, Link, LinkProps, Text, VStack } from '@chakra-ui/react'
import { Link as RouterLink } from 'react-router-dom'
import { useAccount } from 'wagmi'
import { useAuth } from '~components/Auth/useAuth'
import { getChain } from '~util/chain'
import { chainFromId, numberFromId } from '~util/mappings'

type CommunityCardProps = LinkProps & {
  name: string
  id?: CommunityID
  pfpUrl: string
  admins?: Profile[]
  disabled?: boolean
}
export const CommunityCard = ({ name, id, pfpUrl, admins, disabled, ...props }: CommunityCardProps) => {
  const { profile } = useAuth()
  const adminsFid = admins?.map((admin) => admin.fid) ?? []
  const isAdmin = profile && adminsFid.includes(profile.fid)
  const alias = chainFromId(id)
  const nid = numberFromId(id)

  return (
    <Link
      as={RouterLink}
      to={id ? `/communities/${alias}/${nid}` : undefined}
      w='full'
      border='1px solid'
      borderColor='gray.200'
      p={2}
      boxShadow='sm'
      borderRadius='lg'
      bg='white'
      _hover={{ boxShadow: 'none', bg: 'purple.100' }}
      color={disabled ? 'gray.400' : 'black'}
      {...props}
    >
      <HStack>
        <CommunityLogo disabled={disabled} pfpUrl={pfpUrl} alias={alias} />
        <Flex
          mx={2}
          w={'full'}
          justifyItems={'start'}
          alignItems={'center'}
          justifyContent={'space-between'}
          flexWrap={'wrap'}
        >
          <Text fontWeight='bold' noOfLines={1} wordBreak='break-all'>
            {name}
          </Text>

          <VStack>
            {isAdmin && <Badge colorScheme='green'>Admin</Badge>}
            {disabled && <Text fontSize={'xs'}>Disabled</Text>}
          </VStack>
        </Flex>
      </HStack>
    </Link>
  )
}

type CommunityLogoProps = Omit<CommunityCardProps, 'name' | 'admins' | 'id'> & {
  alias: ChainKey
}

export const CommunityLogo = ({ alias, pfpUrl, disabled }: CommunityLogoProps) => {
  const chain = getChain(alias)

  return (
    <Box pos='relative'>
      <Avatar src={pfpUrl} filter={disabled ? 'grayscale(1)' : ''} />
      <Avatar pos='absolute' bottom={0} right={0} src={chain.logo} width={4} height={4} />
    </Box>
  )
}
