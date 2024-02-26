import {
  Alert,
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
  FormHelperText,
  FormLabel,
  Heading,
  IconButton,
  Image,
  Input,
  InputGroup,
  InputRightElement,
  Link,
  Radio,
  RadioGroup,
  Stack,
  Text,
  VStack,
} from '@chakra-ui/react'
import { SignInButton } from '@farcaster/auth-kit'
import axios from 'axios'
import React, { useState } from 'react'
import { FormProvider, useFieldArray, useForm } from 'react-hook-form'
import { BiTrash } from 'react-icons/bi'
import { Done } from './Done'
import { useLogin } from './useLogin'
import logo from '/poweredby.svg'

interface FormValues {
  question: string
  choices: { choice: string }[]
  duration?: number
  csv: File | undefined
  censusType: string
}

const appUrl = import.meta.env.APP_URL

const Form: React.FC = (props: FlexProps) => {
  const methods = useForm<FormValues>({
    defaultValues: {
      choices: [{ choice: '' }, { choice: '' }],
      censusType: 'farcaster',
    },
  })
  const {
    register,
    handleSubmit,
    formState: { errors },
    control,
    setValue,
    watch,
  } = methods
  const { fields, append, remove } = useFieldArray({
    control,
    name: 'choices',
  })
  const { isAuthenticated, profile, logout } = useLogin()
  const [loading, setLoading] = useState<boolean>(false)
  const [pid, setPid] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const checkElection = async (pid: string) => {
    try {
      const checkRes = await axios.get(`${appUrl}/create/check/${pid}`)
      if (checkRes.status === 200) {
        setPid(pid)
        return true
      }
    } catch (error) {
      console.error('error checking election status:', error)
      return false
    }
  }

  const checkCensus = async (pid: string) => {
    try {
      const checkRes = await axios.get(`${appUrl}/census/check/${pid}`)
      if (checkRes.status === 200) {
        return checkRes.data
      }
      // wait 3 seconds between requests
      await new Promise((resolve) => setTimeout(resolve, 3000))
      // continue retrying until we get a 200 status
      return await checkCensus(pid)
    } catch (error) {
      console.error('error checking census status:', error)
      return false
    }
  }

  const onSubmit = async (data: FormValues) => {
    setError(null)
    try {
      setLoading(true)

      const election = {
        profile,
        question: data.question,
        duration: Number(data.duration),
        options: data.choices.map((c) => c.choice),
      }

      if (data.csv) {
        // create the census
        const csv = await axios.post(`${appUrl}/census/csv`, data.csv[0])
        const census = await checkCensus(csv.data.censusId)
        if (census === false) {
          setError('There was an error creating the census')
          setLoading(false)
          return
        }
        console.log('census properly created:', census)
        election.census = census
      }

      const res = await axios.post(`${appUrl}/create`, election)
      const intervalId = window.setInterval(async () => {
        const success = await checkElection(res.data.replace('\n', ''))
        if (success) {
          clearInterval(intervalId)
          setLoading(false)
        }
      }, 1000)
    } catch (e) {
      console.error('there was an error creating the election:', e)
      if ('message' in e) {
        setError(e.message)
      }
      setLoading(false)
    }
  }

  const required = {
    value: true,
    message: 'This field is required',
  }
  const maxLength = {
    value: 50,
    message: 'Max length is 50 characters',
  }

  const censusType = watch('censusType')

  return (
    <Flex flexDir='column' alignItems='center' {...props}>
      <Card maxW={{ base: '100%', md: 400, lg: 600 }}>
        <CardHeader align='center'>
          <Heading as='h1' size='2xl'>
            farcaster.vote
          </Heading>
          <Image src={logo} alt='powered by vocdoni' mb={4} width='50%' />
          <Heading as='h2' size='lg' textAlign='center'>
            Create a framed poll
          </Heading>
        </CardHeader>
        <CardBody>
          <FormProvider {...methods}>
            <VStack as='form' onSubmit={handleSubmit(onSubmit)} spacing={4} align='left'>
              {pid ? (
                <Done pid={pid} setPid={setPid} />
              ) : (
                <>
                  <FormControl isRequired isDisabled={loading} isInvalid={!!errors.question}>
                    <FormLabel htmlFor='question'>Question</FormLabel>
                    <Input
                      id='question'
                      placeholder='Enter your question'
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
                          placeholder={`Enter choice ${index + 1}`}
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
                  <FormControl isDisabled={loading}>
                    <FormLabel>Census/voters</FormLabel>
                    <RadioGroup onChange={(val: string) => setValue('censusType', val)} value={censusType}>
                      <Stack direction='row'>
                        <Radio value='farcaster'>All farcaster users</Radio>
                        <Radio value='custom'>Custom via CSV</Radio>
                      </Stack>
                    </RadioGroup>
                  </FormControl>
                  {censusType === 'custom' && (
                    <FormControl isDisabled={loading} isRequired>
                      <FormLabel htmlFor='csv'>CSV</FormLabel>
                      <Input
                        id='csv'
                        placeholder='Select CSV'
                        type='file'
                        accept='text/csv,.csv'
                        {...register('csv', {
                          required: {
                            value: true,
                            message: 'This field is required',
                          },
                        })}
                      />
                      {errors.csv ? (
                        <FormErrorMessage>{errors.csv?.message?.toString()}</FormErrorMessage>
                      ) : (
                        <FormHelperText>Requires hex addresses linked to farcaster accounts</FormHelperText>
                      )}
                    </FormControl>
                  )}
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
                  {error && (
                    <Alert status='error'>
                      <AlertIcon />
                      {error}
                    </Alert>
                  )}

                  {isAuthenticated ? (
                    <>
                      <Button type='submit' colorScheme='purple' isLoading={loading}>
                        Create
                      </Button>
                      <Box fontSize='xs' align='right'>
                        or{' '}
                        <Button variant='text' size='xs' p={0} onClick={logout} height='auto'>
                          logout
                        </Button>
                      </Box>
                    </>
                  ) : (
                    <Box display='flex' justifyContent='center' alignItems='center' flexDir='column'>
                      <SignInButton />
                      to create a poll
                    </Box>
                  )}
                </>
              )}
            </VStack>
          </FormProvider>
        </CardBody>
      </Card>
      <Text mt={3} fontSize='.8em' textAlign='center'>
        <Link href='https://warpcast.com/vocdoni' target='_blank'>
          By @vocdoni
        </Link>
      </Text>
    </Flex>
  )
}

export default Form
