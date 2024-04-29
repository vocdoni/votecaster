import {
  Button,
  Flex,
  HStack,
  Icon,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Progress,
  Switch,
  Text,
  UseModalProps,
  VStack,
} from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useForm, SubmitHandler, FormProvider } from 'react-hook-form'
import { FaBell, FaEyeSlash } from 'react-icons/fa6'
import { useAuth } from '~components/Auth/useAuth'
import { appUrl } from '~constants'
import { fetchCommunity } from '~queries/communities'
import { Meta } from './Create/Meta'
import { CommunityFormValues } from './Create/Form'
import { useCallback } from 'react'

export type ManageCommunityProps = {
  communityID: number
} & UseModalProps

export type ManageCommunityFormValues = {
  disabled: boolean
} & CommunityFormValues

export const ManageCommunity = ({ communityID, ...props }: ManageCommunityProps) => {
  const { bfetch, isAuthenticated } = useAuth()
  const { data: community, refetch } = useQuery<ManageCommunityFormValues, Error>({
    queryKey: ['community'],
    queryFn: async () => {
      const c = await fetchCommunity(bfetch, `${communityID}`)()
      return {
        name: c.name,
        admins: c.admins.map((admin) => ({ label: admin.displayName, value: admin.fid })),
        logo: c.logoURL,
        groupChat: c.groupChat,
        channels: c.channels.map((channel) => ({ label: channel, value: channel })),
        enableNotifications: c.notifications,
        disabled: c.disabled,
      }
    },
  })

  const methods = useForm<ManageCommunityFormValues>({})
  
  const onSubmit: SubmitHandler<Community> = useCallback(
    async (values: Community) => {
      if (!community) return
      try {
        await bfetch(`${appUrl}/communities/${community.id}`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(values),
        }).then(() => refetch())
      } catch (e) {
        console.error('could not swithc the community notifications', e)
      }
    },
    [bfetch, community, refetch]
  )

  if (!isAuthenticated) return null
  if (!community) return null
  if (!props.isOpen) return null
  if (!props.onClose) return null


  console.log(community)


  return (
    <Modal
      size={'xl'}
      {...props}
      onClose={() => {
        props.onClose()
        refetch()
      }}
    >
      <ModalOverlay />
      <ModalContent as={'form'} onSubmit={methods.handleSubmit(onSubmit)}>
        <ModalHeader>{community.name} settings</ModalHeader>
        <ModalCloseButton />
        <Progress size='sm' isIndeterminate visibility={methods.formState.isSubmitting ? 'visible' : 'hidden'} />
        <FormProvider {...methods}>
          <ModalBody mt={2} mb={6}>
            <VStack gap={6}>
              <Meta />
              <Flex w={'100%'} justifyContent={'space-between'} alignItems={'center'} gap={6}>
                <VStack alignItems={'start'}>
                  <HStack gap={2} alignItems={'center'}>
                    <Icon as={FaBell} />
                    <Text>Notifications</Text>
                  </HStack>
                  <Text fontSize={'xs'} color={'gray'}>
                    Allow to notify community members about new polls.
                  </Text>
                </VStack>
                <HStack gap={2} alignItems={'center'}>
                  <Text fontSize={'xs'}>Disabled</Text>
                  <Switch
                    id={'notifications'}
                    disabled={methods.formState.isSubmitting}
                    colorScheme='green'
                    {...methods.register('notifications', {value: community.enableNotifications})}
                  />
                  <Text fontSize={'xs'}>Enabled</Text>
                </HStack>
              </Flex>
              <Flex w={'100%'} justifyContent={'space-between'} alignItems={'center'} gap={6}>
                <VStack alignItems={'start'}>
                  <HStack gap={2} alignItems={'center'}>
                    <Icon as={FaEyeSlash} />
                    <Text>Status</Text>
                  </HStack>
                  <Text fontSize={'xs'} color={'gray'}>
                    Disabled communities are hidden and won't be used to create polls.
                  </Text>
                </VStack>
                <HStack gap={2} alignItems={'center'}>
                  <Text fontSize={'xs'}>Enabled</Text>
                  <Switch
                    id={'status'}
                    disabled={methods.formState.isSubmitting}
                    colorScheme='red'
                    {...methods.register('disabled', {value: community.disabled})}
                  />
                  <Text fontSize={'xs'}>Disabled</Text>
                </HStack>
              </Flex>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button mt={4} colorScheme='teal' isLoading={methods.formState.isSubmitting} type='submit'>Submit</Button>
          </ModalFooter>
        </FormProvider>
      </ModalContent>
    </Modal>
  )
}
