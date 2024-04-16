import { createHashRouter, RouterProvider } from 'react-router-dom'
import { useAuth } from '../components/Auth/useAuth'
import { Layout } from '../components/Layout'
import { About } from '../pages/About'
import { App } from '../pages/App'
import { Communities } from '../pages/communities'
import { CommunitiesNew } from '../pages/communities/new'
import { Community } from '../pages/communities/view'
import { Leaderboards } from '../pages/Leaderboards'
import { Poll } from '../pages/Poll'
import { Profile } from '../pages/Profile'
import { appUrl } from '../util/constants'
import FarcasterAccountProtectedRoute from './FarcasterAccountProtectedRoute'
import ProtectedRoute from './ProtectedRoute'

export const Router = () => {
  const { bfetch } = useAuth()
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
          element: <Community />,
          loader: ({ params }) => bfetch(`${appUrl}/communities/${params.id}`),
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

  return <RouterProvider router={router} />
}
