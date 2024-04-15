import { Avatar, HStack, Link, LinkProps, Text } from '@chakra-ui/react'
import { Link as RouterLink } from 'react-router-dom'

type CommunityCardProps = LinkProps & {
  name: string
  slug?: string
  pfpUrl: string
}

export const CommunityCard = ({ name, slug, pfpUrl }: CommunityCardProps) => (
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
  >
    <HStack>
      <Avatar src={pfpUrl} />
      <Text fontWeight='500'>{name}</Text>
    </HStack>
  </Link>
)
