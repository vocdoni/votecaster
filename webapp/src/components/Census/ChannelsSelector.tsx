import { FormControl, FormErrorMessage, FormLabel, Heading, Text } from '@chakra-ui/react'
import { AsyncSelect } from 'chakra-react-select'
import { useState } from 'react'
import { Controller, useFormContext } from 'react-hook-form'
import { useAuth } from '~components/Auth/useAuth'
import { fetchChannelQuery } from '~queries/channels'
import { ChannelSelectOption } from './ChannelSelectOption'

export type ChannelsFormValues = {
  channels: { label: string; value: string; image: string }[]
}

export const ChannelsSelector = () => {
  const {
    formState: { errors },
    setError,
    clearErrors,
  } = useFormContext<ChannelsFormValues>()
  const [loading, setLoading] = useState<boolean>(false)
  const { bfetch } = useAuth()

  return (
    <FormControl display='flex' flexDir='column' gap={4} isInvalid={!!errors.channels}>
      <Heading as={FormLabel} size='sm'>
        Add Farcaster Channels
      </Heading>
      <Text>Add the farcaster channels used by your community</Text>
      <Controller
        name='channels'
        render={({ field }) => (
          <AsyncSelect
            id='channels'
            size='sm'
            // @ts-expect-error bad typing definition (allows false or undefined but not true, which is... false)
            isMulti
            isLoading={loading}
            noOptionsMessage={({ inputValue }) => (inputValue ? 'No channels found' : 'Start typing to search')}
            placeholder='Search and add channels'
            {...field}
            components={{ Option: ChannelSelectOption }}
            loadOptions={async (inputValue) => {
              try {
                clearErrors('channels')
                setLoading(true)
                return (await fetchChannelQuery(bfetch, inputValue)()).map((channel) => ({
                  label: channel.name,
                  image: channel.image,
                  value: channel.id,
                }))
              } catch (e) {
                console.error('Could not fetch channels:', e)
                if (e instanceof Error) {
                  setError('channels', { message: e.message })
                }
                return []
              } finally {
                setLoading(false)
              }
            }}
          />
        )}
      />
      <FormErrorMessage>{errors.channels?.message?.toString()}</FormErrorMessage>
    </FormControl>
  )
}
