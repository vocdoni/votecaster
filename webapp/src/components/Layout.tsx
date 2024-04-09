import { Avatar, Box, Button, Flex, Heading, HStack, Icon, IconButton, Stack, useDisclosure } from '@chakra-ui/react'
import { GiHamburgerMenu } from 'react-icons/gi'
import { IoClose } from 'react-icons/io5'
import { Link, Outlet, useLocation } from 'react-router-dom'
import { ReputationProgress } from './Auth/Reputation'
import { useAuth } from './Auth/useAuth'

const MenuButton = ({ to, children }) => {
  const location = useLocation()
  const isActive = location.pathname === to

  return (
    <Button
      as={Link}
      to={to}
      variant='ghost'
      colorScheme='blackAlpha'
      color='white'
      bgColor={isActive ? 'purple.600' : 'purple.300'}
      _hover={{ bg: 'purple.200' }}
      size='sm'
      borderRadius='md'
    >
      {children}
    </Button>
  )
}

export const Layout = () => {
  return (
    <>
      <Navbar />
      <Flex minH='100vh' flexDir='column' justifyContent='center' alignItems='center' p={{ base: 0, sm: 5, xl: 10 }}>
        <Outlet />
      </Flex>
    </>
  )
}

const links = [
  {
    name: 'App',
    to: '/',
  },
  {
    name: 'About',
    to: '/about',
  },
  {
    name: 'Leaderboards',
    to: '/leaderboards',
  },
]

export const Navbar = () => {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { isAuthenticated, profile } = useAuth()

  return (
    <Box px={{ base: 3 }}>
      <Flex h={16} alignItems={'center'} justifyContent={'space-between'}>
        <IconButton
          icon={isOpen ? <Icon as={IoClose} /> : <Icon as={GiHamburgerMenu} />}
          aria-label={'Open Menu'}
          display={{ md: 'none' }}
          onClick={isOpen ? onClose : onOpen}
        />
        <HStack spacing={8} alignItems={'center'}>
          <Heading fontSize='2xl'>farcaster.vote</Heading>
          <HStack as={'nav'} spacing={4} display={{ base: 'none', md: 'flex' }}>
            {links.map((link, key) => (
              <MenuButton key={key} to={link.to}>
                {link.name}
              </MenuButton>
            ))}
          </HStack>
        </HStack>
        <Flex alignItems={'center'}>
          {isAuthenticated && (
            <Link to='/profile'>
              <ReputationProgress mr={3} size='32px' />
              <Avatar size={'sm'} src={profile?.pfpUrl} />
            </Link>
          )}
        </Flex>
      </Flex>

      {isOpen && (
        <Box pb={4} display={{ md: 'none' }}>
          <Stack as={'nav'} spacing={4}>
            {links.map((link, key) => (
              <MenuButton key={key} to={link.to}>
                {link.name}
              </MenuButton>
            ))}
          </Stack>
        </Box>
      )}
    </Box>
  )
}
