import {
  Alert,
  AlertIcon,
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
import { useCallback, useEffect, useState } from 'react'
import { FormProvider, SubmitHandler, useForm } from 'react-hook-form'
import { FaBell, FaEyeSlash } from 'react-icons/fa6'
import { useAuth } from '~components/Auth/useAuth'
import { appUrl } from '~constants'
import { fetchCommunity } from '~queries/communities'
import { CommunityFormValues } from './Create/Form'
import { Meta } from './Create/Meta'
import { CensusSelector } from './Create/CensusSelector'
import { Channels } from './Create/Channels'
import { GroupChat } from './Create/GroupChat'

export type ManageCommunityProps = {
  communityID: number
} & UseModalProps

export type ManageCommunityFormValues = {
  disabled: boolean
} & CommunityFormValues

export const ManageCommunity = ({ communityID, ...props }: ManageCommunityProps) => {
  const { bfetch, isAuthenticated } = useAuth()
  const { data: community, refetch } = useQuery<Community, Error, ManageCommunityFormValues>({
    queryKey: ['community', communityID],
    queryFn: fetchCommunity(bfetch, `${communityID}`),
    select: (data) => ({
      censusType: data.censusType as CensusType,
      name: data.name,
      admins: data.admins.map((admin) => ({ label: admin.username, value: admin.fid })),
      src: data.logoURL,
      groupChat: data.groupChat,
      channel: data.censusChannel.id,
      channels: data.channels.map((channel) => ({ label: channel, value: channel })),
      enableNotifications: data.notifications,
      disabled: data.disabled,
    }),
  })
  const [error, setError] = useState<Error | null>(null)
  const methods = useForm<ManageCommunityFormValues>({
    defaultValues: community,
  })

  const onSubmit: SubmitHandler<ManageCommunityFormValues> = useCallback(
    async (values: ManageCommunityFormValues) => {
      if (!community) return
      try {
        const community: Community = {
          id: communityID,
          name: values.name,
          logoURL: values.src,
          admins: values.admins.map((admin) => ({ fid: admin.value, username: admin.label })) as Profile[],
          notifications: values.enableNotifications,
          censusType: values.censusType as CensusType,
          censusAddresses: values.addresses || [],
          censusChannel: (values.channel ? { id: values.channel } : {}) as Channel,
          channels: values.channels.map((channel) => channel.value),
          groupChat: values.groupChat,
          disabled: values.disabled,
        }
        await bfetch(`${appUrl}/communities/${communityID}`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(community),
        }).then(() => refetch())
      } catch (e) {
        console.error('could not update the community data', e)
        setError(new Error(`could not update the community data`))
      }
    },
    [bfetch, community, refetch, communityID]
  )

  // Reset form when community changes
  useEffect(() => {
    if (community) methods.reset(community)
  }, [community])

  // Modal should not be rendered in some cases
  if (!isAuthenticated || !community || !props.isOpen || !props.onClose) return

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
              {!!error && (
                <Alert status='warning'>
                  <AlertIcon />
                  {error.toString()}
                </Alert>
              )}
              <Meta />
              <CensusSelector />
              <GroupChat />
              <Channels />
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
                    id={'enableNotifications'}
                    disabled={methods.formState.isSubmitting}
                    colorScheme='green'
                    {...methods.register('enableNotifications')}
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
                    id={'disabled'}
                    disabled={methods.formState.isSubmitting}
                    colorScheme='red'
                    {...methods.register('disabled')}
                  />
                  <Text fontSize={'xs'}>Disabled</Text>
                </HStack>
              </Flex>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button mt={4} colorScheme='teal' isLoading={methods.formState.isSubmitting} type='submit'>
              Submit
            </Button>
          </ModalFooter>
        </FormProvider>
      </ModalContent>
    </Modal>
  )
}
