import { Box, BoxProps, Button, FormControl, FormErrorMessage, HStack, Input, Link, Text } from '@chakra-ui/react'
import { useMutation } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { appUrl } from '~constants'
import { useWarpcastApiEnabled } from '~queries/profile'
import { useAuth } from './Auth/useAuth'
import { Check } from './Check'

type WarpcastApiKeyFormValues = { apikey: string }

export const WarpcastApiKey: React.FC = (props: BoxProps) => {
  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors },
  } = useForm<WarpcastApiKeyFormValues>({
    defaultValues: { apikey: '' },
  })
  const { bfetch } = useAuth()
  const { data: isAlreadyEnabled, error, isLoading, refetch } = useWarpcastApiEnabled()

  const { mutate: setApiKey, status: setApiKeyStatus } = useMutation({
    mutationFn: async (apikey: string) => {
      await bfetch(`${appUrl}/profile/warpcast`, {
        method: 'POST',
        body: JSON.stringify({ apikey }),
      })
    },
    onSuccess: () => {
      reset({ apikey: '' })
      refetch()
    },
    onError: (error: any) => {
      if (error instanceof Error) {
        setError('apikey', { message: error.message })
      }
      console.error('could not set apikey', error)
    },
  })

  const { mutate: revokeApiKey, status: revokeApiKeyStatus } = useMutation({
    mutationFn: async () => {
      await bfetch(`${appUrl}/profile/warpcast`, {
        method: 'POST',
        body: JSON.stringify({ apikey: null }),
      })
    },
    onSuccess: () => {
      reset({ apikey: '' })
      refetch()
    },
    onError: (error: any) => {
      if (error instanceof Error) {
        setError('apikey', { message: error.message })
      }
      console.error('could not set apikey', error)
    },
  })

  const onSubmit = (data: WarpcastApiKeyFormValues) => {
    setApiKey(data.apikey)
  }

  return (
    <Box borderRadius='md' {...props}>
      {(isLoading || error) && <Check isLoading={isLoading} error={error} />}
      {!isLoading && !isAlreadyEnabled && (
        <>
          <form onSubmit={handleSubmit(onSubmit)}>
            <Box borderRadius='md' p={4} bg='purple.50'>
              <HStack spacing={4}>
                <FormControl isInvalid={!!errors.apikey}>
                  <Input
                    placeholder='Paste here your API Key'
                    {...register('apikey', { required: 'This field is required' })}
                  />
                  <FormErrorMessage>{errors.apikey?.message?.toString()}</FormErrorMessage>
                </FormControl>
                <Button type='submit' colorScheme='purple' flexGrow={1} isLoading={setApiKeyStatus === 'pending'}>
                  Save
                </Button>
              </HStack>
              <Text fontSize={'sm'} mt={2}>
                Get your Warpcast API Key from the{' '}
                <Link textDecoration='underline' href='https://warpcast.com/~/developers/api-keys' isExternal>
                  official developer portal
                </Link>
                .
              </Text>
            </Box>
          </form>
        </>
      )}
      {!isLoading && isAlreadyEnabled && (
        <>
          <HStack spacing={4} p={4} bg='purple.50' borderRadius='md'>
            <Text>You already registered a valid API Key.</Text>
            <Button colorScheme='red' isLoading={revokeApiKeyStatus === 'pending'} onClick={() => revokeApiKey()}>
              Revoke
            </Button>
          </HStack>
        </>
      )}
    </Box>
  )
}
