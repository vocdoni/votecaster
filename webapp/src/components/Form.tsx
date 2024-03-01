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
  FormHelperText,
  FormLabel,
  Heading,
  IconButton,
  Image,
  Input,
  InputGroup,
  InputRightElement,
  Link,
  ListItem,
  Radio,
  RadioGroup,
  Stack,
  Text,
  UnorderedList,
  VStack,
} from '@chakra-ui/react'
import { SignInButton } from '@farcaster/auth-kit'
import axios from 'axios'
import React, { SetStateAction, useState } from 'react'
import { FormProvider, useFieldArray, useForm } from 'react-hook-form'
import { BiTrash } from 'react-icons/bi'
import { useLogin } from '../useLogin'
import { Done } from './Done'
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
  const [usernames, setUsernames] = useState<string[]>([])
  const [status, setStatus] = useState<string | null>(null)
  const [censusRecords, setCensusRecords] = useState<number>(0)

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

  const checkCensus = async (pid: string, setStatus: Dispatch<SetStateAction<string | null>>) => {
    const checkRes = await axios.get(`${appUrl}/census/check/${pid}`)
    if (checkRes.status === 200) {
      return checkRes.data
    }
    if (checkRes.data.progress) {
      setStatus(`Creating census... ${checkRes.data.progress}%`)
    }
    // wait 3 seconds between requests
    await new Promise((resolve) => setTimeout(resolve, 3000))
    // continue retrying until we get a 200 status
    return await checkCensus(pid, setStatus)
  }

  const onSubmit = async (data: FormValues) => {
    setError(null)
    setStatus(null)
    try {
      setLoading(true)

      const election = {
        profile,
        question: data.question,
        duration: Number(data.duration),
        options: data.choices.map((c) => c.choice),
      }

      if (data.csv) {
        setStatus('Creating census...')
        // create the census
        try {
          const lineBreak = new Uint8Array([10]) // 10 is the byte value for '\n'
          const contents = new Blob(
            Array.from(data.csv).flatMap((file) => [file, lineBreak]),
            { type: 'text/csv' }
          )
          const csv = await axios.post(`${appUrl}/census/csv`, contents)
          const census = await checkCensus(csv.data.censusId, setStatus)
          setCensusRecords(census.fromTotalAddresses)
          setUsernames(census.usernames)
          census.size = census.usernames.length
          delete census.usernames
          election.census = census
        } catch (e) {
          console.error('there was an error creating the census:', e)
          if ('response' in e && 'data' in e.response) {
            setError(e.response.data)
          } else {
            if ('message' in e) {
              setError(e.message)
            }
          }
          setLoading(false)
          return
        }
      }

      setStatus('Storing poll in blockchain...')
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
      <Card maxW={{ base: '100%', md: 400, lg: 500 }}>
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
                <Done
                  pid={pid}
                  setPid={setPid}
                  usernames={usernames}
                  setUsernames={setUsernames}
                  censusRecords={censusRecords}
                />
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
                        <Radio value='custom'>Token gated via CSV</Radio>
                      </Stack>
                    </RadioGroup>
                  </FormControl>
                  {censusType === 'custom' && (
                    <FormControl isDisabled={loading} isRequired>
                      <FormLabel htmlFor='csv'>CSV files</FormLabel>
                      <Input
                        id='csv'
                        placeholder='Upload CSV'
                        type='file'
                        multiple
                        accept='text/csv,application/csv,.csv'
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
                        <FormHelperText>
                          <Alert status='info'>
                            <AlertDescription>
                              The CSV files <strong>must include Ethereum addresses and their balances</strong> from any
                              network. You can build your own at:
                              <UnorderedList>
                                <ListItem>
                                  <Link target='_blank' href='https://holders.at' variant='primary'>
                                    holders.at
                                  </Link>{' '}
                                  for NFTs
                                </ListItem>
                                <ListItem>
                                  <Link target='_blank' href='https://collectors.poap.xyz' variant='primary'>
                                    collectors.poap.xyz
                                  </Link>{' '}
                                  for POAPs
                                </ListItem>
                              </UnorderedList>
                              <strong>If an address appears multiple times, its balances will be aggregated.</strong>
                            </AlertDescription>
                          </Alert>
                        </FormHelperText>
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
                      <Button type='submit' colorScheme='purple' isLoading={loading} loadingText={status}>
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
