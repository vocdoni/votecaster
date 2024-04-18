import { FormControl, FormErrorMessage, FormLabel, Heading, Text } from '@chakra-ui/react'
import { AsyncSelect } from 'chakra-react-select'
import { useState } from 'react'
import { Controller, useFormContext } from 'react-hook-form'
import { fetchChannelQuery } from '../../../queries/channels'
import { useAuth } from '../../Auth/useAuth'

export type ChannelsFormValues = {
  channels: { label: string; value: string }[]
}

export const Channels = () => {
  const {
    formState: { errors },
    setError,
  } = useFormContext<ChannelsFormValues>()
  const [loading, setLoading] = useState<boolean>(false)
  const { bfetch } = useAuth()

  return (
    <FormControl display='flex' flexDir='column' gap={4} isInvalid={!!errors.channels} isRequired>
      <Heading as={FormLabel} size='sm'>
        Add Farcaster Channels
      </Heading>
      <Text>Add the farcaster channels used by your community</Text>
      <Controller
        name='channels'
        render={({ field }) => (
          <AsyncSelect
            id='channels'
            isMulti
            size='sm'
            isLoading={loading}
            noOptionsMessage={() => 'No channels found'}
            placeholder='Search and add channels'
            {...field}
            loadOptions={async (inputValue) => {
              try {
                setLoading(true)
                return (await fetchChannelQuery(bfetch)(inputValue)).map((channel) => ({
                  label: channel.name,
                  value: channel.id,
                }))
              } catch (e) {
                console.error('Could not fetch channels:', e)
                if (e instanceof Error) {
                  setError('channels', { message: e.message })
                }
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
