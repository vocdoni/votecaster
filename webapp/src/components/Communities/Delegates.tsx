import {
  Alert,
  Box,
  Button,
  Flex,
  FormControl,
  FormErrorMessage,
  Heading,
  Input,
  InputGroup,
  Link,
  Spinner,
  Text,
  VStack,
} from '@chakra-ui/react'
import { QueryObserverResult, RefetchOptions } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useAuth } from '~components/Auth/useAuth'
import { appUrl } from '~constants'
import { useDelegations } from '~queries/communities'

type FormData = {
  to: string
}

type Refetch = (options?: RefetchOptions) => Promise<QueryObserverResult<Delegation[] | null, Error>>

type CommunityDelegateProps = {
  community: Community
  refetch: Refetch
}

type CommunityDelegationsProps = {
  delegations: Delegation[]
  refetch: Refetch
}

export const Delegates = ({ community }: { community: Community }) => {
  const { isAuthenticated } = useAuth()
  const { data, isLoading, error, refetch } = useDelegations(community)

  if (!isAuthenticated || !community) return null

  return (
    <VStack alignItems='start'>
      <Heading size='sm'>Delegate your voting power</Heading>
      {!data && !isLoading && <CommunityDelegate community={community} refetch={refetch} />}
      {data && <CommunityDelegations delegations={data} refetch={refetch} />}
      {error && <Alert status='error'>{error.toString()}</Alert>}
    </VStack>
  )
}

export const CommunityDelegations = ({ delegations, refetch }: CommunityDelegationsProps) => {
  const { bfetch, profile } = useAuth()
  const [delegatedUser, setDelegatedUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [delegation, setDelegation] = useState<Delegation | null>(null)

  useEffect(() => {
    if (!delegations || !profile) return

    // find our delegation
    const foundDelegation = delegations.find((d) => d.from === profile.fid)
    if (!foundDelegation) return
    setDelegation(foundDelegation)

    const fetchDelegatedUser = async () => {
      setLoading(true)
      try {
        const response = await bfetch(`${appUrl}/profile/fid/${foundDelegation.to}`)
        if (!response.ok) {
          throw new Error('Failed to fetch delegated user')
        }
        const { user } = (await response.json()) as UserProfileResponse
        setDelegatedUser(user)
        setError(null)
      } catch (err: unknown) {
        setError((err as Error).message)
      } finally {
        setLoading(false)
      }
    }

    fetchDelegatedUser()
  }, [delegations, profile])

  const revokeDelegation = async () => {
    if (!delegation) return

    if (!window.confirm('Are you sure?')) return

    setLoading(true)
    try {
      const response = await bfetch(`${appUrl}/profile/delegation/${delegation.id}`, {
        method: 'DELETE',
      })
      if (!response.ok) {
        throw new Error('Failed to revoke delegation')
      }
      setDelegatedUser(null)
      setDelegation(null)
      setError(null)
      refetch()
    } catch (err: unknown) {
      setError((err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  if (!delegations || !delegations.length || !profile || !delegatedUser) return null

  if (loading) {
    return <Spinner />
  }

  if (error) {
    return <Alert status='error'>{error}</Alert>
  }

  return (
    <Box display='flex' fontSize='sm' gap={2}>
      <Text>
        You delegated your vote to{` `}
        <Link href={`https://warpcast.com/${delegatedUser.username}`}>{delegatedUser.displayName}</Link>
      </Text>
      <Button size='xs' colorScheme='purple' variant='link' onClick={revokeDelegation}>
        Revoke
      </Button>
    </Box>
  )
}

export const CommunityDelegate = ({ community, refetch }: CommunityDelegateProps) => {
  const { bfetch } = useAuth()
  const {
    register,
    handleSubmit,
    formState: { errors },
    setError,
    clearErrors,
  } = useForm<FormData>()
  const [loading, setLoading] = useState(false)

  const onSubmit = async (data: FormData) => {
    setLoading(true)
    try {
      // Check if the user exists and retrieve their ID
      const userResponse = await bfetch(`${appUrl}/profile/user/${data.to}`)
      if (!userResponse.ok) {
        throw new Error('User not found')
      }
      const { user } = (await userResponse.json()) as UserProfileResponse

      // Perform the delegation
      const response = await bfetch(`${appUrl}/profile/delegation`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          to: user.userID,
          communityId: community.id,
        }),
      })

      if (!response.ok) {
        throw new Error('Delegation failed')
      }

      clearErrors()
      refetch()
    } catch (error: unknown) {
      setError('to', { type: 'manual', message: (error as Error).message })
      console.error('Error delegating voting power:', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Flex gap={2} as='form' onSubmit={handleSubmit(onSubmit)}>
      <FormControl isInvalid={!!errors.to}>
        <InputGroup>
          <Input
            placeholder='Type a username to delegate your power'
            size='sm'
            {...register('to', { required: 'Username is required' })}
          />
        </InputGroup>
        {errors.to && <FormErrorMessage>{errors.to.message}</FormErrorMessage>}
      </FormControl>
      <Button type='submit' colorScheme='purple' variant='outline' isLoading={loading} size='sm'>
        Set
      </Button>
    </Flex>
  )
}
