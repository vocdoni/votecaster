import { Alert, Spinner } from '@chakra-ui/react'

export const Check = ({ isLoading, error }: { isLoading: boolean; error: Error | null }) => {
  if (isLoading) {
    return <Spinner />
  }

  if (error) {
    return <Alert status='warning'>An error has occurred: {error.message}</Alert>
  }
}
