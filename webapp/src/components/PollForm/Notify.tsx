import { FormControl, FormLabel, Input, Switch, Textarea } from '@chakra-ui/react'
import { usePollForm } from './usePollForm'

export const Notify = () => {
  const {
    censusType,
    loading,
    notifyAllowed,
    form: { watch, register },
  } = usePollForm()
  const notify = watch('notify')

  return (
    <>
      {notifyAllowed.includes(censusType) && (
        <FormControl isDisabled={loading}>
          <Switch {...register('notify')} lineHeight={6}>
            🔔 Notify farcaster users via cast (only for censuses &lt; 1k)
          </Switch>
        </FormControl>
      )}
      {notify && (
        <FormControl isDisabled={loading}>
          <FormLabel>Custom notification text</FormLabel>
          <Input
            as={Textarea}
            placeholder='Additional text when notifying users (optional, max 150 characters)'
            maxLength={150}
            {...register('notificationText')}
          />
        </FormControl>
      )}
    </>
  )
}
