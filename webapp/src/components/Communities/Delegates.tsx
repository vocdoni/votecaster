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
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useAuth } from '~components/Auth/useAuth'
import { useDelegations } from '~queries/communities'
import { useDelegateVote, useFetchProfileMutation, useRevokeDelegation } from '~queries/profile'

type FormData = {
  to: string
}

type CommunityDelegateProps = {
  community: Community
}

export const Delegates = ({ community }: { community: Community }) => {
  const { isAuthenticated } = useAuth()
  const { data, isLoading, error } = useDelegations(community)

  if (!isAuthenticated || !community) return null

  return (
    <VStack alignItems='start' maxW={{ base: 'full', lg: '50%' }}>
      <Heading size='sm'>Delegate your vote</Heading>
      <Text fontSize='small' fontStyle='italic'>
        You can delegate your voting power to any community member to vote on your behalf. Revoke the delegation at any
        time, though this won't affect votes already in progress.
      </Text>
      {!data && !isLoading && <CommunityDelegate community={community} />}
      {data && <CommunityDelegations delegations={data} />}
      {error && <Alert status='error'>{error.toString()}</Alert>}
    </VStack>
  )
}

export const CommunityDelegate = ({ community }: CommunityDelegateProps) => {
  const {
    register,
    handleSubmit,
    formState: { errors },
    clearErrors,
  } = useForm<FormData>()
  const { mutate, status, error } = useDelegateVote()

  const onSubmit = (data: FormData) => {
    mutate(
      { to: data.to, communityId: community.id },
      {
        onSuccess: () => {
          clearErrors()
        },
      }
    )
  }

  return (
    <Flex gap={2} as='form' onSubmit={handleSubmit(onSubmit)}>
      <FormControl isInvalid={!!errors.to || status === 'error'}>
        <InputGroup>
          <Input placeholder='Farcaster username' size='sm' {...register('to', { required: 'Username is required' })} />
        </InputGroup>
        {errors.to && <FormErrorMessage>{errors.to.message}</FormErrorMessage>}
        {status === 'error' && <FormErrorMessage>{(error as Error).message}</FormErrorMessage>}
      </FormControl>
      <Button type='submit' colorScheme='purple' variant='outline' isLoading={status === 'pending'} size='sm'>
        Delegate
      </Button>
    </Flex>
  )
}

type CommunityDelegationsProps = {
  delegations: Delegation[]
}

export const CommunityDelegations = ({ delegations }: CommunityDelegationsProps) => {
  const { profile } = useAuth()
  const [delegatedUser, setDelegatedUser] = useState<User | null>(null)
  const [delegation, setDelegation] = useState<Delegation | null>(null)
  const delegatedUserMutation = useFetchProfileMutation()
  const revokeMutation = useRevokeDelegation()

  useEffect(() => {
    if (!delegations || !profile) return

    // find our delegation
    const foundDelegation = delegations.find((d) => d.from === profile.fid)
    if (!foundDelegation) return
    setDelegation(foundDelegation)

    delegatedUserMutation.mutate(foundDelegation.to, {
      onSuccess: (user) => {
        setDelegatedUser(user)
      },
    })
  }, [delegations, profile])

  const revokeDelegation = () => {
    if (!delegation) return

    if (!window.confirm("Are you sure? Remember this won't affect votes already in progress")) return

    revokeMutation.mutate(delegation.id, {
      onSuccess: () => {
        setDelegatedUser(null)
        setDelegation(null)
      },
    })
  }

  if (delegatedUserMutation.status === 'pending' || revokeMutation.status === 'pending') {
    return <Spinner />
  }

  if (delegatedUserMutation.status === 'error') {
    return <Alert status='error'>{(delegatedUserMutation.error as Error).message}</Alert>
  }

  if (revokeMutation.status === 'error') {
    return <Alert status='error'>{(revokeMutation.error as Error).message}</Alert>
  }

  if (!delegations || !delegations.length || !profile || !delegatedUser) return null

  return (
    <Box display='flex' fontSize='sm' gap={2}>
      <Text>
        You delegated your vote to{` `}
        <Link fontWeight='bold' href={`https://warpcast.com/${delegatedUser.username}`}>
          {delegatedUser.displayName}
        </Link>
      </Text>
      <Button size='xs' colorScheme='purple' variant='link' onClick={revokeDelegation}>
        Revoke
      </Button>
    </Box>
  )
}
