import {
  Avatar,
  Box,
  Button,
  FormControl,
  FormErrorMessage,
  Heading,
  HStack,
  IconButton,
  Input,
  Link,
  Spacer,
  StackProps,
  VStack,
} from '@chakra-ui/react'
import { useForm } from 'react-hook-form'
import { FaTrash } from 'react-icons/fa6'
import { useQuery } from 'react-query'
import { fetchMutedUsers } from '../queries/profile'
import { appUrl } from '../util/constants'
import { useAuth } from './Auth/useAuth'
import { Check } from './Check'

export const MutedUsersList: React.FC = (props: StackProps) => {
  const {
    register,
    handleSubmit,
    reset,
    setError,
    trigger,
    formState: { errors },
  } = useForm({
    defaultValues: {
      username: '',
    },
  })
  const { bfetch } = useAuth()
  const { data, error, isLoading, refetch } = useQuery<Profile[], Error>('mutedUsers', fetchMutedUsers(bfetch))

  const handleUnmute = async (username: string) => {
    try {
      await bfetch(`${appUrl}/profile/mutedUsers/${username}`, { method: 'DELETE' }).then(refetch)
    } catch (e: any) {
      console.error('could not unmute user', e)
    }
  }

  const onSubmit = async (data) => {
    try {
      await bfetch(`${appUrl}/profile/mutedUsers`, {
        method: 'POST',
        body: JSON.stringify({ username: data.username }),
      }).then(refetch)
      reset({ username: '' }) // Reset only the username field
    } catch (e: any) {
      if ('message' in e) {
        console.log('error received as message:', e.message)
        setError('username', { message: e.message })
      }
      console.error('could not add muted user', e)
    }
  }

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  return (
    <Box borderRadius='md' p={4} bg='purple.100'>
      <Heading fontSize='xl' mb={4} fontWeight='600' color='purple.800'>
        Muted users
      </Heading>
      <VStack spacing={4} align='stretch'>
        {data ? (
          data?.map((user) => (
            <HStack
              key={user.userID}
              spacing={4}
              p={4}
              bg='white'
              borderRadius='md'
              align='center'
              border='1px'
              boxShadow='sm'
              borderColor='purple.200'
            >
              <Avatar src={user.pfpUrl} name={user.username} size='sm' />
              <Link href={`https://warpcast.com/${user.username}`} isExternal fontWeight='medium' color='purple.500'>
                {user.username}
              </Link>
              <Spacer />
              <IconButton
                aria-label={`Unmute ${user.username}`}
                icon={<FaTrash />}
                onClick={() => handleUnmute(user.username)}
                colorScheme='purple'
                title={`Unmute "${user.username}"`}
                size='sm'
              />
            </HStack>
          ))
        ) : (
          <p>No muted users yet</p>
        )}
        <form onSubmit={handleSubmit(onSubmit)}>
          <Box borderRadius='md' p={4} bg='purple.50'>
            <HStack spacing={4}>
              <FormControl isInvalid={!!errors.username}>
                <Input
                  placeholder='user to be muted'
                  {...register('username', { required: 'This field is required' })}
                />
                <FormErrorMessage>{errors.username?.message?.toString()}</FormErrorMessage>
              </FormControl>
              <Button type='submit' colorScheme='purple' flexGrow={1}>
                Mute
              </Button>
            </HStack>
          </Box>
        </form>
      </VStack>
    </Box>
  )
}
