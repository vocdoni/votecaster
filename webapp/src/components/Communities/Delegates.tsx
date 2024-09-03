import {
  Alert,
  Avatar,
  Badge,
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
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Progress,
  StackProps,
  Table,
  Tbody,
  Td,
  Text,
  Th,
  Thead,
  Tr,
  useDisclosure,
  VStack,
} from '@chakra-ui/react'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { FaHandHoldingDroplet } from 'react-icons/fa6'
import { generatePath, Link as RouterLink } from 'react-router-dom'
import { SignInButton } from '~components/Auth/SignInButton'
import { useAuth } from '~components/Auth/useAuth'
import { Delegation } from '~components/Delegations'
import { RoutePath } from '~constants'
import { useCommunity, useDelegations } from '~queries/communities'
import { useDelegateVote } from '~queries/profile'
import { transformDelegations } from '~util/mappings'

type FormData = {
  to: string
}

type CommunityDelegateProps = {
  community: Community
}

export const Delegates = ({ community, ...props }: { community: Community } & StackProps) => {
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

  if (!community) return null

  return (
    <VStack {...props}>
      <Heading size='sm'>Delegate your vote</Heading>
      <Text fontSize='small' fontStyle='italic'>
        You can delegate your voting power to any community member to vote on your behalf. You may revoke the delegation
        at any time, though this won't affect votes already in progress.
      </Text>
      {!isAuthenticated ? (
        <SignInButton size='sm' />
      ) : (
        <>
          {!delegation && !isLoading && <CommunityDelegate community={community} />}
          {delegation && <Delegation delegation={delegation} />}
          {error && <Alert status='error'>{error.toString()}</Alert>}
        </>
      )}
    </VStack>
  )
}

type DelegatedCommunityProps = LinkProps & {
  delegation?: Delegation
}

export const DelegationsModal = ({ community }: { community: Community }) => {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { data, isLoading, error } = useDelegations(community)

  if (!data) return null

  return (
    <>
      <Button onClick={onOpen} size='sm' leftIcon={<FaHandHoldingDroplet />}>
        Delegations
      </Button>

      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Delegations</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            {isLoading && <Progress size='xs' colorScheme='purple' isIndeterminate />}
            {error && <Alert status='error'>{error.toString()}</Alert>}
            {data && (
              <>
                <DelegatesTable delegates={transformDelegations(data)} />
                <Box fontSize='small' mt={2} textAlign='center'>
                  {data.length} people delegated their vote
                </Box>
              </>
            )}
          </ModalBody>

          <ModalFooter>
            <Button colorScheme='purple' variant='ghost' onClick={onClose}>
              Close
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </>
  )
}

export const DelegatesTable = ({ delegates }: { delegates: Delegated[] }) => (
  <Table fontSize='small'>
    <Thead>
      <Tr>
        <Th width='10rem' pl={0}>
          Delegate
        </Th>
        <Th pl={0}>Delegated from</Th>
      </Tr>
    </Thead>
    <Tbody>
      {delegates.map((delegate) => (
        <Tr key={delegate.to.userID}>
          <Td pl={0} display='flex' flexDir='row' gap={2}>
            <DelegateUser user={delegate.to} />
            <Badge alignSelf='center' colorScheme='purple'>
              {delegate.list.length}
            </Badge>
          </Td>
          <Td pl={0}>
            <AvatarStack users={delegate.list} />
          </Td>
        </Tr>
      ))}
    </Tbody>
  </Table>
)

export const AvatarStack = ({ users }: { users: User[] }) => (
  <Box display='flex' alignItems='center' flexWrap='wrap'>
    {users.map((user, index) => (
      <AvatarItem user={user} index={index} key={user.userID} total={users.length} />
    ))}
  </Box>
)

export const AvatarItem = ({ user, index, total }: { user: User; index: number; total: number }) => {
  const [isHovered, setIsHovered] = useState(false)

  return (
    <Link
      as={RouterLink}
      to={generatePath(RoutePath.ProfileView, { id: user.username })}
      key={user.userID}
      style={{
        transition: 'all 0.25s',
        zIndex: total - index,
        marginRight: isHovered ? '15px' : 0,
        marginLeft: isHovered ? (index !== 0 ? '5px' : -10) : '-10px',
      }}
      display='flex'
      alignItems='center'
      gap={2}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <Avatar src={user.avatar} size='xs' />
      {isHovered && (
        <Text display='flex' alignItems='center'>
          {user.displayName}
        </Text>
      )}
    </Link>
  )
}

export const DelegateUser = ({ user }: { user: User }) => (
  <Link
    as={RouterLink}
    to={generatePath(RoutePath.ProfileView, { id: user.username })}
    display='flex'
    alignItems='center'
    gap={2}
  >
    <Avatar src={user.avatar} size='xs' />
    <Text display='flex' alignItems='center'>
      {user.displayName}
    </Text>
  </Link>
)

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
