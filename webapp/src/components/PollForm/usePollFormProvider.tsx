import { Dispatch, SetStateAction, useEffect, useMemo, useState } from 'react'
import { useFieldArray, useForm } from 'react-hook-form'
import { useLocation } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'
import { CensusFormValues } from '~components/CensusTypeSelector'
import { appUrl, getRandomPollOption, getRandomPollQuestion } from '~constants'
import { cleanChannel } from '~util/strings'
import { isErrorWithHTTPResponse } from '~util/types'

export type PollFormProviderProps = NonNullable<unknown>

type FormValues = CensusFormValues & {
  question: string
  choices: { choice: string }[]
  duration?: number
  notify?: boolean
  notificationText?: string
  community?: Community
}

export const usePollFormProvider = () => {
  const { profile, bfetch } = useAuth()
  const [loading, setLoading] = useState<boolean>(false)
  const [pid, setPid] = useState<string | null>(null)
  const [shortened, setShortened] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [usernames, setUsernames] = useState<string[]>([])
  const [status, setStatus] = useState<string | null>(null)
  const [censusRecords, setCensusRecords] = useState<number>(0)
  const [cast, setCast] = useState<string | null>(null)
  const { search } = useLocation()
  const form = useForm<FormValues>({
    defaultValues: {
      censusType: 'farcaster',
      question: '',
      choices: [{ choice: '' }, { choice: '' }], // Minimum two options
    },
  })
  const questionPlaceholder = useMemo(() => getRandomPollQuestion(), [])
  const optionPlaceholders = useMemo(() => {
    const uniqueOptions = new Set<string>()
    while (uniqueOptions.size < 4) {
      uniqueOptions.add(getRandomPollOption())
    }
    return Array.from(uniqueOptions)
  }, [])
  const choices = useFieldArray({
    control: form.control,
    name: 'choices',
  })
  const addOption = () => {
    if (choices.fields.length < 4) {
      choices.append({ choice: '' })
    }
  }
  const censusType = form.watch('censusType')
  const notifyAllowed = ['community']

  // reset shortened when no pid received
  useEffect(() => {
    if (pid) return

    setShortened(null)
  }, [pid])

  // reset notify field when censusType changes
  useEffect(() => {
    if (!notifyAllowed.includes(censusType)) {
      form.resetField('notify')
      form.resetField('notificationText')
    }
  }, [censusType])

  // set question if received via GET query param
  useEffect(() => {
    const params = new URLSearchParams(search)
    const question = params.get('question')
    if (!question) return

    form.setValue('question', question)
    setCast(question)
  }, [search])

  const waitForElection = async (id: string) => {
    const success = await checkElection(id)
    if (!success) {
      await new Promise((resolve) => setTimeout(resolve, 1000))
      await waitForElection(id)
      return
    }
    setLoading(false)
    // composer actions required stuff (injects the cast into the cast composer)
    window.parent.postMessage(
      {
        type: 'createCast',
        data: {
          cast: {
            parent: '',
            text: cast,
            embeds: [success],
          },
        },
      },
      '*'
    )
  }

  const checkElection = async (pid: string) => {
    try {
      const res = await bfetch(`${appUrl}/create/check/${pid}`)
      if (res.status === 200) {
        setPid(pid)
        const { url } = await res.json()
        if (url) {
          setShortened(url)
        }
        return url
      }
    } catch (error) {
      console.error('error checking election status:', error)
      return false
    }
  }

  const checkCensus = async (
    pid: string,
    setStatus: Dispatch<SetStateAction<string | null>>
  ): Promise<CensusResponse> => {
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

      if (data.community && !data.community.admins.find((admin: Profile) => admin.fid === profile.fid)) {
        throw new Error('you are not an admin of this community')
      }

      const election: PollRequest = {
        profile,
        question: data.question,
        duration: Number(data.duration),
        options: data.choices.map((c) => c.choice),
        notifyUsers: data.notify || false,
        community: data.community?.id || undefined,
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
          case 'alfafrens': {
            call = bfetch(`${appUrl}/census/alfafrens`, { method: 'POST' })
            break
          }
          case 'farcaster':
            break
          default:
            throw new Error('specified census type does not exist')
        }

        if (data.censusType !== 'farcaster') {
          // @ts-expect-error false positive since we know `call` is defined for all census type but farcaster
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

      await waitForElection(id)
    } catch (e) {
      console.error('there was an error creating the election:', e)
      if (e instanceof Error) {
        setError(e.message)
      }
      setLoading(false)
    }
  }

  return {
    addOption,
    cast,
    censusRecords,
    censusType,
    choices,
    error,
    form,
    loading,
    onSubmit,
    optionPlaceholders,
    pid,
    questionPlaceholder,
    setPid,
    setUsernames,
    shortened,
    status,
    usernames,
    notifyAllowed,
  }
}
