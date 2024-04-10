import { createHashRouter } from 'react-router-dom'
import { Layout } from '../components/Layout'
import { About } from '../pages/About'
import { App } from '../pages/App'
import { Leaderboards } from '../pages/Leaderboards'
import { Profile } from '../pages/Profile'
import { Voters } from '../pages/Voters'
import ProtectedRoute from './ProtectedRoute'

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
      {
        element: <ProtectedRoute />,
        children: [
          {
            path: '/profile',
            element: <Profile />,
          },
        ],
      },
    ],
  },
])

export default router
