import { Alert, AlertIcon } from '@chakra-ui/react'

export const MaintenanceAlert = () =>
  import.meta.env.MAINTENANCE && (
    <Alert status='warning'>
      <AlertIcon />
      App is under maintenance, some features may not work as expected.
    </Alert>
  )
