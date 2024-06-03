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
import { QueryObserverResult, RefetchOptions } from '@tanstack/react-query'
import { useState } from 'react'
import { FormProvider, SubmitHandler, useForm } from 'react-hook-form'
import { FaBell, FaEyeSlash } from 'react-icons/fa6'
import { useAuth } from '~components/Auth/useAuth'
import { appUrl } from '~constants'
import { community2CommunityForm } from '~util/mappings'
import { ChannelsSelector } from '../Census/ChannelsSelector'
import { CommunityFormValues } from './Form'
import { CensusSelector } from './Form/CensusSelector'
import { GroupChat } from './Form/GroupChat'
import { Meta } from './Form/Meta'

export type ManageCommunityProps = {
  community: Community
  refetch: (options?: RefetchOptions | undefined) => Promise<QueryObserverResult<Community, Error>>
} & UseModalProps

export type ManageCommunityFormValues = {
  disabled: boolean
} & CommunityFormValues

export const ManageCommunity = ({ community, refetch, onClose, ...props }: ManageCommunityProps) => {
  const { bfetch, isAuthenticated } = useAuth()
  const [error, setError] = useState<Error | null>(null)
  const methods = useForm<ManageCommunityFormValues>({
    defaultValues: community2CommunityForm(community),
  })

  const onSubmit: SubmitHandler<ManageCommunityFormValues> = async (values: ManageCommunityFormValues) => {
    if (!community) return

    setError(null)
    try {
      const com: CommunityCreate = {
        ...values,
        id: community.id,
        logoURL: values.src,
        disabled: !values.disabled,
        admins: values.admins.map((admin) => ({ fid: admin.value, username: admin.label })) as Profile[],
        notifications: values.enableNotifications,
        censusAddresses: values.addresses || [],
        censusChannel: (values.channel ? { id: values.channel } : {}) as Channel,
        channels: values.channels.map((channel) => channel.value),
      }
      await bfetch(`${appUrl}/communities/${community.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(com),
      })
      // refetch data after saving to update state
      await refetch()
      onClose()
    } catch (e) {
      console.error('could not update the community data', e)
      setError(new Error(`could not update the community data`))
    }
  }

  // Modal should not be rendered in some cases
  if (!isAuthenticated || !community || !props.isOpen || !onClose) return

  return (
    <Modal
      size={'xl'}
      {...props}
      onClose={() => {
        onClose()
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
              <ChannelsSelector />
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
                  <Text fontSize={'xs'}>Disabled</Text>
                  <Switch
                    id={'disabled'}
                    disabled={methods.formState.isSubmitting}
                    colorScheme='green'
                    {...methods.register('disabled')}
                  />
                  <Text fontSize={'xs'}>Enabled</Text>
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
