import { Alert, Progress } from '@chakra-ui/react'

export const Check = ({ isLoading, error, success }: { isLoading: boolean; error: Error | null; success?: string }) => {
  if (isLoading) {
    return <Progress w='full' colorScheme='purple' size='xs' isIndeterminate />
  }

  if (error) {
    return <Alert status='warning'>An error has occurred: {error.message}</Alert>
  }

  if (success) {
    return <Alert status='success'>{success}</Alert>
  }
}
