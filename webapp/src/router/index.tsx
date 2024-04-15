import { createHashRouter } from 'react-router-dom'
import { Layout } from '../components/Layout'
import { About } from '../pages/About'
import { App } from '../pages/App'
import { Communities } from '../pages/Communities'
import { Leaderboards } from '../pages/Leaderboards'
import { Poll } from '../pages/Poll'
import { Profile } from '../pages/Profile'
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
        element: <Poll />,
      },
      {
        element: <ProtectedRoute />,
        children: [
          {
            path: '/profile',
            element: <Profile />,
          },
          {
            path: '/communities',
            element: <Communities />,
          },
        ],
      },
    ],
  },
])

export default router
