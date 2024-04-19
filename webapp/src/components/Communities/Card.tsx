import {Avatar, Badge, Flex, HStack, Link, LinkProps, Text} from '@chakra-ui/react'
import {Link as RouterLink} from 'react-router-dom'
import {useAuth} from "../Auth/useAuth.ts";
import {Community} from "../../queries/communities.ts";

type CommunityCardProps = LinkProps & {
  community: Community
}

export const CommunityCard = ({community}: CommunityCardProps) => {
  const {profile} = useAuth()
  const adminsFid = community.admins.map((admin) => admin.fid)
  const isAdmin = profile && adminsFid.includes(profile.fid)
  const slug = community.id
  const pfpUrl = community.logoURL
  const name = community.name

  return <Link
    as={RouterLink}
    to={slug ? `/communities/${slug}` : undefined}
    w='full'
    border='1px solid'
    borderColor='gray.200'
    p={2}
    boxShadow='sm'
    borderRadius='lg'
    bg='white'
    _hover={{boxShadow: 'none', bg: 'purple.100'}}
  >
    <HStack>
      <Avatar src={pfpUrl}/>
      <Flex mx={2} w={'full'} justifyItems={'start'} alignItems={'start'} justifyContent={'space-between'}>
        <Text fontWeight='bold'>
          {name}
        </Text>
        {isAdmin && <Badge ml='1' colorScheme='green'>Admin</Badge>}

      </Flex>
    </HStack>
  </Link>
}
