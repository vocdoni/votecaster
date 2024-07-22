import { useFetchPoints } from '~queries/rankings'

export const ReputationLeaderboard = () => {
  const { data, isLoading } = useFetchPoints()

  console.log('data:', data, isLoading)
  return null
}
