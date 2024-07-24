import { Button, FormControl, FormLabel, IconButton, Input, InputGroup, InputRightElement } from '@chakra-ui/react'
import { FC } from 'react'
import { FaTrash } from 'react-icons/fa6'
import { CharCountIndicator } from './CharCountIndicator'
import { usePollForm } from './usePollForm'

export const Choices: FC = () => {
  const {
    addOption,
    choices: { fields, remove },
    form: {
      formState: { errors },
      register,
      watch,
    },
    loading,
    optionPlaceholders,
  } = usePollForm()

  const maxLength = 50

  return (
    <>
      <FormControl as='fieldset' display='flex' flexDir='column' gap={4} isDisabled={loading} isRequired>
        <FormLabel as='legend'>Choices</FormLabel>
        {fields.map((field, index) => (
          <FormControl key={field.id} isInvalid={!!errors.choices?.[index]?.choice} position='relative'>
            <InputGroup>
              <Input
                {...register(`choices.${index}.choice`, {
                  required: 'This field is required',
                  maxLength: { value: maxLength, message: `Max length is ${maxLength} characters` },
                })}
                defaultValue={field.choice}
                placeholder={optionPlaceholders[index]}
              />
              {fields.length > 2 && (
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
            <CharCountIndicator
              currentLength={watch(`choices.${index}.choice`, field.choice).length}
              maxLength={maxLength}
              right={fields.length > 2 ? 10 : 2}
              bottom={2.5}
            />
          </FormControl>
        ))}
      </FormControl>
      {fields.length < 4 && (
        <Button
          onClick={addOption}
          size='sm'
          variant='outline'
          colorScheme='purple'
          alignSelf='end'
          isDisabled={loading}
        >
          Add choice
        </Button>
      )}
    </>
  )
}
