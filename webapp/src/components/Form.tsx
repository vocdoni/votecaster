import {
  Alert,
  AlertDescription,
  AlertIcon,
  Box,
  Button,
  Card,
  CardBody,
  CardHeader,
  Checkbox,
  Flex,
  FlexProps,
  FormControl,
  FormErrorMessage,
  FormHelperText,
  FormLabel,
  Heading,
  Icon,
  IconButton,
  Image,
  Input,
  InputGroup,
  InputRightElement,
  Link,
  ListItem,
  Radio,
  RadioGroup,
  Select,
  Stack,
  Text,
  UnorderedList,
  VStack,
} from '@chakra-ui/react'
import { SignInButton } from '@farcaster/auth-kit'
import axios from 'axios'
import React, { SetStateAction, useEffect, useState } from 'react'
import { FormProvider, useFieldArray, useForm } from 'react-hook-form'
import { BiTrash } from 'react-icons/bi'
import Airstack from '../assets/airstack.svg?react'
import { useLogin } from '../useLogin'
import { Done } from './Done'
import logo from '/poweredby.svg'

interface Address {
  address: string
  blockchain: string
}

interface FormValues {
  question: string
  choices: { choice: string }[]
  duration?: number
  csv: File | undefined
  censusType: 'farcaster' | 'channel' | 'followers' | 'custom' | 'erc20' | 'nft'
  addresses?: Address[]
  channel?: string
  notify?: boolean
}

