import { ReputationTable } from '~components/Auth/Reputation'
import { useReputation } from '~components/Reputation/useReputation'

const Points = () => {
  const { reputation } = useReputation()

  // this is a restricted route, reputation should be there
  if (!reputation) return null

  return <ReputationTable reputation={reputation} />
}

export default Points
