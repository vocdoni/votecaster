import {
  Avatar,
  Box,
  Button,
  MenuButton as CMenuButton,
  Flex,
  Heading,
  HStack,
  Icon,
  IconButton,
  Menu,
  MenuItem,
  MenuList,
  Stack,
  useDisclosure,
} from '@chakra-ui/react'
import { GiHamburgerMenu } from 'react-icons/gi'
import { IoChevronDownCircleOutline, IoChevronUpCircleOutline, IoClose } from 'react-icons/io5'
import { Link } from 'react-router-dom'
import { useReputation } from '~components/Reputation/useReputation'
import { SignInButton } from '../Auth/SignInButton'
import { useAuth } from '../Auth/useAuth'
import { ReputationProgress } from '../Reputation/Reputation'
import { MenuButton } from './MenuButton'
import logo from '/logo-farcastervote.png'

type NavbarLink = {
  name: string
  to: string
  private?: boolean
}

const links: NavbarLink[] = [
  {
    name: 'Home',
    to: '/',
  },
  {
    name: 'Create poll',
    to: '/form',
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

  return (
    <Box px={{ base: 3 }}>
      <Flex h={16} alignItems={'center'} justifyContent={'space-between'}>
        <IconButton
          icon={isOpen ? <Icon as={IoClose} /> : <Icon as={GiHamburgerMenu} />}
          aria-label={'Open Menu'}
          display={{ lg: 'none' }}
          onClick={isOpen ? onClose : onOpen}
        />
        <HStack spacing={8} alignItems={'center'}>
          <Heading fontSize='2xl'>
            <Link to='/'>
              <Avatar src={logo} aria-label='votecaster logo' size='sm' verticalAlign='middle' /> Votecaster
            </Link>
          </Heading>
          <HStack as={'nav'} spacing={4} display={{ base: 'none', lg: 'flex' }}>
            <NavbarMenuLinks />
          </HStack>
        </HStack>
        <ProfileMenu />
      </Flex>

      {isOpen && (
        <Box pb={4} display={{ lg: 'none' }}>
          <Stack as={'nav'} spacing={4}>
            <NavbarMenuLinks />
          </Stack>
        </Box>
      )}
    </Box>
  )
}

const ProfileMenu = () => {
  const { isAuthenticated, logout, profile } = useAuth()
  const { reputation } = useReputation()

  return (
    <Flex alignItems={'center'}>
      {isAuthenticated ? (
        <Menu size='xs'>
          {({ isOpen }) => (
            <>
              <CMenuButton
                as={Button}
                size='lg'
                variant='link'
                rightIcon={isOpen ? <IoChevronUpCircleOutline /> : <IoChevronDownCircleOutline />}
              >
                <ReputationProgress mr={3} size='32px' reputation={reputation} />
                <Avatar size={'sm'} src={profile?.pfpUrl} />
              </CMenuButton>
              <MenuList>
                <MenuItem as={Link} to='/profile'>
                  Profile
                </MenuItem>
                <MenuItem as={Link} onClick={logout}>
                  Logout
                </MenuItem>
              </MenuList>
            </>
          )}
        </Menu>
      ) : (
        <SignInButton size='sm' />
      )}
    </Flex>
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
