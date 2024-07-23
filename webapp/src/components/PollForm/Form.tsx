import {
  Alert,
  AlertDescription,
  AlertIcon,
  Box,
  Button,
  Card,
  CardBody,
  CardHeader,
  Flex,
  FlexProps,
  FormControl,
  FormErrorMessage,
  FormLabel,
  Heading,
  IconButton,
  Input,
  InputGroup,
  InputRightElement,
  VStack,
} from '@chakra-ui/react'
import type { FC } from 'react'
import { BiTrash } from 'react-icons/bi'
import { ReputationCard } from '~components/Auth/Reputation'
import { SignInButton } from '~components/Auth/SignInButton'
import { useAuth } from '~components/Auth/useAuth'
import CensusTypeSelector from '~components/CensusTypeSelector'
import { Done } from './Done'
import { Duration } from './Duration'
import { Notify } from './Notify'
import { usePollForm } from './usePollForm'

type FormProps = FlexProps & {
  communityId?: CommunityID
  composer?: boolean
}

const Form: FC<FormProps> = ({ communityId, composer, ...props }) => {
  const {
    error,
    pid,
    usernames,
    form: methods,
    onSubmit,
    loading,
    choices: { fields, append, remove },
    optionPlaceholders,
    questionPlaceholder,
    status,
  } = usePollForm()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = methods
  const notify = methods.watch('notify')
  const { isAuthenticated, reputation, logout } = useAuth()

  const required = {
    value: true,
    message: 'This field is required',
  }
  const maxLength = {
    value: 50,
    message: 'Max length is 50 characters',
  }

  return (
    <Flex flexDir='column' alignItems='center' w={{ base: 'full', sm: 450, md: 550 }} {...props}>
      <Card w='100%' borderRadius={composer ? 0 : 6}>
        <CardHeader textAlign='center'>
          <Heading as='h2' size='lg' textAlign='center'>
            Create a framed poll
          </Heading>
        </CardHeader>
        <CardBody>
          <VStack as='form' onSubmit={handleSubmit(onSubmit)} spacing={4} align='left'>
            {pid ? (
              <Done />
            ) : (
              <>
                <FormControl isRequired isDisabled={loading} isInvalid={!!errors.question}>
                  <FormLabel htmlFor='question'>Question</FormLabel>
                  <Input
                    id='question'
                    placeholder={questionPlaceholder}
                    {...register('question', {
                      required,
                      maxLength: { value: 250, message: 'Max length is 250 characters' },
                    })}
                  />
                  <FormErrorMessage>{errors.question?.message?.toString()}</FormErrorMessage>
                </FormControl>
                {fields.map((field, index) => (
                  <FormControl
                    key={field.id}
                    isRequired={index < 2}
                    isDisabled={loading}
                    isInvalid={!!errors.choices?.[index]}
                  >
                    <FormLabel>Choice {index + 1}</FormLabel>
                    <InputGroup>
                      <Input
                        placeholder={optionPlaceholders[index]}
                        {...register(`choices.${index}.choice`, { required, maxLength })}
                      />
                      {fields.length > 2 && (
                        <InputRightElement>
                          <IconButton
                            size='sm'
                            aria-label='Remove choice'
                            icon={<BiTrash />}
                            onClick={() => remove(index)}
                          />
                        </InputRightElement>
                      )}
                    </InputGroup>
                    <FormErrorMessage>{errors.choices?.[index]?.choice?.message?.toString()}</FormErrorMessage>
                  </FormControl>
                ))}
                {fields.length < 4 && (
                  <Button alignSelf='end' onClick={() => append({ choice: '' })} isDisabled={loading}>
                    Add Choice
                  </Button>
                )}
                <CensusTypeSelector complete isDisabled={loading} composer={composer} communityId={communityId} />
                <Notify />
                <Duration />

                {error && (
                  <Alert status='error'>
                    <AlertIcon />
                    {error}
                  </Alert>
                )}

                {notify && usernames.length > 1000 && (
                  <Alert status='warning'>
                    <AlertIcon />
                    <AlertDescription>
                      Selected census contains more than 1,000 farcaster users. Won't be notifying them.
                    </AlertDescription>
                  </Alert>
                )}
                {isAuthenticated ? (
                  <>
                    <Button type='submit' isLoading={loading} loadingText={status}>
                      Create
                    </Button>
                    {!composer && (
                      <>
                        <Box fontSize='xs' textAlign='right'>
                          or{' '}
                          <Button variant='text' size='xs' p={0} onClick={logout} height='auto'>
                            logout
                          </Button>
                        </Box>
                        <ReputationCard reputation={reputation!} />
                      </>
                    )}
                  </>
                ) : (
                  <Box display='flex' justifyContent='center' alignItems='center' flexDir='column'>
                    <SignInButton size='lg' />
                    to create a poll
                  </Box>
                )}
              </>
            )}
          </VStack>
        </CardBody>
      </Card>
    </Flex>
  )
}

export default Form
