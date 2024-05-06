import { Button, useDisclosure } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { FaUserGroup } from 'react-icons/fa6'
import { useAuth } from '~components/Auth/useAuth'
import { fetchPollsRemainingVoters } from '~queries/polls'
import { UsersTableModal } from './UsersTableModal'

export const RemainingVotersTableModal = ({ id }: { id?: string }) => {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { bfetch } = useAuth()
  const { data, error, isLoading } = useQuery({
    queryKey: ['remainingVoters', id],
    queryFn: fetchPollsRemainingVoters(bfetch, id!),
    enabled: !!id && isOpen,
  })

  if (!id) return

  return (
    <>
      <Button size='sm' onClick={onOpen} isLoading={isLoading} rightIcon={<FaUserGroup />}>
        Remaining voters
      </Button>
      <UsersTableModal
        isOpen={isOpen}
        onClose={onClose}
        error={error}
        isLoading={isLoading}
        title='Remaining voters'
        filename='remaining-voters.csv'
        data={data?.map((username) => [username])}
        downloadText='Download remaining voters list'
      />
    </>
  )
}
