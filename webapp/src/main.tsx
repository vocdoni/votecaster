import { ChakraProvider } from '@chakra-ui/react'
import { AuthKitProvider } from '@farcaster/auth-kit'
import React from 'react'
import ReactDOM from 'react-dom/client'
import { createHashRouter, RouterProvider } from 'react-router-dom'
import { App } from './components/App'
import { Layout } from './components/Layout'
import { Voters } from './components/Voters'
import { theme } from './theme'

const router = createHashRouter([
  {
    path: '/',
    element: <Layout />,
    children: [
      {
        path: '/',
        element: <App />,
      },
      {
        path: '/poll/:pid',
        element: <Voters />,
      },
    ],
  },
])

const rootElement = document.getElementById('root')
ReactDOM.createRoot(rootElement).render(
  <React.StrictMode>
    <ChakraProvider theme={theme}>
      <AuthKitProvider>
        <RouterProvider router={router} />
      </AuthKitProvider>
    </ChakraProvider>
  </React.StrictMode>
)
