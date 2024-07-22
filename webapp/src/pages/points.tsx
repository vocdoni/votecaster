import { Card, VStack } from '@chakra-ui/react'
import { ReputationLeaderboard } from '~components/Reputation/Leaderboard'
import { ReputationTable } from '~components/Reputation/Reputation'
import { useReputation } from '~components/Reputation/useReputation'

const Points = () => {
  const { reputation } = useReputation()

  // this is a restricted route, reputation should be there
  if (!reputation) return null

  return (
    <VStack>
      <Card>
        <ReputationTable reputation={reputation} />
      </Card>
      <ReputationLeaderboard />
    </VStack>
  )
}

export default Points
