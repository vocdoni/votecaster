import {
  Box,
  Button,
  FormControl,
  FormErrorMessage,
  FormLabel,
  Input,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Text,
  Textarea,
  useDisclosure,
  useToast,
  VStack
} from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { MdSend } from "react-icons/md"
import { appUrl } from '~constants'
import { useAuth } from '~components/Auth/useAuth'
import { Check } from '~components/Check'
import { WarpcastApiKey } from '~components/WarpcastApiKey'
import { UsersTable } from '~components/Census/UsersTable'
import { fetchPollsReminders } from '~queries/polls'
import { fetchWarpcastAPIEnabled } from '~queries/profile'

type ReminderFormValues = {
  castURL: string
  message: string
}

export const PollRemindersModal = ({ poll }: { poll: PollInfo }) => {
  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors, isValid },
    trigger,
  } = useForm<ReminderFormValues>({
    defaultValues: {
      message: '',
      castURL: '',
    },
  })
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { bfetch } = useAuth()
  const toast = useToast()
  const [queueId, setQueueId] = useState<string>();
  const [status, setStatus] = useState<PollReminderStatus>();
  const [loading, setLoading] = useState<boolean>(false)
  const [success, setSuccess] = useState<string>()
  const { data: isAlreadyEnabled } = useQuery<boolean, Error>({
    queryKey: ['apiKeyEnabled'],
    queryFn: fetchWarpcastAPIEnabled(bfetch),
  })
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

  useEffect(() => {
    if (!success) return
    const timer = setTimeout(() => {
      setSuccess(undefined)
    }, 5000)
    return () => clearTimeout(timer)
  }, [success])

  useEffect(() => {
    if (!queueId) return
    const interval = setInterval(async () => {
      try {
        const res = await bfetch(`${appUrl}/poll/${poll.electionId}/reminders/queue/${queueId}`)
        const data = await res.json() as PollReminderStatus
        if (data.completed) {
          setStatus(data)
          refetch()
          if (data.fails.length > 0) {
            const failedUsers = data.fails.map(([username, error]) => `${username}: ${error}`).join('\n')
            setError('message', { message: `Failed to send reminders to the following users:\n${failedUsers}` })
          }
          clearInterval(interval)
        }
      } catch (e) {
        if (e instanceof Error) {
          setError('message', { message: e.message })
        }
        console.error('could not send reminders', e)
      } finally {
        setLoading(false)
      }
    }, 500)
    return () => clearInterval(interval)
  }, [queueId])
  
  if (!poll || !poll.electionId) return

  const sendReminders = async (data: ReminderFormValues) => {
    setLoading(true)
    setSuccess(undefined)

    const users = {} as { [key: string]: string }
    selectedUsers.forEach(({ fid, username }) => {
      users[fid.toString()] = username
    })
    try {
      const res = await bfetch(`${appUrl}/poll/${poll.electionId}/reminders`, {
        method: 'POST',
        body: JSON.stringify({ 
          type: "individual",
          content: data.message + `\n\n${data.castURL}`,
          users: users,
        }),
      })
      const { queueId } = await res.json() as PollReminderQueue
      setQueueId(queueId)
      reset({ message: '', castURL: '' }) // Reset the message field
      setSuccess('Reminders sent successfully')
    } catch (e) {
      if (e instanceof Error) {
        setError('message', { message: e.message })
      }
      console.error('could not send reminders', e)
      setLoading(false)
    }
  }

  return (
    <>
      <Button
        size='sm'
        onClick={onOpen}
        isLoading={isLoading}
        rightIcon={<MdSend />}
      >
        Send reminders
      </Button>

      {isAlreadyEnabled ? (
          <Modal isOpen={isOpen} onClose={onClose} scrollBehavior='inside'>
            <ModalOverlay />
            <ModalContent>
              <ModalHeader>
                Reminders
                <Text fontSize={'sm'} color='gray' fontWeight='normal'>Send a Direct Cast to members, inviting them to vote in the poll. Please note that they will only receive the reminder if you both follow each other.</Text>
              </ModalHeader>
              <ModalCloseButton />
              <ModalHeader>
                <form onSubmit={handleSubmit(sendReminders)}>
                  <VStack spacing={4} alignItems={'start'}>
                    <Box w={'full'}>
                      <FormLabel>Content</FormLabel>
                      <FormControl isInvalid={!!errors.message} flexGrow={1} mr={2}>
                        <Textarea
                          size='sm'
                          placeholder='Type a personalized message here to invite users to participate in the poll.'
                          {...register('message', { required: 'This field is required' })}
                          onBlur={() => trigger('message')}
                        />
                        <FormErrorMessage>{errors.message?.message?.toString()}</FormErrorMessage>
                      </FormControl>
                    </Box>
                    <Box w={'full'}>
                      <FormLabel>Cast URL</FormLabel>
                      <FormControl isInvalid={!!errors.castURL} flexGrow={1} mr={2}>
                        <Input
                          size='sm'
                          placeholder='Paste the URL of a cast that includes the poll frame.'
                          {...register('castURL', {
                            required: 'Please enter Warpcast URL',
                            pattern: {
                              value: /^(https?:\/\/).+(0x[a-f\d]+)$/,
                              message: 'Invalid URL format'
                            },
                          })}
                          onBlur={() => trigger('castURL')}
                        />
                        <FormErrorMessage>{errors.castURL?.message?.toString()}</FormErrorMessage>
                      </FormControl>
                    </Box>
                  </VStack>
                </form>
              </ModalHeader>
              <ModalBody>
                {(error || success) && <Check error={error} success={success} isLoading={isLoading} />}
                <FormLabel>Select users</FormLabel>
                  <UsersTable 
                    size='sm' 
                    users={reminders?.remindableVoters.map((profile) => [profile.username, reminders?.votersWeight[profile.username]])} 
                    selectable={true}
                    findable={true}
                    onSelectionChange={(selected) => {
                      const profiles : Profile[] = []
                      selected.forEach(([username]) => {
                        const profile = reminders?.remindableVoters.find((profile) => profile.username === username)
                        if (profile) {
                          profiles.push(profile)
                        }
                      })
                      setSelectedUsers(profiles)
                    }}/>
                </ModalBody>
              <ModalFooter justifyContent='space-between' flexWrap='wrap'>
                <Text fontSize={'sm'} color='gray' fontWeight='normal' mt={2} mb={8}>You already sent {reminders?.alreadySent} reminders. You can send {reminders?.maxReminders} more.</Text>
                <Button w={'full'} size='sm' onClick={handleSubmit(sendReminders)} rightIcon={<MdSend />} isLoading={loading} flexGrow={1} isDisabled={selectedUsers.length == 0 || !isValid}>
                  Send
                </Button>
              </ModalFooter>
            </ModalContent>
          </Modal>
      ) : (
        <Modal isOpen={isOpen} onClose={onClose} >
          <ModalOverlay />
          <ModalContent p={0}>
            <ModalBody p={0}>
              <WarpcastApiKey />
            </ModalBody>
            <ModalCloseButton />
          </ModalContent>
        </Modal>
      )}
    </>
  )
}
