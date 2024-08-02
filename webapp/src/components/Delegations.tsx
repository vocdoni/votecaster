import { Alert, Box, Button, Heading, Link, Progress } from '@chakra-ui/react'
import { useEffect, useState } from 'react'
import { Link as RouteLink } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'
import { useFetchProfileMutation, useRevokeDelegation } from '~queries/profile'
import { DelegatedCommunity } from './Communities/Delegates'
import { PurpleBox } from './Layout/PurpleBox'

type DelegationProps = {
  delegation: Delegation
  communityId?: CommunityID
}

export const Delegation = ({ delegation, communityId }: DelegationProps) => {
  const { profile } = useAuth()
  const [delegatedUser, setDelegatedUser] = useState<User | undefined>()
  const delegatedUserMutation = useFetchProfileMutation()
  const revokeMutation = useRevokeDelegation()

  useEffect(() => {
    if (!delegation) return

    delegatedUserMutation.mutate(delegation.to, {
      onSuccess: (user) => {
        setDelegatedUser(user)
      },
    })
  }, [delegation, profile])

  const revokeDelegation = () => {
    if (!delegation) return

    if (!window.confirm("Are you sure? Remember this won't affect votes already in progress")) return

    revokeMutation.mutate(delegation.id, {
      onSuccess: () => {
        setDelegatedUser(undefined)
      },
    })
  }

  if (delegatedUserMutation.status === 'pending' || revokeMutation.status === 'pending') {
    return <Progress w='full' isIndeterminate size='xs' colorScheme='purple' />
  }

  if (delegatedUserMutation.status === 'error') {
    return <Alert status='error'>{(delegatedUserMutation.error as Error).message}</Alert>
  }

  if (revokeMutation.status === 'error') {
    return <Alert status='error'>{(revokeMutation.error as Error).message}</Alert>
  }

  if (!profile || !delegatedUser) return null

  return (
    <Box fontSize='sm' gap={2} display='flex' alignItems='end'>
      <Box>
        {communityId && <DelegatedCommunity delegation={delegation} mr={1} />}
        Vote delegated to{` `}
        <Link as={RouteLink} fontWeight='bold' to={`/profile/${delegatedUser.username}`}>
          {delegatedUser.displayName}
        </Link>
      </Box>
      <Button size='sm' colorScheme='purple' variant='link' onClick={revokeDelegation}>
        Revoke
      </Button>
    </Box>
  )
}

type DelegationsProps = {
  delegations?: Delegation[]
}

export const Delegations = ({ delegations }: DelegationsProps) => (
  <PurpleBox>
    <Heading fontSize='xl' fontWeight='600' color='purple.800'>
      Delegations
    </Heading>
    {delegations ? (
      delegations.map((delegation) => (
        <Delegation key={delegation.id} delegation={delegation} communityId={delegation.communityId} />
      ))
    ) : (
      <Box>No delegations yet. You can delegate your vote via the communities view</Box>
    )}
  </PurpleBox>
)
