import { Button } from '@chakra-ui/react'
import type { PropsWithChildren } from 'react'
import { Link, useLocation } from 'react-router-dom'

export const MenuButton = ({ to, children }: PropsWithChildren<{ to: string }>) => {
  const location = useLocation()
  const isActive = new RegExp(`^${to === '/' ? `${to}$` : to}`).test(location.pathname)

  return (
    <Button
      as={Link}
      to={to}
      variant='ghost'
      colorScheme='blackAlpha'
      color={isActive ? 'gray.600' : 'gray.500'}
      borderBottom={isActive ? '2px solid' : 'none'}
      borderColor={isActive ? 'purple.200' : 'transparent'}
      _hover={{ bg: 'purple.200' }}
      size='sm'
      borderRadius='0'
    >
      {children}
    </Button>
  )
}
