import {
  Flex,
  HStack,
  Icon,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalHeader,
  ModalOverlay,
  Progress,
  Switch,
  Text,
  UseModalProps,
  VStack,
} from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { FaBell, FaEyeSlash } from 'react-icons/fa6'
import { useAuth } from '~components/Auth/useAuth'
import { appUrl } from '~constants'
import { fetchCommunity } from '~queries/communities'

export type ManageCommunityProps = {
  communityID: number
} & UseModalProps

export const ManageCommunity = ({ communityID, ...props }: ManageCommunityProps) => {
  const { bfetch, isAuthenticated } = useAuth()
  const { data: community, refetch } = useQuery<Community, Error>({
    queryKey: ['community'],
    queryFn: fetchCommunity(bfetch, `${communityID}`),
  })

  const [loadingStatus, setLoadingStatus] = useState(false)
  const [loadingNotifications, setLoadingNotifications] = useState(false)

  const isLoading = useMemo(() => loadingStatus || loadingNotifications, [loadingStatus, loadingNotifications])

  if (!isAuthenticated) return null
  if (!community) return null
  if (!props.isOpen) return null
  if (!props.onClose) return null

  const switchNotifications = async () => {
    console.log('switching notifications')
    try {
      setLoadingNotifications(true)
      await bfetch(`${appUrl}/communities/${community.id}/notifications?enabled=${!community.notifications}`, {
        method: 'PUT',
      }).then(() => refetch())
      setLoadingNotifications(false)
    } catch (e) {
      console.error('could not swithc the community notifications', e)
    }
  }

  const switchStatus = async () => {
    console.log('switching status')
    try {
      setLoadingStatus(true)
      await bfetch(`${appUrl}/communities/${community.id}/status?disabled=${!community.disabled}`, {
        method: 'PUT',
      }).then(() => refetch())
      setLoadingStatus(false)
    } catch (e) {
      console.error('could not switch the community status', e)
    }
  }

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
      <ModalContent>
        <ModalHeader>{community.name} settings</ModalHeader>
        <ModalCloseButton />
        <Progress size='sm' isIndeterminate visibility={isLoading ? 'visible' : 'hidden'} />
        <ModalBody mt={2} mb={6}>
          <VStack gap={6}>
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
                  key={'notifications'}
                  disabled={loadingNotifications}
                  onChange={switchNotifications}
                  isChecked={community.notifications}
                  colorScheme='green'
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
                  key={'status'}
                  disabled={loadingStatus}
                  onChange={switchStatus}
                  isChecked={!community.disabled}
                  colorScheme='green'
                />
                <Text fontSize={'xs'}>Enabled</Text>
              </HStack>
            </Flex>
          </VStack>
        </ModalBody>
      </ModalContent>
    </Modal>
  )
}
