import { Checkbox, Box, Input, Table, TableProps, Tbody, Td, Th, Thead, Tr, Button, HStack } from '@chakra-ui/react'
import { useState } from 'react';

interface UsersTableProps extends TableProps {
  users?: string[][];
  selectable?: boolean;
  hasWeight?: boolean;
  findable?: boolean;
  onSelectionChange?: (selected: string[][]) => void;
}

export const UsersTable = ({ users, selectable, onSelectionChange, hasWeight, findable, ...props }: UsersTableProps) => {
  const [selectedUsers, setSelectedUsers] = useState<string[][]>([]);
  const [filterText, setFilterText] = useState('');
  const [selectedAll, setSelectedAll] = useState(false);

  if (!users || !users.length) return

  // check if table has weight column only if has the hasWeight prop is not defined
  if (hasWeight === undefined) {
    hasWeight = users[0].length > 1
  }

  const filteredUsers = users
    .filter(([username]) => username.toLowerCase().includes(filterText.toLowerCase()))
    .sort((a, b) => hasWeight ? (BigInt(a[1] || '0') < BigInt(b[1] || '0') ? 1 : -1) : 0);

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

  const isSelected = (username: string) => {
    return selectedUsers.some(user => user[0] === username);
  };

  const selectAll = () => {
    if (selectedAll) {
      setSelectedUsers([]);
      onSelectionChange && onSelectionChange([]);
      setSelectedAll(false);
      return
    }
    setSelectedUsers(filteredUsers);
    onSelectionChange && onSelectionChange(filteredUsers);
    setSelectedAll(true);
  }

  return (
    <Box>
        {!!findable && <Box px={2}>
          <HStack my={4} justifyItems={'center'} alignItems={'center'} align={'center'} alignContent={'center'}>
            <Button size={'xs'} px='4' onClick={selectAll}>{selectedAll ? 'Clear' : 'SelectAll'}</Button>
            <Input
              size={'xs'}
              p={2}
              rounded={'md'}
              placeholder="Filter by username"
              value={filterText}
              onChange={(e) => setFilterText(e.target.value)}
            />
          </HStack>
        </Box>}
      <Table {...props}>
        <Thead>
          <Tr>
            <Th>Username</Th>
            {hasWeight && <Th>Weight</Th>}
          </Tr>
        </Thead>
        <Tbody>
          {filteredUsers.map(([username, weight]) => (
            <Tr key={username}>
              <Td>
                {selectable && (
                  <Checkbox
                    pr={3}
                    isChecked={isSelected(username)}
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
    </Box>
  )
}
