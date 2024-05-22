import {
  Box,
  Button,
  FormControl,
  FormErrorMessage,
  Input,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Tooltip,
  useDisclosure,
  useToast
} from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { MdSend } from "react-icons/md"
import { appUrl } from '~constants'
import { useAuth } from '~components/Auth/useAuth'
import { UsersTable } from '~components/Census/UsersTable'
import { Check } from '~components/Check'
import { fetchPollsReminders } from '~queries/polls'

type ReminderFormValues = {
  message: string
}

export const PollRemindersModal = ({ poll }: { poll: PollInfo }) => {
  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors },
  } = useForm<ReminderFormValues>({
    defaultValues: {
      message: '',
    },
  })
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { bfetch } = useAuth()
  const toast = useToast()
  const [loading, setLoading] = useState<boolean>(false)
  const { data: reminders, error, isLoading, refetch } = useQuery({
    queryKey: ['reminders', poll.electionId],
    queryFn: fetchPollsReminders(bfetch, poll.electionId),
    enabled: !!poll.electionId && isOpen,
    refetchOnWindowFocus: false,
    retry: (count, error: any) => {
      if (error.status !== 200) {
        return count < 1
      }
      return false
    },
  })

  const [selectedUsers, setSelectedUsers] = useState<Profile[]>([]);

  useEffect(() => {
    if (!error) return

    toast({
      title: 'Error',
      description: error?.message || 'Failed to retrieve poll reminders',
      status: 'error',
      duration: 5000,
      isClosable: true,
    })
  }, [error])

  if (!poll || !poll.electionId) return

  const sendReminders = async (data: ReminderFormValues) => {
    setLoading(true)

    const users = {} as { [key: string]: string }
    selectedUsers.forEach(({ fid, username }) => {
      users[fid.toString()] = username
    })
    try {
      await bfetch(`${appUrl}/poll/${poll.electionId}/reminders`, {
        method: 'POST',
        body: JSON.stringify({ 
          type: "individual",
          content: data.message,
          users: users,
        }),
      }).then(() => refetch())
      reset({ message: '' }) // Reset the message field
    } catch (e) {
      if (e instanceof Error) {
        setError('message', { message: e.message })
      }
      console.error('could not send reminders', e)
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <Tooltip hasArrow label={!poll.voteCount && `No voters yet; check census.`} placement='top'>
        <Button
          size='sm'
          onClick={onOpen}
          isLoading={isLoading}
          rightIcon={<MdSend />}
        >
          Send reminders
        </Button>
      </Tooltip>
      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Reminders</ModalHeader>
          <ModalCloseButton />
          <ModalHeader>
            <form onSubmit={handleSubmit(sendReminders)}>
              <Box display='flex' justifyContent='end'>
              <FormControl isInvalid={!!errors.message} flexGrow={1} mr={2}>
                <Input
                  size='sm'
                  placeholder='Type here the content of the reminder message...'
                  {...register('message', { required: 'This field is required' })}
                />
                <FormErrorMessage>{errors.message?.message?.toString()}</FormErrorMessage>
              </FormControl>
              <Button size='sm' type='submit' rightIcon={<MdSend />} isLoading={loading} flexGrow={1} isDisabled={selectedUsers.length == 0}>
                Send
              </Button>
              </Box>
            </form>
          </ModalHeader>
          <ModalBody>
            {error && <Check error={error} isLoading={isLoading} />}
            <UsersTable 
              size='sm' 
              users={reminders?.remindableVoters.map((profile) => [profile.username, profile.fid.toString()])} 
              selectable={true}
              onSelectionChange={(selected) => {
                setSelectedUsers(selected.map(([username, fid]) => ({ username, fid: parseInt(fid) } as Profile)))
              }}
              hasWeight={false}/>
          </ModalBody>
          <ModalFooter justifyContent='space-between' flexWrap='wrap'>
            <Button size='sm' onClick={onClose} variant='ghost' alignSelf='start'>
              Close
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </>
  )
}
