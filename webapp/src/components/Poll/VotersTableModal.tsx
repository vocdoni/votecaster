import { Button, Tooltip, useDisclosure, useToast } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { FaVoteYea } from 'react-icons/fa'
import { useAuth } from '~components/Auth/useAuth'
import { fetchPollsVoters } from '~queries/polls'
import { UsersTableModal } from './UsersTableModal'

export const VotersTableModal = ({ poll }: { poll: PollInfo }) => {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { bfetch } = useAuth()
  const toast = useToast()
  const { data, error, isLoading } = useQuery({
    queryKey: ['voters', poll.electionId],
    queryFn: fetchPollsVoters(bfetch, poll.electionId),
    enabled: !!poll.electionId && isOpen,
    refetchOnWindowFocus: false,
    retry: (count, error: any) => {
      if (error.status !== 200) {
        return count < 1
      }
      return false
    },
  })

  useEffect(() => {
    if (!error) return

    toast({
      title: 'Error',
      description: error?.message || 'Failed to retrieve voters list',
      status: 'error',
      duration: 5000,
      isClosable: true,
    })
  }, [error])

  if (!poll || !poll.electionId) return

  return (
    <>
      <Tooltip hasArrow label={!poll.voteCount && `No voters yet`} placement='top'>
        <Button size='sm' onClick={onOpen} isLoading={isLoading} rightIcon={<FaVoteYea />} isDisabled={!poll.voteCount}>
          Voters
        </Button>
      </Tooltip>
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
