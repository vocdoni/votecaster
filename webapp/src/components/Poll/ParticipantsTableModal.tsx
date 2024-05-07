import { Button, useDisclosure, useToast } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { FaUserGroup } from 'react-icons/fa6'
import { useAuth } from '~components/Auth/useAuth'
import { fetchCensus } from '~queries/census'
import { UsersTableModal } from './UsersTableModal'

export const ParticipantsTableModal = ({ id }: { id?: string }) => {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { bfetch } = useAuth()
  const toast = useToast()

  const { data, error, isLoading } = useQuery({
    queryKey: ['census', id],
    queryFn: fetchCensus(bfetch, id!),
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
      description: 'Failed to retrieve census participants list',
      status: 'error',
      duration: 5000,
      isClosable: true,
    })
  }, [error])

  if (!id) return

  return (
    <>
      <Button size='sm' onClick={onOpen} isLoading={isLoading} rightIcon={<FaUserGroup />} isDisabled={!!error}>
        Census
      </Button>
      <UsersTableModal
        isOpen={isOpen}
        onClose={onClose}
        downloadText='Download full census'
        error={error}
        isLoading={isLoading}
        title='Participants / census'
        filename='participants.csv'
        data={
          data?.participants &&
          Object.keys(data.participants).map((username) => [username, data.participants[username]])
        }
      />
    </>
  )
}
