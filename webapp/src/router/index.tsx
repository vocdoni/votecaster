import { createHashRouter } from 'react-router-dom'
import { Layout } from '../components/Layout'
import { About } from '../pages/About'
import { App } from '../pages/App'
import { Communities } from '../pages/communities'
import { CommunitiesNew } from '../pages/communities/new'
import { Leaderboards } from '../pages/Leaderboards'
import { Poll } from '../pages/Poll'
import { Profile } from '../pages/Profile'
import FarcasterAccountProtectedRoute from './FarcasterAccountProtectedRoute'
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
        path: '/communities',
        element: <Communities />,
      },
      {
        path: '/communities/:id',
        element: <Communities />,
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
      {
        element: <FarcasterAccountProtectedRoute />,
        children: [
          {
            path: '/communities/new',
            element: <CommunitiesNew />,
          },
        ],
      },
    ],
  },
])

export default router
