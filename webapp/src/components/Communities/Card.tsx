import { Avatar, Badge, Flex, HStack, Link, LinkProps, Text, VStack } from '@chakra-ui/react'
import { Link as RouterLink } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'

type CommunityCardProps = LinkProps & {
  name: string
  slug?: string
  pfpUrl: string
  admins?: Profile[]
  disabled?: boolean
}
export const CommunityCard = ({ name, slug, pfpUrl, admins, disabled, ...props }: CommunityCardProps) => {
  const { profile } = useAuth()
  const adminsFid = admins?.map((admin) => admin.fid) ?? []
  const isAdmin = profile && adminsFid.includes(profile.fid)

  return (
    <Link
      as={RouterLink}
      to={slug ? `/communities/${slug}` : undefined}
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
        <Avatar src={pfpUrl} filter={disabled ? 'grayscale(1)' : ''} />
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
