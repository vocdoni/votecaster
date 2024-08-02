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
  LinkProps,
  Progress,
  Text,
  VStack,
} from '@chakra-ui/react'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { Link as RouterLink } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'
import { Delegation } from '~components/Delegations'
import { useCommunity, useDelegations } from '~queries/communities'
import { useDelegateVote } from '~queries/profile'
import { getDelegationsPath } from '~util/objects'

type FormData = {
  to: string
}

type CommunityDelegateProps = {
  community: Community
}

export const Delegates = ({ community }: { community: Community }) => {
  const { isAuthenticated, profile } = useAuth()
  const { data, isLoading, error } = useDelegations(community)
  const [delegation, setDelegation] = useState<Delegation | undefined>()

  useEffect(() => {
    if (!data || !profile) return

    // find our delegation
    const foundDelegation = data.find((d) => d.from === profile.fid && d.communityId === community.id)
    if (!foundDelegation) {
      setDelegation(undefined)
      return
    }
    setDelegation(foundDelegation)
  }, [data, profile])

  if (!isAuthenticated || !community) return null

  const path = getDelegationsPath(data || [])
  if (path.length) {
    for (const p of path) {
      console.info('Delegation path:', p.join(' -> '))
    }
  }

  return (
    <VStack alignItems='start' maxW={{ base: 'full', lg: '50%' }}>
      <Heading size='sm'>Delegate your vote</Heading>
      <Text fontSize='small' fontStyle='italic'>
        You can delegate your voting power to any community member to vote on your behalf. You may revoke the delegation
        at any time, though this won't affect votes already in progress.
      </Text>
      {!delegation && !isLoading && <CommunityDelegate community={community} />}
      {delegation && <Delegation delegation={delegation} />}
      {error && <Alert status='error'>{error.toString()}</Alert>}
    </VStack>
  )
}

type DelegatedCommunityProps = LinkProps & {
  delegation?: Delegation
}

export const DelegatedCommunity = ({ delegation, ...props }: DelegatedCommunityProps) => {
  const { data, isLoading, error } = useCommunity(delegation?.communityId)

  if (isLoading) {
    return <Progress isIndeterminate colorScheme='purple' size='xs' />
  }

  if (!data) return null

  if (error) {
    return (
      <Alert status='error' size='xs'>
        {error.toString()}
      </Alert>
    )
  }

  return (
    <Box fontWeight='bold'>
      <Link as={RouterLink} to={`/communities/${data.id.replace(':', '/')}`} {...props}>
        {data.name}
      </Link>
    </Box>
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
