import { Table, TableProps, Tbody, Td, Th, Thead, Tr } from '@chakra-ui/react'

export const UsersTable = ({ users, ...props }: { users?: string[][] } & TableProps) => {
  if (!users || !users.length) return

  // check if table has weight column
  const hasWeight = users[0].length > 1

  return (
    <Table {...props}>
      <Thead>
        <Tr>
          <Th>Username</Th>
          {hasWeight && <Th>Weight</Th>}
        </Tr>
      </Thead>
      <Tbody>
        {users.map(([username, weight]) => (
          <Tr key={username}>
            <Td>{username}</Td>
            {!!weight && <Td>{weight}</Td>}
          </Tr>
        ))}
      </Tbody>
    </Table>
  )
}
