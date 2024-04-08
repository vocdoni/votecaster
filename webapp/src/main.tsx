import { ChakraProvider } from '@chakra-ui/react'
import React from 'react'
import ReactDOM from 'react-dom/client'
import { createHashRouter, RouterProvider } from 'react-router-dom'
import { AuthProvider } from './components/Auth/AuthContext'
import { Layout } from './components/Layout'
import { About } from './pages/About'
import { App } from './pages/App'
import { Leaderboards } from './pages/Leaderboards'
import { Voters } from './pages/Voters'
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
        path: '/about',
        element: <About />,
      },
      {
        path: '/leaderboards',
        element: <Leaderboards />,
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
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </ChakraProvider>
  </React.StrictMode>
)
