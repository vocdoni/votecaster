import { Button, useDisclosure, useToast } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { FaVoteYea } from 'react-icons/fa'
import { useAuth } from '~components/Auth/useAuth'
import { fetchPollsVoters } from '~queries/polls'
import { UsersTableModal } from './UsersTableModal'

export const VotersTableModal = ({ id }: { id?: string }) => {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { bfetch } = useAuth()
  const toast = useToast()
  const { data, error, isLoading } = useQuery({
    queryKey: ['voters', id],
    queryFn: fetchPollsVoters(bfetch, id!),
    enabled: !!id && isOpen,
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
