import { Alert, Box, Button, Heading, Link, Progress } from '@chakra-ui/react'
import { useEffect, useState } from 'react'
import { useAuth } from '~components/Auth/useAuth'
import { useFetchProfileMutation, useRevokeDelegation } from '~queries/profile'
import { DelegatedCommunity } from './Communities/Delegates'

type DelegationProps = {
  delegations: Delegation[]
  communityId?: CommunityID
}

export const Delegation = ({ delegations, communityId }: DelegationProps) => {
  const { profile } = useAuth()
  const [delegatedUser, setDelegatedUser] = useState<User | undefined>()
  const [delegation, setDelegation] = useState<Delegation | undefined>()
  const delegatedUserMutation = useFetchProfileMutation()
  const revokeMutation = useRevokeDelegation()

  useEffect(() => {
    if (!delegations || !profile) return

    // find our delegation
    const foundDelegation = delegations.find(
      (d) => d.from === profile.fid && (communityId ? d.communityId === communityId : true)
    )
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
        setDelegatedUser(undefined)
        setDelegation(undefined)
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

  if (!delegations || !delegations.length || !profile || !delegatedUser) return null

  return (
    <Box fontSize='sm' gap={2} display='flex' alignItems='end'>
      <Box>
        {communityId && <DelegatedCommunity delegation={delegation} mr={1} />}
        Vote delegated to{` `}
        <Link fontWeight='bold' href={`https://warpcast.com/${delegatedUser.username}`}>
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

export const Delegations = ({ delegations }: DelegationsProps) => {
  return (
    <Box display='flex' flexDir='column' gap={4} boxShadow='md' borderRadius='md' bg='purple.100' p={4}>
      <Heading fontSize='xl' mb={4} fontWeight='600' color='purple.800'>
        Delegations
      </Heading>
      {delegations ? (
        delegations.map((delegation) => (
          <Delegation key={delegation.id} delegations={delegations} communityId={delegation.communityId} />
        ))
      ) : (
        <Box>No delegations yet. You can delegate your vote via the communities view</Box>
      )}
    </Box>
  )
}
