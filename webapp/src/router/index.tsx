import { lazy } from 'react'
import { createHashRouter, RouterProvider } from 'react-router-dom'
import { useAuth } from '../components/Auth/useAuth'
import { Layout } from '../components/Layout'
import { SuspenseLoader } from './SuspenseLoader'

const About = lazy(() => import('../pages/About'))
const App = lazy(() => import('../pages/App'))
const Communities = lazy(() => import('../pages/communities'))
const CommunitiesNew = lazy(() => import('../pages/communities/new'))
const Leaderboards = lazy(() => import('../pages/Leaderboards'))
const Poll = lazy(() => import('../pages/Poll'))
const Profile = lazy(() => import('../pages/Profile'))

const FarcasterAccountProtectedRoute = lazy(() => import('./FarcasterAccountProtectedRoute'))
const ProtectedRoute = lazy(() => import('./ProtectedRoute'))
const Community = lazy(() => import('../pages/communities/view'))

export const Router = () => {
  const { bfetch } = useAuth()
  const router = createHashRouter([
    {
      path: '/',
      element: <Layout />,
      children: [
        {
          path: '/',
          element: (
            <SuspenseLoader>
              <App />
            </SuspenseLoader>
          ),
        },
        {
          path: '/about',
          element: (
            <SuspenseLoader>
              <About />
            </SuspenseLoader>
          ),
        },
        {
          path: '/leaderboards',
          element: (
            <SuspenseLoader>
              <Leaderboards />
            </SuspenseLoader>
          ),
        },
        {
          path: '/poll/:pid',
          element: (
            <SuspenseLoader>
              <Poll />
            </SuspenseLoader>
          ),
        },
        {
          path: '/communities',
          element: (
            <SuspenseLoader>
              <Communities />
            </SuspenseLoader>
          ),
        },
        {
          path: '/communities/:id',
          element: (
            <SuspenseLoader>
              <Community />
            </SuspenseLoader>
          ),
        },
        {
          element: (
            <SuspenseLoader>
              <ProtectedRoute />
            </SuspenseLoader>
          ),
          children: [
            {
              path: '/profile',
              element: (
                <SuspenseLoader>
                  <Profile />
                </SuspenseLoader>
              ),
            },
          ],
        },
        {
          element: (
            <SuspenseLoader>
              <FarcasterAccountProtectedRoute />
            </SuspenseLoader>
          ),
          children: [
            {
              path: '/communities/new',
              element: (
                <SuspenseLoader>
                  <CommunitiesNew />
                </SuspenseLoader>
              ),
            },
          ],
        },
      ],
    },
  ])

  return <RouterProvider router={router} />
}
