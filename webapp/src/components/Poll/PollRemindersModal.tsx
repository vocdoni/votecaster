import {
  Alert,
  AlertDescription,
  AlertIcon,
  Box,
  Button,
  FormControl,
  FormErrorMessage,
  FormLabel,
  HStack,
  Icon,
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
  VStack,
} from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { FaFlagCheckered } from 'react-icons/fa6'
import { MdSend } from 'react-icons/md'
import { useAuth } from '~components/Auth/useAuth'
import { UsersTable } from '~components/Census/UsersTable'
import { Check } from '~components/Check'
import { WarpcastApiKey } from '~components/WarpcastApiKey'
import { appUrl } from '~constants'
import { fetchPollsReminders } from '~queries/polls'
import { fetchWarpcastAPIEnabled } from '~queries/profile'

type ReminderFormValues = {
  message: string
}

export const PollRemindersModal = ({ poll, frameURL }: { poll: PollInfo; frameURL?: string }) => {
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
    },
  })
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { bfetch } = useAuth()
  const toast = useToast()
  const [queueId, setQueueId] = useState<string>()
  const [loading, setLoading] = useState<boolean>(false)
  const [success, setSuccess] = useState<string>()
  const { data: isAlreadyEnabled } = useQuery<boolean, Error>({
    queryKey: ['apiKeyEnabled'],
    queryFn: fetchWarpcastAPIEnabled(bfetch),
  })
  const {
    data: reminders,
    error,
    isLoading,
    refetch,
  } = useQuery({
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

  const [selectedUsers, setSelectedUsers] = useState<Profile[]>([])

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
        const data = (await res.json()) as PollReminderStatus
        if (data.completed) {
          clearInterval(interval)
          setQueueId(undefined)
          if (!!data.fails && data.fails.length > 0) {
            const failedUsers = data.fails.map(([username, error]) => `${username}: ${error}`).join('\n')
            setError('message', { message: `Failed to send reminders to the following users:\n${failedUsers}` })
          }
          refetch()
          setLoading(false)
        }
      } catch (e) {
        if (e instanceof Error) {
          setError('message', { message: e.message })
        }
        console.error('could not send reminders', e)
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
          type: 'individual',
          content: data.message + `\n\n${frameURL}`,
          users: users,
        }),
      })
      const { queueId } = (await res.json()) as PollReminderQueue
      setQueueId(queueId)
      reset({ message: '' }) // Reset the message field
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
      <Button size='sm' onClick={onOpen} isLoading={isLoading} rightIcon={<MdSend />}>
        Send reminders
      </Button>

      {isAlreadyEnabled ? (
        <Modal isOpen={isOpen && !isLoading} onClose={onClose} scrollBehavior='inside'>
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>
              Reminders
              <Text fontSize={'sm'} color='gray' fontWeight='normal'>
                Send a Direct Cast to members, inviting them to vote in the poll. Please note that they will only
                receive the reminder if you both follow each other.
              </Text>
            </ModalHeader>
            <ModalCloseButton />
            {reminders?.remindableVoters?.length && !error ? (
              <>
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
                    </VStack>
                  </form>
                </ModalHeader>
                <ModalBody>
                  {(error || success) && <Check error={error} success={success} isLoading={isLoading} />}
                  <FormLabel>Select users</FormLabel>
                  <UsersTable
                    size='sm'
                    users={reminders?.remindableVoters.map((profile) => [
                      profile.username,
                      reminders?.votersWeight[profile.username],
                    ])}
                    selectable={true}
                    findable={true}
                    onSelectionChange={(selected) => {
                      const profiles: Profile[] = []
                      selected.forEach(([username]) => {
                        const profile = reminders?.remindableVoters.find((profile) => profile.username === username)
                        if (profile) {
                          profiles.push(profile)
                        }
                      })
                      setSelectedUsers(profiles)
                    }}
                  />
                </ModalBody>
                <ModalFooter justifyContent='space-between' flexWrap='wrap'>
                  <Text fontSize={'sm'} color='gray' fontWeight='normal' mt={2} mb={8}>
                    You already sent {reminders?.alreadySent} reminders. You can send {reminders?.maxReminders} more.
                  </Text>
                  <Button
                    w={'full'}
                    size='sm'
                    onClick={handleSubmit(sendReminders)}
                    rightIcon={<MdSend />}
                    isLoading={loading}
                    flexGrow={1}
                    isDisabled={selectedUsers.length == 0 || !isValid}
                  >
                    Send
                  </Button>
                </ModalFooter>
              </>
            ) : (
              <ModalBody>
                <HStack spacing={4} alignItems={'center'} mb={4}>
                  {error ? (
                    <Alert status='error'>
                      <AlertIcon />
                      <AlertDescription>{error.toString()}</AlertDescription>
                    </Alert>
                  ) : (
                    <>
                      <Icon as={FaFlagCheckered} />
                      <VStack align={'start'} spacing={1}>
                        <Text fontWeight='semibold' fontSize={'sm'}>
                          All the reminders have been sent.
                        </Text>
                        <Text fontSize={'sm'}>There are no users left to send reminders to.</Text>
                      </VStack>
                    </>
                  )}
                </HStack>
              </ModalBody>
            )}
          </ModalContent>
        </Modal>
      ) : (
        <Modal isOpen={isOpen} onClose={onClose}>
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
