import {Avatar, Badge, Flex, HStack, Link, LinkProps, Text} from '@chakra-ui/react'
import {Link as RouterLink} from 'react-router-dom'
import {useAuth} from "../Auth/useAuth.ts";
import {Profile} from "../../util/types.ts";

type CommunityCardProps = LinkProps & {
  name: string
  slug?: string
  pfpUrl: string
  admins?: Profile[]
}
export const CommunityCard = ({name, slug, pfpUrl, admins}: CommunityCardProps) => {
  const {profile} = useAuth()
  const adminsFid = admins?.map((admin) => admin.fid) ?? []
  const isAdmin = profile && adminsFid.includes(profile.fid)

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
