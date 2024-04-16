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
  Input,
  InputGroup,
  InputRightElement,
  Switch,
  Textarea,
  VStack,
} from '@chakra-ui/react'
import React, { Dispatch, SetStateAction, useEffect, useState } from 'react'
import { FormProvider, useFieldArray, useForm } from 'react-hook-form'
import { BiTrash } from 'react-icons/bi'
import { Community } from '../queries/communities'
import { cleanChannel } from '../util/strings'
import { isErrorWithHTTPResponse, Profile } from '../util/types'
import { ReputationCard } from './Auth/Reputation'
import { SignInButton } from './Auth/SignInButton'
import { useAuth } from './Auth/useAuth'
import CensusTypeSelector, { CensusFormValues } from './CensusTypeSelector'
import { Done } from './Done'

type FormValues = CensusFormValues & {
  question: string
  choices: { choice: string }[]
  duration?: number
  notify?: boolean
  notificationText?: string
  community?: Community
}

type ElectionRequest = {
  profile: Profile
  question: string
  duration: number
  options: string[]
  notifyUsers: boolean
  notificationText?: string
  census?: CensusResponse
}

interface CID {
  censusId: string
}

interface CensusResponse {
  root: string
  size: number
  uri: string
}

interface CensusResponseWithUsernames extends CensusResponse {
  usernames: string[]
  fromTotalAddresses: number
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
    watch,
    resetField,
  } = methods
  const { fields, append, remove } = useFieldArray({
    control,
    name: 'choices',
  })
  const { isAuthenticated, profile, logout, bfetch } = useAuth()
  const [loading, setLoading] = useState<boolean>(false)
  const [pid, setPid] = useState<string | null>(null)
  const [shortened, setShortened] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [usernames, setUsernames] = useState<string[]>([])
  const [status, setStatus] = useState<string | null>(null)
  const [censusRecords, setCensusRecords] = useState<number>(0)

  const censusType = watch('censusType')
  const notify = watch('notify')

  const notifyAllowed = ['custom', 'nft', 'erc20']

  // reset shortened when no pid received
  useEffect(() => {
    if (pid) return

    setShortened(null)
  }, [pid])

  // reset notify field when censusType changes
  useEffect(() => {
    if (!notifyAllowed.includes(censusType)) {
      resetField('notify')
      resetField('notificationText')
    }
  }, [censusType])

  const checkElection = async (pid: string) => {
    try {
      const res = await bfetch(`${appUrl}/create/check/${pid}`)
      if (res.status === 200) {
        setPid(pid)
        const { url } = await res.json()
        if (url) {
          setShortened(url)
        }
        return true
      }
    } catch (error) {
      console.error('error checking election status:', error)
      return false
    }
  }

  const checkCensus = async (pid: string, setStatus: Dispatch<SetStateAction<string | null>>): CensusResponse => {
    const res = await bfetch(`${appUrl}/census/check/${pid}`)
    if (res.status === 200) {
      return (await res.json()) as CensusResponse
    }
    const data = await res.json()
    if (data.progress) {
      setStatus(`Creating census... ${data.progress}%`)
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

      if (!profile) {
        throw new Error('user not authenticated')
      }

      const election: ElectionRequest = {
        profile,
        question: data.question,
        duration: Number(data.duration),
        options: data.choices.map((c) => c.choice),
        notifyUsers: data.notify || false,
      }

      if (data.notificationText?.length) {
        election.notificationText = data.notificationText
      }

      if (!profile) {
        throw new Error('user not authenticated')
      }

      setStatus('Creating census...')
      try {
        let call: Promise<Response>
        switch (data.censusType) {
          case 'channel': {
            const channel = cleanChannel(data.channel as string)
            call = bfetch(`${appUrl}/census/channel-gated/${channel}`, { method: 'POST' })
            break
          }
          case 'nft':
          case 'erc20':
            call = bfetch(`${appUrl}/census/airstack/${data.censusType}`, {
              method: 'POST',
              body: JSON.stringify({ tokens: data.addresses }),
            })
            break
          case 'followers': {
            call = bfetch(`${appUrl}/census/followers/${profile.fid}`, {
              method: 'POST',
              body: JSON.stringify({ profile }),
            })
            break
          }
          case 'custom': {
            const lineBreak = new Uint8Array([10]) // 10 is the byte value for '\n'
            const contents = new Blob(
              Array.from(data.csv as unknown as Iterable<unknown>).flatMap((file: unknown) => [
                file as BlobPart,
                lineBreak,
              ]),
              { type: 'text/csv' }
            )
            call = bfetch(`${appUrl}/census/csv`, { method: 'POST', body: contents })
            break
          }
          case 'community': {
            if (!data.community) {
              throw new Error('community not received 🤔')
            }
            call = bfetch(`${appUrl}/census/community`, {
              method: 'POST',
              body: JSON.stringify({
                communityID: data.community?.id,
              }),
            })
            break
          }
          case 'farcaster':
            break
          default:
            throw new Error('specified census type does not exist')
        }

        if (data.censusType !== 'farcaster') {
          const res = await call
          const { censusId } = (await res.json()) as CID
          const census = (await checkCensus(censusId, setStatus)) as CensusResponseWithUsernames
          if (census.usernames && census.usernames.length) {
            setUsernames(census.usernames)
          }
          if (census.fromTotalAddresses) {
            setCensusRecords(census.fromTotalAddresses)
          }
          if (data.censusType === 'custom') {
            census.size = census.usernames.length
          }

          election.census = census
        }
      } catch (e) {
        console.error('there was an error creating the census:', e)
        if (isErrorWithHTTPResponse(e) && e.response) {
          setError(e.response.data)
        } else if (e instanceof Error) {
          setError(e.message)
        }
        setLoading(false)
        return
      }

      setStatus('Storing poll in blockchain...')
      const res = await bfetch(`${appUrl}/create`, {
        method: 'POST',
        body: JSON.stringify(election),
      })
      const id = (await res.text()).replace('\n', '')

      // this is a piece of 💩 made by GPT and I should rewrite it anytime soon
      const intervalId = window.setInterval(async () => {
        const success = await checkElection(id)
        if (success) {
          clearInterval(intervalId)
          setLoading(false)
        }
      }, 1000)
    } catch (e) {
      console.error('there was an error creating the election:', e)
      if (e instanceof Error) {
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
    <Flex flexDir='column' alignItems='center' w={{ base: 'full', sm: 450, md: 500 }} {...props}>
      <Card w='100%'>
        <CardHeader textAlign='center'>
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
                  <CensusTypeSelector complete isDisabled={loading} />
                  {notifyAllowed.includes(censusType) && (
                    <FormControl isDisabled={loading}>
                      <Switch {...register('notify')} lineHeight={6}>
                        Notify farcaster users via cast (only for censuses &lt; 1k)
                      </Switch>
                    </FormControl>
                  )}
                  {notify && (
                    <FormControl isDisabled={loading}>
                      <FormLabel>Custom notification text</FormLabel>
                      <Textarea
                        placeholder='Additional text when notifying users (optional, max 150 characters)'
                        maxLength={150}
                        {...register('notificationText')}
                      />
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
                      <Box fontSize='xs' textAlign='right'>
                        or{' '}
                        <Button variant='text' size='xs' p={0} onClick={logout} height='auto'>
                          logout
                        </Button>
                      </Box>
                      <ReputationCard />
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
          </FormProvider>
        </CardBody>
      </Card>
    </Flex>
  )
}

export default Form
