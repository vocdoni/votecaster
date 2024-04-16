import {
  Avatar,
  Box,
  BoxProps,
  Button,
  FormControl,
  FormErrorMessage,
  Heading,
  HStack,
  IconButton,
  Input,
  Link,
  Spacer,
  VStack,
} from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { FaTrash } from 'react-icons/fa6'
import { fetchMutedUsers } from '../queries/profile'
import { appUrl } from '../util/constants'
import { Profile } from '../util/types'
import { useAuth } from './Auth/useAuth'
import { Check } from './Check'

type MutedUsersFormValues = {
  username: string
}

export const MutedUsersList: React.FC = (props: BoxProps) => {
  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors },
  } = useForm<MutedUsersFormValues>({
    defaultValues: {
      username: '',
    },
  })
  const { bfetch } = useAuth()
  const [loading, setLoading] = useState<boolean>(false)
  const { data, error, isLoading, refetch } = useQuery<Profile[], Error>({
    queryKey: ['mutedUsers'],
    queryFn: fetchMutedUsers(bfetch),
  })

  const handleUnmute = async (username: string) => {
    try {
      setLoading(true)
      await bfetch(`${appUrl}/profile/mutedUsers/${username}`, { method: 'DELETE' }).then(() => refetch())
    } catch (e) {
      console.error('could not unmute user', e)
    } finally {
      setLoading(false)
    }
  }

  const onSubmit = async (data: MutedUsersFormValues) => {
    setLoading(true)
    try {
      await bfetch(`${appUrl}/profile/mutedUsers`, {
        method: 'POST',
        body: JSON.stringify({ username: data.username }),
      }).then(() => refetch())
      reset({ username: '' }) // Reset only the username field
    } catch (e) {
      if (e instanceof Error) {
        console.log('error received as message:', e.message)
        setError('username', { message: e.message })
      }
      console.error('could not add muted user', e)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Box borderRadius='md' p={4} bg='purple.100' {...props}>
      <Heading fontSize='xl' mb={4} fontWeight='600' color='purple.800'>
        Muted users
      </Heading>
      <VStack spacing={4} align='stretch'>
        {data ? (
          data?.map((user) => (
            <HStack
              key={user.fid}
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
                title={`Unmute "${user.username}"`}
                icon={<FaTrash />}
                onClick={() => handleUnmute(user.username)}
                colorScheme='purple'
                isLoading={loading}
                size='sm'
              />
            </HStack>
          ))
        ) : isLoading || error ? (
          <Check isLoading={isLoading} error={error} />
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
              <Button type='submit' colorScheme='purple' flexGrow={1} isLoading={loading}>
                Mute
              </Button>
            </HStack>
          </Box>
        </form>
      </VStack>
    </Box>
  )
}
