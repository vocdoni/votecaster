import { lazy } from 'react'
import { createHashRouter, RouterProvider } from 'react-router-dom'
import { Layout } from '~components/Layout'
import { SuspenseLoader } from './SuspenseLoader'

const About = lazy(() => import('~pages/About'))
const App = lazy(() => import('~pages/App'))
const CommunitiesLayout = lazy(() => import('~pages/communities/layout'))
const CommunitiesNew = lazy(() => import('~pages/communities/new'))
const AllCommunitiesList = lazy(() => import('~pages/communities'))
const MyCommunitiesList = lazy(() => import('~pages/communities/mine'))
const Community = lazy(() => import('~pages/communities/view'))
const CommunityPoll = lazy(() => import('~pages/communities/poll'))
const FarcasterAccountProtectedRoute = lazy(() => import('./FarcasterAccountProtectedRoute'))
const Leaderboards = lazy(() => import('~pages/Leaderboards'))
const Poll = lazy(() => import('~pages/Poll'))
const Profile = lazy(() => import('~pages/Profile'))
const ProtectedRoute = lazy(() => import('./ProtectedRoute'))

export const Router = () => {
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
          element: (
            <SuspenseLoader>
              <CommunitiesLayout />
            </SuspenseLoader>
          ),
          children: [
            {
              path: '/communities',
              element: (
                <SuspenseLoader>
                  <AllCommunitiesList />
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
                  path: '/communities/mine',
                  element: (
                    <SuspenseLoader>
                      <MyCommunitiesList />
                    </SuspenseLoader>
                  ),
                },
              ],
            },
          ],
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
          path: '/communities/:id/poll/:pid',
          element: (
            <SuspenseLoader>
              <CommunityPoll />
            </SuspenseLoader>
          ),
        },
        {
          path: '/profile/:id',
          element: (
            <SuspenseLoader>
              <Profile />
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
