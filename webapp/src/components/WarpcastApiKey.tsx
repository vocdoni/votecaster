import {
  Box,
  BoxProps,
  Button,
  FormControl,
  FormErrorMessage,
  Heading,
  HStack,
  Input,
  Link,
  Text,
  VStack,
} from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import { appUrl } from '~constants'
import { fetchWarpcastAPIEnabled } from '~queries/profile'
import { useAuth } from './Auth/useAuth'
import { Check } from './Check'

type WarpcastApiKeyFormValues = {
  apikey: string
}

export const WarpcastApiKey: React.FC = (props: BoxProps) => {
  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors },
  } = useForm<WarpcastApiKeyFormValues>({
    defaultValues: {
      apikey: '',
    },
  })
  const { bfetch } = useAuth()
  const [loading, setLoading] = useState<boolean>(false)
  const { data: isAlreadyEnabled, error, isLoading, refetch } = useQuery<boolean, Error>({
    queryKey: ['apiKeyEnabled'],
    queryFn: fetchWarpcastAPIEnabled(bfetch),
  })

  const onSubmit = async (data: WarpcastApiKeyFormValues) => {
    setLoading(true)
    try {
      await bfetch(`${appUrl}/profile/warpcast`, {
        method: 'POST',
        body: JSON.stringify({ apikey: data.apikey }),
      }).then(() => refetch())
      reset({ apikey: '' }) // Reset the apikey field
    } catch (e) {
      if (e instanceof Error) {
        setError('apikey', { message: e.message })
      }
      console.error('could not set apikey', e)
    } finally {
      setLoading(false)
    }
  }

  const revokeApiKey = async () => {
    setLoading(true)
    try {
      await bfetch(`${appUrl}/profile/warpcast`, {
        method: 'POST',
        body: JSON.stringify({ apikey: null }),
      }).then(() => refetch())
      reset({ apikey: '' }) // Reset the apikey field
    } catch (e) {
      if (e instanceof Error) {
        setError('apikey', { message: e.message })
      }
      console.error('could not set apikey', e)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Box borderRadius='md' p={4} bg='purple.100' {...props}>
      <Heading fontSize='xl' mb={4} fontWeight='600' color='purple.800' pos='relative'>
        Warpcast Api Key 
      </Heading>
      <VStack spacing={4} align='stretch'>
        <Text>Set your Warpcast API Key here to unlock awesome features like poll reminders.</Text>
        { (isLoading || error) && <Check isLoading={isLoading} error={error} />}
        
        { !isAlreadyEnabled && <>
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
                <Button type='submit' colorScheme='purple' flexGrow={1} isLoading={loading}>
                  Save
                </Button>
              </HStack>
              <Text fontSize={'sm'} mt={2}>Get your Warpcast API Key from the <Link textDecoration='underline' href="https://warpcast.com/~/developers/api-keys" isExternal>official developer portal</Link>.</Text>
            </Box>
          </form>
          </>
        } 

        { isAlreadyEnabled && <>
          <HStack spacing={4} p={4} bg='purple.50' borderRadius='md'>
            <Text>You already registered a valid API Key.</Text>
            <Button colorScheme='red' isLoading={loading} onClick={revokeApiKey}>
              Revoke
            </Button>
          </HStack>
        </>}
      </VStack>
    </Box>
  )
}
