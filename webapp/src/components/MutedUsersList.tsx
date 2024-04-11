import {
  Avatar,
  Box,
  Button,
  HStack,
  Icon,
  IconButton,
  Input,
  Link,
  Spacer,
  StackProps,
  Text,
  VStack,
} from '@chakra-ui/react'
import { FaSquarePlus, FaTrash } from 'react-icons/fa6'
import { useQuery } from 'react-query'
import { fetchMutedUsers } from '../queries/profile'
import { useAuth } from './Auth/useAuth'
import { Check } from './Check'

export const MutedUsersList: React.FC = (props: StackProps) => {
  const { bfetch } = useAuth()
  const { data, error, isLoading } = useQuery<Profile[], Error>('mutedUsers', fetchMutedUsers(bfetch))

  // Function to handle the unmute action
  const handleUnmute = (username: string) => {
    // Implement unmute logic here
  }

  // Function to handle adding a new muted user
  const handleAddMutedUser = (username: string) => {
    // Implement add muted user logic here
  }

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  return (
    <VStack spacing={4} align='stretch' w='full' {...props}>
      {data?.map((user) => (
        <HStack key={user.fid} spacing={4} p={4} bg='white' boxShadow='md' borderRadius='md' align='center'>
          <Avatar src={user.pfpUrl} name={user.username} />
          <Link href={`https://warpcast.com/${user.username}`} isExternal color='purple.500'>
            <Text fontWeight='medium'>{user.username}</Text>
          </Link>
          <Spacer />
          <IconButton
            aria-label={`Unmute ${user.username}`}
            icon={<Icon as={FaTrash} />}
            onClick={() => handleUnmute(user.username)}
            colorScheme='purple'
            size='sm'
          />
        </HStack>
      ))}
      <Box p={4} boxShadow='md' borderRadius='md' bg='purple.50'>
        <HStack>
          <Input placeholder='Add a user to mute' />
          <Button colorScheme='purple' leftIcon={<Icon as={FaSquarePlus} />} onClick={handleAddMutedUser}>
            Add
          </Button>
        </HStack>
      </Box>
    </VStack>
  )
}
