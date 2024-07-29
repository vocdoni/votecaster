import { Table, Tbody, Th, Thead, Tr } from '@chakra-ui/react'
import { Check } from '~components/Check'
import { useFetchPoints } from '~queries/rankings'

export const ReputationLeaderboard = () => {
  const { data, isLoading, error } = useFetchPoints()

  return (
    <>
      <Check isLoading={isLoading} error={error} />
      <Table>
        <Thead>
          <Tr>
            <Th>Position</Th>
            <Th>User</Th>
            <Th>Points</Th>
          </Tr>
        </Thead>
        <Tbody>
          {data?.map((user, index) => (
            <Tr key={index}>
              <Th>{index + 1}</Th>
              <Th>{user.username || user.communityName}</Th>
              <Th>{user.totalPoints}</Th>
            </Tr>
          ))}
        </Tbody>
      </Table>
    </>
  )
}
