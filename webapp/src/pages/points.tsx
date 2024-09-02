import { Card, CardBody, CardHeader, Heading, Text, VStack } from '@chakra-ui/react'
import { useParams } from 'react-router-dom'
import { ReputationLeaderboard } from '~components/Reputation/Leaderboard'
import { ReputationTable } from '~components/Reputation/Reputation'

type Params = {
  username: string
}

const Points = () => {
  const { username } = useParams<Params>()

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
          Your reputation score
        </CardHeader>
        <CardBody>
          <ReputationTable username={username} />
        </CardBody>
      </Card>
    </VStack>
  )
}

export default Points
