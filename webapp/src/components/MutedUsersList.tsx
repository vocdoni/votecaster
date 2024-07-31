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
  Text,
  VStack,
} from '@chakra-ui/react'
import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { FaTrash } from 'react-icons/fa6'
import { useMuteUser, useUnmuteUser } from '~queries/profile'
import { useAuth } from './Auth/useAuth'

type MutedUsersFormValues = {
  username: string
}

type MutedUsersListProps = BoxProps & {
  list?: Profile[]
}

export const MutedUsersList: React.FC<MutedUsersListProps> = ({ list, ...props }) => {
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

  const { profile } = useAuth()
  const queryClient = useQueryClient()
  const muteUserMutation = useMuteUser()
  const unmuteUserMutation = useUnmuteUser()
  const [loading, setLoading] = useState<boolean>(false)

  const onSubmit = async (data: MutedUsersFormValues) => {
    setLoading(true)
    muteUserMutation.mutate(data.username, {
      onSuccess: () => {
        reset({ username: '' })
        queryClient.invalidateQueries({
          queryKey: ['profile', profile?.username],
        })
      },
      onError: (error) => {
        if (error instanceof Error) {
          setError('username', { message: error.message })
        }
        console.error('could not add muted user', error)
      },
      onSettled: () => {
        setLoading(false)
      },
    })
  }

  return (
    <Box borderRadius='md' p={4} bg='purple.100' {...props}>
      <Heading fontSize='xl' mb={4} fontWeight='600' color='purple.800'>
        Muted users
      </Heading>
      <VStack spacing={4} align='stretch'>
        {list ? (
          list.map((user) => (
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
                onClick={() =>
                  unmuteUserMutation.mutate(user.username, {
                    onSuccess: () => {
                      queryClient.invalidateQueries({
                        queryKey: ['profile', profile?.username],
                      })
                    },
                    onError: (error) => {
                      console.error('could not unmute user', error)
                    },
                    onSettled: () => {
                      setLoading(false)
                    },
                  })
                }
                colorScheme='purple'
                isLoading={loading}
                size='sm'
              />
            </HStack>
          ))
        ) : (
          <Text>No muted users yet</Text>
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
