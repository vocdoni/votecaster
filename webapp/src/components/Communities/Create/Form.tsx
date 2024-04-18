import {Alert, Box, Heading, Text, VStack} from '@chakra-ui/react'
import {FormProvider, SubmitHandler, useForm} from 'react-hook-form'
import {useWriteContract} from 'wagmi'
import {abi} from '../../../abi.json'
import {degenContractAddress, electionResultsContract} from '../../../util/constants'
import {CensusFormValues} from '../../CensusTypeSelector'
import {CensusSelector} from './CensusSelector'
import {Channels} from './Channels'
import {Confirm} from './Confirm'
import {CommunityMetaFormValues, Meta} from './Meta'
import {censusTypeToEnum} from "../../../util/types.ts";
import {useEffect} from "react";


export type CommunityFormValues = Pick<CensusFormValues, 'addresses' | 'censusType' | 'channel'> &
  CommunityMetaFormValues & {
  channels: { label: string; value: string }[]
}

export const CommunitiesCreateForm = () => {
  const methods = useForm<CommunityFormValues>()
  const {data: hash, isPending, writeContract, error} = useWriteContract()

  const onSubmit: SubmitHandler<CommunityFormValues> = async (data) => {
    try {
      console.info('received form data:', data)


      const metadata = [
        data.name, // name
        data.logo, // logo uri
        "https://t.me/nothing", // groupChatURL
        data.channels.map((chan) => chan.value) ?? [],  // channels
        false // notifications
      ]

      const census = [
        censusTypeToEnum(data.censusType), // Census type
        data.addresses?.filter(({_, address}) => address !== '')
          .map(({blockchain, address}) => [blockchain, address]), // tokens
        data.channel ?? '' // channel
      ]

      const guardians = data.admins.map((admin) => admin.value)
      const createElectionPermission = false

      console.info('Degen contract address', degenContractAddress)
      console.info('mapped for contract write:', [
        metadata,
        census,
        guardians,
        electionResultsContract,
        createElectionPermission,
      ])

      writeContract({
        address: degenContractAddress,
        abi,
        functionName: 'createCommunity',
        args: [metadata, census, guardians, electionResultsContract, createElectionPermission],
      })
    } catch (e) {
      console.error('could not create community:', e)
    }
  }

  useEffect(() => {
    console.info("TX hash:", hash)
  }, [hash])

  if (error) {
    console.error('error creating community:', error)
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
              <Meta/>
              <CensusSelector/>
              <Channels/>
            </VStack>
          </Box>
          <Box bg='white' p={4} boxShadow='md' borderRadius='md'>
            <Confirm isLoading={isPending}/>
          </Box>
        </Box>
        {error && <Alert status='error'>{error.message}</Alert>}
      </FormProvider>
    </Box>
  )
}
