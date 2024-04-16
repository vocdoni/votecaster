import { Spinner, Square, Text } from '@chakra-ui/react'
import { ReactNode, Suspense } from 'react'

export const Loading = () => (
  <Square centerContent size='full' minHeight='100vh'>
    <Spinner size='sm' mr={3} />
    <Text>Loading...</Text>
  </Square>
)

export const SuspenseLoader = ({ children }: { children: ReactNode }) => (
  <Suspense fallback={<Loading />}>{children}</Suspense>
)
