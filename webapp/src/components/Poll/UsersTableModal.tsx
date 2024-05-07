import {
  Button,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
} from '@chakra-ui/react'
import { FaDownload } from 'react-icons/fa6'
import { UsersTable } from '~components/Census/UsersTable'
import { Check } from '~components/Check'
import { CsvGenerator } from '~src/generator'
import { downloadFile } from '~util/files'

type UsersTableModalProps = {
  data: string[][] | undefined
  isOpen: boolean
  onClose: () => void
  error: Error | null
  isLoading: boolean
  title: string
  downloadText: string
  filename: string
}

export const UsersTableModal = ({
  data,
  error,
  filename,
  isLoading,
  isOpen,
  onClose,
  title,
  downloadText,
}: UsersTableModalProps) => {
  if (!data || !data.length) return

  const download = () => {
    const csv = new CsvGenerator(data[0].length > 1 ? ['Username', 'Weight'] : ['Username'], data, filename)

    downloadFile(csv.url, csv.filename)
  }

  return (
    <>
      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>{title}</ModalHeader>
          <ModalCloseButton />
          <ModalHeader display='flex' justifyContent='end'>
            <Button size='sm' rightIcon={<FaDownload />} onClick={download}>
              {downloadText}
            </Button>
          </ModalHeader>
          <ModalBody>
            {error && <Check error={error} isLoading={isLoading} />}
            <UsersTable size='sm' users={data} />
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
