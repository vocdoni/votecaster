import { Box, ChakraProvider } from '@chakra-ui/react'
import { useEffect } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'
import { composer } from '~src/themes/composer'

export const ComposerLayout = () => {
  const { search } = useLocation()
  const { tokenLogin, isAuthenticated } = useAuth()

  // login via token (needs to be handled here because the auth provider is outside of the router context)
  useEffect(() => {
    const params = new URLSearchParams(search.replace(/^\?/, ''))
    const token = params.get('token')

    if (!token || isAuthenticated) return

    tokenLogin(token)
  }, [search])

  return (
    <ChakraProvider theme={composer}>
      <Box>
        <Outlet />
      </Box>
    </ChakraProvider>
  )
}
