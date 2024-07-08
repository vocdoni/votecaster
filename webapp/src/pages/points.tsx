import { ReputationTable } from '~components/Auth/Reputation'
import { useAuth } from '~components/Auth/useAuth'

const Points = () => {
  const { reputation } = useAuth()

  // this is a restricted route, reputation should be there
  if (!reputation) return null

  return <ReputationTable reputation={reputation} />
}

export default Points
