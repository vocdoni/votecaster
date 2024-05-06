import { Button, useDisclosure } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { FaVoteYea } from 'react-icons/fa'
import { useAuth } from '~components/Auth/useAuth'
import { fetchPollsVoters } from '~queries/polls'
import { UsersTableModal } from './UsersTableModal'

export const VotersTableModal = ({ id }: { id?: string }) => {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { bfetch } = useAuth()
  const { data, error, isLoading } = useQuery({
    queryKey: ['voters', id],
    queryFn: fetchPollsVoters(bfetch, id!),
    enabled: !!id && isOpen,
  })

  if (!id) return

  return (
    <>
      <Button size='sm' onClick={onOpen} isLoading={isLoading} rightIcon={<FaVoteYea />}>
        Voters
      </Button>
      <UsersTableModal
        isOpen={isOpen}
        onClose={onClose}
        error={error}
        isLoading={isLoading}
        title='Voters'
        filename='voters.csv'
        data={data?.map((username) => [username])}
        downloadText='Download voters list'
      />
    </>
  )
}
