import { Avatar, Box, Flex, Heading, HStack, Icon, IconButton, Stack, useDisclosure } from '@chakra-ui/react'
import { GiHamburgerMenu } from 'react-icons/gi'
import { IoClose } from 'react-icons/io5'
import { Link } from 'react-router-dom'
import { ReputationProgress } from '../Auth/Reputation'
import { SignInButton } from '../Auth/SignInButton'
import { useAuth } from '../Auth/useAuth'
import { MenuButton } from './MenuButton'

type NavbarLink = {
  name: string
  to: string
  private?: boolean
}

const links: NavbarLink[] = [
  {
    name: 'App',
    to: '/',
  },
  {
    name: 'Communities',
    to: '/communities',
  },
  {
    name: 'Leaderboards',
    to: '/leaderboards',
  },
  {
    name: 'Profile',
    to: '/profile',
    private: true,
  },
  {
    name: 'About',
    to: '/about',
  },
]

export const Navbar = () => {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { isAuthenticated, profile, reputation } = useAuth()

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
            <NavbarMenuLinks />
          </HStack>
        </HStack>
        <Flex alignItems={'center'}>
          {isAuthenticated ? (
            <Link to='/profile'>
              <ReputationProgress mr={3} size='32px' reputation={reputation} />
              <Avatar size={'sm'} src={profile?.pfpUrl} />
            </Link>
          ) : (
            <SignInButton size='sm' />
          )}
        </Flex>
      </Flex>

      {isOpen && (
        <Box pb={4} display={{ md: 'none' }}>
          <Stack as={'nav'} spacing={4}>
            <NavbarMenuLinks />
          </Stack>
        </Box>
      )}
    </Box>
  )
}

const NavbarMenuLinks = () => {
  const { isAuthenticated } = useAuth()
  return links.map((link, key) => {
    if (link.private && !isAuthenticated) return null
    return (
      <MenuButton key={key} to={link.to}>
        {link.name}
      </MenuButton>
    )
  })
}
