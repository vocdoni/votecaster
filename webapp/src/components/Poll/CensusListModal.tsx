import {
  Link,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalHeader,
  ModalOverlay,
  Progress,
  useDisclosure,
} from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { PropsWithChildren } from 'react'
import { useAuth } from '~components/Auth/useAuth'
import { ParticipantsTable } from '~components/Census/ParticipantsTable'
import { Check } from '~components/Check'
import { fetchCensus } from '~queries/census'

export const CensusListModal = ({ children, id }: PropsWithChildren<{ id?: string }>) => {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { bfetch } = useAuth()
  const { data, error, isLoading } = useQuery({
    queryKey: ['census', id],
    queryFn: fetchCensus(bfetch, id!),
    enabled: !!id && isOpen,
  })

  if (!id) return

  if (error) {
    return <Check error={error} isLoading={false} />
  }

  return (
    <>
      <Link onClick={onOpen} height='100%' display='flex' flexDir='column'>
        {children}
        {isLoading && <Progress isIndeterminate size='xs' marginTop='auto' />}
      </Link>
      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Participants</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <ParticipantsTable size='sm' participants={data?.participants} />
          </ModalBody>
        </ModalContent>
      </Modal>
    </>
  )
}
