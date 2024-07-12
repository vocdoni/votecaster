import { Button, Tooltip, useDisclosure, useToast } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { TbUserQuestion } from 'react-icons/tb'
import { useAuth } from '~components/Auth/useAuth'
import { fetchPollsRemainingVoters } from '~queries/polls'
import { UsersTableModal } from './UsersTableModal'

export const RemainingVotersTableModal = ({ poll, census }: { poll: PollInfo; census: Census }) => {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { bfetch } = useAuth()
  const toast = useToast()
  const {
    data,
    error,
    isLoading: isLoadingData,
  } = useQuery({
    queryKey: ['remainingVoters', poll.electionId],
    queryFn: fetchPollsRemainingVoters(bfetch, poll.electionId),
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
      description: error?.message || 'Failed to retrieve remaining voters list',
      status: 'error',
      duration: 5000,
      isClosable: true,
    })
  }, [error])

  if (!poll || !poll.electionId) return

  return (
    <>
      <Tooltip hasArrow label={!poll.voteCount && `No voters yet; check census.`} placement='top'>
        <Button
          size='sm'
          onClick={onOpen}
          isLoading={isLoadingData}
          rightIcon={<TbUserQuestion />}
          isDisabled={!poll.voteCount}
        >
          Remaining voters
        </Button>
      </Tooltip>
      <UsersTableModal
        isOpen={isOpen}
        onClose={onClose}
        error={error}
        isLoading={isLoadingData}
        title='Remaining voters'
        filename='remaining-voters.csv'
        data={data?.map((username) => [username, census?.participants[username]]) as string[][]}
        downloadText='Download remaining voters list'
      />
    </>
  )
}
