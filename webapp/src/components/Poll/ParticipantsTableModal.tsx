import { Button, useDisclosure, } from '@chakra-ui/react'
import { FaUserGroup } from 'react-icons/fa6'
import { UsersTableModal } from './UsersTableModal'

export const ParticipantsTableModal = ({ poll, census }: { poll: PollInfo, census: Census }) => {
  const { isOpen, onOpen, onClose } = useDisclosure()

  if (!poll || !poll.electionId) return

  return (
    <>
      <Button size='sm' onClick={onOpen}  rightIcon={<FaUserGroup />}>
        Census
      </Button>
      <UsersTableModal
        isOpen={isOpen}
        onClose={onClose}
        downloadText='Download full census'
        error={null}
        isLoading={false}
        title='Participants / census'
        filename='participants.csv'
        data={
          census?.participants &&
          Object.keys(census.participants).map((username) => [username, census.participants[username]])
        }
      />
    </>
  )
}
