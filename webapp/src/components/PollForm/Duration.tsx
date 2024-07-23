import { FormControl, FormErrorMessage, FormHelperText, FormLabel, Input } from '@chakra-ui/react'
import { usePollForm } from './usePollForm'

export const Duration = () => {
  const {
    loading,
    form: {
      formState: { errors },
      register,
    },
  } = usePollForm()
  return (
    <FormControl isDisabled={loading} isInvalid={!!errors.duration}>
      <FormLabel htmlFor='duration'>Duration (Optional)</FormLabel>
      <Input
        id='duration'
        placeholder='Enter duration (in hours)'
        {...register('duration')}
        type='number'
        min={1}
        max={360} // 15 days
      />
      <FormErrorMessage>{errors.duration?.message?.toString()}</FormErrorMessage>
      <FormHelperText>24h by default</FormHelperText>
    </FormControl>
  )
}
