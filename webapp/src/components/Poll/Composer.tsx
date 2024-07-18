import { Box, Button, FormControl, IconButton, Input, InputGroup, InputRightElement, VStack } from '@chakra-ui/react'
import React from 'react'
import { Controller, SubmitHandler, useFieldArray, useForm } from 'react-hook-form'
import { FaTrash } from 'react-icons/fa6'

interface IFormInput {
  question: string
  options: { value: string }[]
}

export const Composer: React.FC = () => {
  const {
    control,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<IFormInput>({
    defaultValues: {
      question: '',
      options: [{ value: '' }, { value: '' }], // Minimum two options
    },
  })

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'options',
  })

  const onSubmit: SubmitHandler<IFormInput> = (data) => {
    console.log(data)
    reset()
  }

  const addOption = () => {
    if (fields.length < 4) {
      append({ value: '' })
    }
  }

  return (
    <Box as='form' onSubmit={handleSubmit(onSubmit)} w='full'>
      <VStack spacing={4} alignItems='start'>
        <FormControl id='question' isInvalid={!!errors.question}>
          <Controller
            name='question'
            control={control}
            rules={{ required: 'This field is required' }}
            render={({ field }) => <Input {...field} maxLength={150} placeholder='Enter your question...' />}
          />
        </FormControl>

        {fields.map((field, index) => (
          <FormControl key={field.id} id={`option${index + 1}`} isInvalid={!!errors.options?.[index]?.value}>
            <Controller
              name={`options.${index}.value`}
              control={control}
              rules={{ required: 'This field is required' }}
              render={({ field }) => (
                <InputGroup>
                  <Input {...field} maxLength={20} placeholder={`Option #${index + 1}`} />
                  {index >= 2 && (
                    <InputRightElement>
                      <IconButton
                        aria-label='Remove option'
                        icon={<FaTrash />}
                        size='sm'
                        variant='ghost'
                        onClick={() => remove(index)}
                        colorScheme='red'
                      />
                    </InputRightElement>
                  )}
                </InputGroup>
              )}
            />
          </FormControl>
        ))}

        {fields.length < 4 && (
          <Button onClick={addOption} size='sm' variant='outline'>
            Add option
          </Button>
        )}

        <Button type='submit' colorScheme='purple' w='full'>
          Create poll
        </Button>
      </VStack>
    </Box>
  )
}