interface CensusResponse {
  root: string
  size: number
  uri: string
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
    resetField,
  } = methods
  const { fields, append, remove } = useFieldArray({
    control,
    name: 'choices',
  })
  const { isAuthenticated, profile, logout } = useLogin()
  const [loading, setLoading] = useState<boolean>(false)
  const [pid, setPid] = useState<string | null>(null)
  const [shortened, setShortened] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [usernames, setUsernames] = useState<string[]>([])
  const [status, setStatus] = useState<string | null>(null)
  const [censusRecords, setCensusRecords] = useState<number>(0)
  const [blockchains, setBlockchains] = useState<string[]>(['base', 'zora', 'ethereum', 'polygon'])
  const {
    fields: addressFields,
    append: appendAddress,
    remove: removeAddress,
  } = useFieldArray({
    control,
    name: 'addresses',
  })
  const censusType = watch('censusType')

  // reset shortened when no pid received
  useEffect(() => {
    if (pid) return

    setShortened(null)
  }, [pid])

  // reset notify field when censusType changes
  useEffect(() => {
    if (censusType !== 'custom') {
      resetField('notify')
    }
  }, [censusType])

  // reset address fields when censusType changes
  useEffect(() => {
    if (censusType === 'erc20' || censusType === 'nft') {
      // Remove all fields initially
      setValue('addresses', [])
      // Add one field by default
      for (let i = 0; i < 1; i++) {
        appendAddress({ address: '', blockchain: 'base' })
      }
    }
  }, [censusType, appendAddress, removeAddress])

  const checkElection = async (pid: string) => {
    try {
      const checkRes = await axios.get(`${appUrl}/create/check/${pid}`)
      if (checkRes.status === 200) {
        setPid(pid)
        if (checkRes.data.url) {
          setShortened(checkRes.data.url)
        }
        return true
      }
    } catch (error) {
      console.error('error checking election status:', error)
      return false
    }
  }

  const checkCensus = async (pid: string, setStatus: Dispatch<SetStateAction<string | null>>): CensusResponse => {
    const checkRes = await axios.get(`${appUrl}/census/check/${pid}`)
    if (checkRes.status === 200) {
      return checkRes.data as CensusResponse
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
        notifyUsers: true,
      }

      setStatus('Creating census...')
      try {
        switch (data.censusType) {
          case 'channel': {
            const channel = cleanChannel(data.channel)
            const ccensus = await axios.post(`${appUrl}/census/channel-gated/${channel}`)
            const census = await checkCensus(ccensus.data.censusId, setStatus)
            election.census = census
            break
          }
          case 'nft':
          case 'erc20':
            const tcensus = await axios.post(`${appUrl}/census/${data.censusType}`, {
              tokens: data.addresses,
            })
            const census = await checkCensus(tcensus.data.censusId, setStatus)
            election.census = census
            break
          case 'followers': {
            const fcensus = await axios.post(`${appUrl}/census/followers/${profile.fid}`, { profile })
            const census = await checkCensus(fcensus.data.censusId, setStatus)
            election.census = census
            break
          }
          case 'custom': {
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
            election.census = census
            election.notifyUsers = data.notify || false
            break
          }
          case 'farcaster':
            break
          default:
            throw new Error('specified census type does not exist')
        }
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
                  shortened={shortened}
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
                      <Stack direction='column' flexWrap='wrap'>
                        <Radio value='farcaster'>🌐 All farcaster users</Radio>
                        <Radio value='channel'>⛩ Channel gated</Radio>
                        <Radio value='followers'>❤️ My followers and me</Radio>
                        <Radio value='custom'>🦄 Token based via CSV</Radio>
                        <Radio value='nft'>
                          <Icon as={Airstack} /> NFT based via airstack
                        </Radio>
                        <Radio value='erc20'>
                          <Icon as={Airstack} /> ERC20 based via airstack
                        </Radio>
                      </Stack>
                    </RadioGroup>
                  </FormControl>
                  {['erc20', 'nft'].includes(censusType) &&
                    addressFields.map((field, index) => (
                      <FormControl key={field.id}>
                        <FormLabel>
                          {censusType.toUpperCase()} address {index + 1}
                        </FormLabel>
                        <Flex>
                          <Select
                            {...register(`addresses.${index}.blockchain`, { required })}
                            defaultValue='ethereum'
                            w='auto'
                          >
                            {blockchains.map((blockchain, key) => (
                              <option value={blockchain} key={key}>
                                {ucfirst(blockchain)}
                              </option>
                            ))}
                          </Select>
                          <InputGroup>
                            <Input
                              placeholder='Smart contract address'
                              {...register(`addresses.${index}.address`, { required })}
                            />
                            {(censusType === 'nft' || (censusType === 'erc20' && index > 0)) && (
                              <InputRightElement>
                                <IconButton
                                  aria-label='Remove address'
                                  icon={<BiTrash />}
                                  onClick={() => removeAddress(index)}
                                  size='sm'
                                />
                              </InputRightElement>
                            )}
                          </InputGroup>
                        </Flex>
                      </FormControl>
                    ))}
                  {censusType === 'nft' && addressFields.length < 3 && (
                    <Button variant='ghost' colorScheme='purple' onClick={() => appendAddress({ address: '' })}>
                      Add address
                    </Button>
                  )}
                  {censusType === 'channel' && (
                    <FormControl isDisabled={loading} isRequired isInvalid={!!errors.channel}>
                      <FormLabel htmlFor='channel'>Channel slug (URL identifier)</FormLabel>
                      <Input
                        id='channel'
                        placeholder='Enter channel i.e. degen'
                        {...register('channel', {
                          required,
                          validate: async (val) => {
                            val = cleanChannel(val)
                            try {
                              const res = await axios.get(`${appUrl}/census/channel-gated/${val}/exists`)
                              if (res.status === 200) {
                                return true
                              }
                            } catch (e) {
                              return 'Invalid channel specified'
                            }
                            return 'Invalid channel specified'
                          },
                        })}
                      />
                      <FormErrorMessage>{errors.channel?.message?.toString()}</FormErrorMessage>
                    </FormControl>
                  )}
                  {censusType === 'custom' && (
                    <>
                      <FormControl isDisabled={loading}>
                        <Checkbox {...register('notify')}>Notify farcaster users</Checkbox>
                      </FormControl>
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
                                The CSV files <strong>must include Ethereum addresses and their balances</strong> from
                                any network. You can build your own at:
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
                    </>
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

const cleanChannel = (channel: string) => channel.replace(/.*channel\//, '')

const ucfirst = (str: string) => str.charAt(0).toUpperCase() + str.slice(1)
