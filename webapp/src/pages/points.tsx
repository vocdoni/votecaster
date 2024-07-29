import { Card, CardBody, CardHeader, Heading, Text, VStack } from '@chakra-ui/react'
import { ReputationLeaderboard } from '~components/Reputation/Leaderboard'
import { ReputationTable } from '~components/Reputation/Reputation'
import { useReputation } from '~components/Reputation/useReputation'

const Points = () => {
  const { reputation } = useReputation()

  // this is a restricted route, reputation should be there
  if (!reputation) return null

  return (
    <VStack w='full'>
      <Card w='full'>
        <CardHeader textAlign='center' gap={3} display='flex' flexDir='column'>
          <Heading size='md'>Points Leaderboard</Heading>
          <Text fontSize='sm'>Ranking of users with the most points</Text>
        </CardHeader>
        <CardBody>
          <ReputationLeaderboard />
        </CardBody>
      </Card>
      <Card w='full'>
        <CardHeader as={Heading} size='md'>
          Your reputation points
        </CardHeader>
        <CardBody>
          <ReputationTable reputation={reputation} />
        </CardBody>
      </Card>
    </VStack>
  )
}

export default Points
