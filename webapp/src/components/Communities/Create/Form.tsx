import { Box, Heading, Text, useToast, VStack } from '@chakra-ui/react'
import { MultiValue } from 'chakra-react-select'
import { FormProvider, SubmitHandler, useForm } from 'react-hook-form'
import { CensusSelector } from './Census'
import { Channels } from './Channels'
import { Confirm } from './Confirm'
import { Meta } from './Meta'

export type CommunityFormValues = {
  communityName: string
  admins: MultiValue<{ label: string; value: string }>
  logo: FileList
  channels: MultiValue<{ label: string; value: string }>
}

export const CommunitiesCreateForm = () => {
  const methods = useForm<CommunityFormValues>()
  const toast = useToast()

  const onSubmit: SubmitHandler<CommunityFormValues> = (data) => {
    // Here you will handle the form submission, like sending data to your API
    console.log(data)
    toast({
      title: 'Community created.',
      description: "We've created your community for you.",
      status: 'success',
      duration: 9000,
      isClosable: true,
    })
  }

  return (
    <Box display='flex' flexDir='column' gap={1}>
      <Heading size='md'>Create community</Heading>
      <Text color='gray.400' mb={4}>
        Create your Farcaster.vote community to start managing proposals, creating polls, notify users, etc.
      </Text>
      <FormProvider {...methods}>
        <Box
          as='form'
          onSubmit={methods.handleSubmit(onSubmit)}
          gap={4}
          display='flex'
          flexDir={['column', 'column', 'row']}
          alignItems='start'
        >
          <Box bg='white' p={4} boxShadow='md' borderRadius='md'>
            <VStack spacing={8} alignItems='left'>
              <Meta />
              <CensusSelector />
              <Channels />
            </VStack>
          </Box>
          <Box bg='white' p={4} boxShadow='md' borderRadius='md'>
            <Confirm />
          </Box>
        </Box>
      </FormProvider>
    </Box>
  )
}
