import { Box, Button, Flex, Heading, HStack, Link, Text, VStack } from '@chakra-ui/react'
import { FaRegStar, FaUsers } from 'react-icons/fa6'
import { MdOutlineGroupAdd } from 'react-icons/md'
import { Outlet, Link as RouterLink, useLocation } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'

const CommunitiesLayout = () => (
  <VStack spacing={4} w='full' alignItems='start'>
    <Flex my={4} w='full' justifyContent='space-between' alignItems='center' wrap={'wrap'}>
      <Heading size='md' m={4}>
        Communities
      </Heading>
      <CommunitiesSelector />
    </Flex>
    <Outlet />
    <Box
      w='full'
      boxShadow='sm'
      borderRadius='lg'
      minHeight={300}
      display='flex'
      flexDir='column'
      alignItems='center'
      justifyContent='center'
      bg='white'
      p={10}
      textAlign='center'
      gap={4}
    >
      <Text fontSize='larger' fontWeight='500'>
        Create your own community and start managing its governance
      </Text>
      <Link as={RouterLink} to='/communities/new'>
        <Button leftIcon={<MdOutlineGroupAdd />}>Create a community</Button>
      </Link>
    </Box>
  </VStack>
)

const CommunitiesSelector = () => {
  const { pathname } = useLocation()
  const { isAuthenticated } = useAuth()

  if (!isAuthenticated) return

  return (
    <HStack m={4} align='center' gap={4}>
      <RouterLink to='/communities/mine'>
        <Button size='sm' leftIcon={<FaRegStar />} variant={/\/communities\/mine/.test(pathname) ? 'solid' : 'ghost'}>
          My communities
        </Button>
      </RouterLink>
      <RouterLink to='/communities'>
        <Button size='sm' leftIcon={<FaUsers />} variant={!/\/communities\/mine/.test(pathname) ? 'solid' : 'ghost'}>
          All communities
        </Button>
      </RouterLink>
    </HStack>
  )
}

export default CommunitiesLayout
