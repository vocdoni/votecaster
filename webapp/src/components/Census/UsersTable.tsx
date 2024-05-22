import { Checkbox, Table, TableProps, Tbody, Td, Th, Thead, Tr } from '@chakra-ui/react'
import { useState } from 'react';

interface UsersTableProps extends TableProps {
  users?: string[][];
  selectable?: boolean;
  hasWeight?: boolean;
  onSelectionChange?: (selected: string[][]) => void;
}

export const UsersTable = ({ users, selectable, onSelectionChange, hasWeight, ...props }: UsersTableProps) => {
  const [selectedUsers, setSelectedUsers] = useState<string[][]>([]);

  if (!users || !users.length) return

  // check if table has weight column only if has the hasWeight prop is not defined
  if (hasWeight === undefined) {
    hasWeight = users[0].length > 1
  }

  const handleCheckboxChange = (username: string, weight: string, isChecked: boolean) => {
    const data = [username]
    if (weight) {
      data.push(weight)
    }
    const updatedSelectedUsers = isChecked
      ? [...selectedUsers, data]
      : selectedUsers.filter(user => user[0] !== username);

    setSelectedUsers(updatedSelectedUsers);

    if (onSelectionChange) {
      onSelectionChange(updatedSelectedUsers);
    }
  };

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
            <Td>
              {selectable && (
                <Checkbox
                  pr={3}
                  onChange={(e) => handleCheckboxChange(username, weight, e.target.checked)}
                />
              )}
              {username}
            </Td>
            {hasWeight && !!weight  && <Td>{weight}</Td>}
          </Tr>
        ))}
      </Tbody>
    </Table>
  )
}
