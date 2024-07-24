import { FormControl, FormErrorMessage, FormLabel, Input, Textarea } from '@chakra-ui/react'
import { useEffect, useState } from 'react'
import { Validations } from '~constants'
import { CharCountIndicator } from './CharCountIndicator'
import { usePollForm } from './usePollForm'

export const Question = () => {
  const {
    loading,
    form: {
      register,
      formState: { errors },
      watch,
    },
    questionPlaceholder,
  } = usePollForm()

  const maxLength = 250
  const questionValue = watch('question')
  const [currentLength, setCurrentLength] = useState(questionValue.length)

  useEffect(() => {
    if (currentLength) return
    setCurrentLength(questionValue.length)
  }, [questionValue])

  return (
    <FormControl position='relative' isRequired isDisabled={loading} isInvalid={!!errors.question}>
      <FormLabel htmlFor='question'>Question</FormLabel>
      <Input
        as={Textarea}
        id='question'
        placeholder={questionPlaceholder}
        {...register('question', {
          required: Validations.required,
          maxLength: { value: maxLength, message: 'Max length is 250 characters' },
          onChange: (e) => setCurrentLength(e.target.value.length),
        })}
      />
      <FormErrorMessage>{errors.question?.message?.toString()}</FormErrorMessage>
      <CharCountIndicator currentLength={currentLength} maxLength={maxLength} />
    </FormControl>
  )
}
