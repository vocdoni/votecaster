import {
  Alert,
  AlertDescription,
  AlertIcon,
  Button,
  Checkbox,
  FormControl,
  FormErrorMessage,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalHeader,
  ModalOverlay,
  Progress,
  Textarea,
  useDisclosure,
  VStack,
} from '@chakra-ui/react'
import { useMutation } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { AiOutlineNotification } from 'react-icons/ai'
import { useAuth } from '~components/Auth/useAuth'
import { WarpcastApiKey } from '~components/WarpcastApiKey'
import { appUrl, RoutePath } from '~constants'
import { useWarpcastApiEnabled } from '~queries/profile'

export const NotifyMembers = ({ community }: { community: Community }) => {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { data: apiEnabled, isLoading } = useWarpcastApiEnabled()
  return (
    <>
      <Button onClick={onOpen} leftIcon={<AiOutlineNotification />}>
        Notify members
      </Button>

      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Notify community members</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            {isLoading ? (
              <Progress size='xs' isIndeterminate />
            ) : apiEnabled ? (
              <NotifyMembersForm community={community} onClose={onClose} />
            ) : (
              <WarpcastApiKey />
            )}
          </ModalBody>
        </ModalContent>
      </Modal>
    </>
  )
}

type NotifyMembersFormValues = {
  content: string
  appendUrl: boolean
}

type NotifyMembersFormProps = {
  community: Community
  onClose: () => void
}

export const NotifyMembersForm = ({ community, onClose }: NotifyMembersFormProps) => {
  const { bfetch } = useAuth()
  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<NotifyMembersFormValues>({
    defaultValues: {
      content: '',
      appendUrl: false,
    },
  })

  const mutation = useMutation({
    mutationFn: async (data: NotifyMembersFormValues) => {
      const uri = appUrl + RoutePath.Community.replace(':chain/:id', community.id.replace(':', '/'))
      const content = data.appendUrl ? `${data.content} - ${uri}` : data.content

      console.log('content:', content, uri)

      return await bfetch(`${appUrl}/communities/${community.id}/announcements`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ content }),
      })
    },
    onSuccess: () => {
      onClose()
      reset()
    },
  })

  return (
    <form
      onSubmit={handleSubmit(async (data: NotifyMembersFormValues) => {
        console.log('data:', data)
        return await mutation.mutate(data)
      })}
    >
      <VStack spacing={4} alignItems='start'>
        <FormControl isInvalid={!!errors.content}>
          <Textarea
            id='content'
            placeholder='Enter your message here...'
            {...register('content', { required: 'Content is required' })}
          />
          <FormErrorMessage>{errors.content?.message}</FormErrorMessage>
        </FormControl>

        <FormControl>
          <Checkbox id='appendUrl' {...register('appendUrl')}>
            Append my community URL to the message
          </Checkbox>
        </FormControl>
        {mutation.error && (
          <Alert status='error'>
            <AlertIcon />
            <AlertDescription>{mutation.error.toString()}</AlertDescription>
          </Alert>
        )}

        <Button type='submit' alignSelf='end' colorScheme='purple' isLoading={mutation.status === 'pending'}>
          Send Notification
        </Button>
      </VStack>
    </form>
  )
}
