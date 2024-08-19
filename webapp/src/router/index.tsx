import { lazy } from 'react'
import { createHashRouter, generatePath, redirect, RouterProvider } from 'react-router-dom'
import { Layout } from '~components/Layout'
import { ComposerLayout } from '~components/Layout/Composer'
import { RoutePath } from '~constants'
import { SuspenseLoader } from './SuspenseLoader'

const About = lazy(() => import('~pages/About'))
const Home = lazy(() => import('~pages/Home'))
const AppForm = lazy(() => import('~pages/Form'))
const CommunitiesLayout = lazy(() => import('~pages/communities/layout'))
const CommunitiesNew = lazy(() => import('~pages/communities/new'))
const AllCommunitiesList = lazy(() => import('~pages/communities'))
const MyCommunitiesList = lazy(() => import('~pages/communities/mine'))
const Community = lazy(() => import('~pages/communities/view'))
const CommunityPoll = lazy(() => import('~pages/communities/poll'))
const Composer = lazy(() => import('~pages/composer'))
const FarcasterAccountProtectedRoute = lazy(() => import('./FarcasterAccountProtectedRoute'))
const Leaderboards = lazy(() => import('~pages/Leaderboards'))
const Points = lazy(() => import('~pages/points'))
const Poll = lazy(() => import('~pages/Poll'))
const Profile = lazy(() => import('~pages/Profile'))
const ProtectedRoute = lazy(() => import('./ProtectedRoute'))

export const Router = () => {
  const router = createHashRouter([
    {
      path: RoutePath.Base,
      element: <Layout />,
      children: [
        {
          path: RoutePath.Base,
          element: (
            <SuspenseLoader>
              <Home />
            </SuspenseLoader>
          ),
        },
        {
          path: RoutePath.PollForm,
          element: (
            <SuspenseLoader>
              <AppForm />
            </SuspenseLoader>
          ),
        },
        {
          path: RoutePath.About,
          element: (
            <SuspenseLoader>
              <About />
            </SuspenseLoader>
          ),
        },
        {
          path: RoutePath.Leaderboards,
          element: (
            <SuspenseLoader>
              <Leaderboards />
            </SuspenseLoader>
          ),
        },
        {
          path: RoutePath.Poll,
          element: (
            <SuspenseLoader>
              <Poll />
            </SuspenseLoader>
          ),
        },
        {
          path: RoutePath.CommunityOld,
          loader: ({ params: { id } }) => {
            return redirect(generatePath(RoutePath.Community, { chain: 'degen', id: id as string }))
          },
        },
        {
          path: RoutePath.CommunityOldPoll,
          loader: ({ params: { id, pid } }) => {
            return redirect(
              generatePath(RoutePath.CommunityPoll, { chain: 'degen', community: id as string, poll: pid as string })
            )
          },
        },
        {
          path: RoutePath.Community,
          element: (
            <SuspenseLoader>
              <Community />
            </SuspenseLoader>
          ),
        },
        {
          path: RoutePath.CommunityPoll,
          element: (
            <SuspenseLoader>
              <CommunityPoll />
            </SuspenseLoader>
          ),
        },
        {
          path: RoutePath.ProfileView,
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
              path: RoutePath.Profile,
              element: (
                <SuspenseLoader>
                  <Profile />
                </SuspenseLoader>
              ),
            },
            {
              path: RoutePath.Points,
              element: (
                <SuspenseLoader>
                  <Points />
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
              path: RoutePath.CommunitiesForm,
              element: (
                <SuspenseLoader>
                  <CommunitiesNew />
                </SuspenseLoader>
              ),
            },
          ],
        },
        {
          element: (
            <SuspenseLoader>
              <CommunitiesLayout />
            </SuspenseLoader>
          ),
          children: [
            {
              path: RoutePath.CommunitiesPaginatedList,
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
                  path: RoutePath.MyCommunitiesPaginatedList,
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
      ],
    },
    {
      path: '/composer',
      element: <ComposerLayout />,
      children: [
        {
          path: '',
          element: (
            <SuspenseLoader>
              <Composer />
            </SuspenseLoader>
          ),
        },
      ],
    },
  ])

  return <RouterProvider router={router} />
}
