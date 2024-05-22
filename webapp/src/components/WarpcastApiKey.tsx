import {
  Box,
  BoxProps,
  Button,
  FormControl,
  FormErrorMessage,
  Heading,
  HStack,
  Icon,
  Input,
  Link,
  Popover,
  PopoverArrow,
  PopoverBody,
  PopoverCloseButton,
  PopoverContent,
  PopoverTrigger,
  Text,
  VStack,
} from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { MdOutlineInfo } from "react-icons/md";

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
      <Popover placement='auto' trigger='hover' closeOnBlur>
        <PopoverTrigger>
          <Heading fontSize='xl' mb={4} fontWeight='600' color='purple.800' pos='relative'>
            Warpcast Api Key 
            <Icon as={MdOutlineInfo} color='purple.500' mt={2} ml={5}/>
          </Heading>
        </PopoverTrigger>
        <PopoverContent bg='purple.500' border='none'>
          <PopoverArrow bg='purple.500' />
          <PopoverCloseButton color='white' />
          <PopoverBody color='white' p={5} fontSize='md' fontWeight={'normal'}>
            To unlock features like poll remindersGet your Warpcast API Key from the <Link href="https://warpcast.com/~/developers/api-keys" _hover={{ bg: "white", color: "purple" }} isExternal >official developer portal</Link>.
          </PopoverBody>
        </PopoverContent>
      </Popover>
      <VStack spacing={4} align='stretch'>
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
