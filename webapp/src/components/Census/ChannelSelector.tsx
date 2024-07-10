import { AsyncSelect } from 'chakra-react-select'
import { forwardRef, useState } from 'react'
import { ControllerRenderProps, FieldValues, useFormContext } from 'react-hook-form'
import { useAuth } from '~components/Auth/useAuth'
import { fetchChannelQuery } from '~queries/channels'
import { ChannelSelectOption } from './ChannelSelectOption'

export type ChannelFormValues = {
  channel?: string
}

const ChannelSelector = forwardRef(
  (props: ControllerRenderProps<FieldValues, 'channel' | 'admin'> & { admin?: boolean }, ref: React.Ref<any>) => {
    const { bfetch } = useAuth()
    const { clearErrors, setError } = useFormContext<ChannelFormValues>()
    const [loading, setLoading] = useState<boolean>(false)

    return (
      <AsyncSelect
        id='channel'
        size='sm'
        isLoading={loading}
        noOptionsMessage={({ inputValue }) => (inputValue ? 'No channels found' : 'Start typing to search')}
        placeholder='Search and add channels'
        components={{ Option: ChannelSelectOption }}
        {...props}
        onChange={({ value }: { value: string }) => {
          props.onChange(value)
        }}
        value={{
          value: props.value,
          label: props.value,
        }}
        loadOptions={async (inputValue) => {
          try {
            clearErrors('channel')
            setLoading(true)
            return (await fetchChannelQuery(bfetch, inputValue, props.admin)()).map((channel) => ({
              label: channel.name,
              image: channel.image,
              value: channel.id,
            }))
          } catch (e) {
            console.error('Could not fetch channels:', e)
            if (e instanceof Error) {
              setError('channel', { message: e.message })
            }
            return []
          } finally {
            setLoading(false)
          }
        }}
        ref={ref}
      />
    )
  }
)

export default ChannelSelector
