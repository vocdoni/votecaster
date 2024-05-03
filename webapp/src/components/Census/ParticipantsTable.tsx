import { Table, TableProps, Tbody, Td, Th, Thead, Tr } from '@chakra-ui/react'

export const ParticipantsTable = ({
  participants,
  ...props
}: { participants?: { [key: string]: string } } & TableProps) => {
  if (!participants || !Object.keys(participants).length) return

  return (
    <Table {...props}>
      <Thead>
        <Tr>
          <Th>Username</Th>
          <Th>Weight</Th>
        </Tr>
      </Thead>
      <Tbody>
        {Object.entries(participants).map(([username, weight]) => (
          <Tr key={username}>
            <Td>{username}</Td>
            <Td>{weight}</Td>
          </Tr>
        ))}
      </Tbody>
    </Table>
  )
}
