import {
  Accordion,
  AccordionButton,
  AccordionIcon,
  AccordionItem,
  AccordionPanel,
  Alert,
  AlertIcon,
  Box,
  Button,
  FormControl,
  FormLabel,
  IconButton,
  Input,
  InputGroup,
  InputRightElement,
  Text,
  Textarea,
  VStack,
} from '@chakra-ui/react'
import React from 'react'
import { FaTrash } from 'react-icons/fa6'
import { TokenSignin } from '~components/Auth/TokenSignIn'
import { useAuth } from '~components/Auth/useAuth'
import CensusTypeSelector from '~components/CensusTypeSelector'
import { Duration } from './Duration'
import { Notify } from './Notify'
import { usePollForm } from './usePollForm'

export const Composer: React.FC = () => {
  const { isAuthenticated } = useAuth()
  const {
    addOption,
    choices: { fields, remove },
    form: {
      register,
      handleSubmit,
      formState: { errors },
    },
    error,
    status,
    onSubmit,
    questionPlaceholder,
    optionPlaceholders,
    loading,
  } = usePollForm()

  return (
    <Box as='form' onSubmit={handleSubmit(onSubmit)}>
      <VStack spacing={4} alignItems='start'>
        <FormControl id='question' isInvalid={!!errors.question}>
          <FormLabel>Question</FormLabel>
          <Input
            as={Textarea}
            {...register('question', { required: 'This field is required' })}
            maxLength={150}
            placeholder={questionPlaceholder}
          />
        </FormControl>

        <FormControl as='fieldset' display='flex' flexDir='column' gap={4}>
          <FormLabel as='legend'>Choices</FormLabel>
          {fields.map((field, index) => (
            <FormControl key={field.id} isInvalid={!!errors.choices?.[index]?.choice}>
              <InputGroup>
                <Input
                  {...register(`choices.${index}.choice`, { required: 'This field is required' })}
                  defaultValue={field.choice}
                  maxLength={20}
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
            </FormControl>
          ))}
        </FormControl>

        {fields.length < 4 && (
          <Button onClick={addOption} size='sm' variant='outline' alignSelf='end' isDisabled={loading}>
            Add choice
          </Button>
        )}

        {error && (
          <Alert status='error'>
            <AlertIcon />
            {error}
          </Alert>
        )}

        <TokenSignin />
        {isAuthenticated && (
          <Button
            type='submit'
            colorScheme='purple'
            isDisabled={!isAuthenticated}
            isLoading={loading}
            w='full'
            loadingText={status}
          >
            Create poll
          </Button>
        )}
        <Accordion allowToggle>
          <AccordionItem>
            <AccordionButton>
              <AccordionIcon color='purple.500' />
              <Text fontWeight={600}>Advanced settings</Text>
            </AccordionButton>
            <AccordionPanel as={VStack} gap={4}>
              <CensusTypeSelector composer complete showAsSelect isDisabled={loading} />
              <Notify />
              <Duration />
            </AccordionPanel>
          </AccordionItem>
        </Accordion>
      </VStack>
    </Box>
  )
}
