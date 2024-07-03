import { Alert, AlertDescription, AlertIcon, Box, Flex, Heading, Text, VStack } from '@chakra-ui/react'
import { waitForTransactionReceipt, writeContract } from '@wagmi/core'
import { useCallback, useEffect, useState } from 'react'
import { FormProvider, SubmitHandler, useForm } from 'react-hook-form'
import { decodeEventLog } from 'viem'
import { base, baseSepolia, degen } from 'viem/chains'
import { Config, useAccount, useReadContract, useWalletClient, useWatchContractEvent } from 'wagmi'
import { useAuth } from '~components/Auth/useAuth'
import { CensusFormValues } from '~components/CensusTypeSelector'
import { appUrl } from '~constants'
import { communityHubAbi } from '~src/bindings'
import { getContractForChain } from '~util/chain'
import { config } from '~util/rainbow'
import { cleanChannel } from '~util/strings'
import { censusTypeToEnum, ContractCensusType } from '~util/types'
import { ChannelsFormValues, ChannelsSelector } from '../../Census/ChannelsSelector'
import { CensusSelector } from './CensusSelector'
import { Confirm } from './Confirm'
import CommunityDone from './Done'
import { GroupChat } from './GroupChat'
import { CommunityMetaFormValues, Meta } from './Meta'

export type CommunityFormValues = Pick<CensusFormValues, 'addresses' | 'censusType' | 'channel'> &
  CommunityMetaFormValues &
  ChannelsFormValues & {
    enableNotifications: boolean
  }

export const CommunitiesCreateForm = () => {
  const { profile, bfetch } = useAuth()
  const methods = useForm<CommunityFormValues>()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [tx, setTx] = useState<string | null>(null)
  const [cid, setCid] = useState<string | null>(null)
  const { data: walletClient } = useWalletClient()
  const { address, chain } = useAccount()

  useWatchContractEvent({
    abi: communityHubAbi,
    address: getContractForChain(chain),
    chainId: chain?.id,
    eventName: 'CommunityCreated',
    onLogs: (logs) => {
      console.log('logs received:', logs)
    },
    enabled: !!chain?.id,
  })

  const {
    data: price,
    error: priceError,
    isLoading: isLoadingPrice,
  } = useReadContract({
    abi: communityHubAbi,
    address: getContractForChain(chain),
    chainId: chain?.id,
    functionName: 'getCreateCommunityPrice',
    query: {
      enabled: typeof chain?.name !== 'undefined',
    },
  })

  // propagate price error
  useEffect(() => {
    if (!priceError?.message) return

    setError(priceError.message.toString())
  }, [priceError])

  // clear error on chain change
  useEffect(() => {
    setError(null)
  }, [chain])

  const onSubmit: SubmitHandler<CommunityFormValues> = useCallback(
    async (data) => {
      if (loading) return
      setLoading(true)
      setError(null)
      try {
        if (typeof price !== 'bigint') {
          throw Error('Price is not calculated yet')
        }
        const metadata = {
          name: data.name, // name
          imageURI: `${appUrl}/images/avatar/${data.hash}.jpg`, // logo uri
          groupChatURL: data.groupChat ?? '', // groupChatURL
          channels: data.channels?.map((chan) => chan.value) ?? [], // channels
          notifications: true, // notifications
        }

        const cencusType = censusTypeToEnum(data.censusType)
        switch (cencusType) {
          case ContractCensusType.CHANNEL:
            if (!data.channel) throw Error('Channel is not set')
            break
          case ContractCensusType.FOLLOWERS:
            // to include the reference of the user in the contract, we need to
            // add the fid to the channel field in the census metadata with type
            // follower. The prefix fid: is used to identify the field as a
            // farcaster id. It could be used in the future to add more types of
            // followers like alfafrens.
            data.channel = `fid:${profile?.fid}`
            break
          case ContractCensusType.ERC20:
          case ContractCensusType.NFT:
            if (data.addresses?.length === 0) throw Error('Tokens is not set')
            break
          default:
            throw Error('Census type is not allowed')
        }

        const census = {
          censusType: cencusType,
          tokens:
            data.addresses
              ?.filter(({ address }: Address) => address !== '')
              .map(({ blockchain, address: contractAddress }) => {
                return {
                  blockchain,
                  contractAddress: contractAddress as `0x${string}`,
                }
              }) ?? [], // tokens
          channel: data.channel ? cleanChannel(data.channel as string) : '', // channel
        }

        const guardians = data.admins.map((admin) => BigInt(admin.value))
        const createElectionPermission = 0

        console.info('Contract address', getContractForChain(chain))
        console.info('Mapped for contract write:', [
          metadata,
          census,
          guardians,
          createElectionPermission,
          'price: ' + price,
        ])

        const hash = await writeContract(config as Config, {
          abi: communityHubAbi,
          address: getContractForChain(chain),
          chainId: chain?.id as typeof degen.id | typeof baseSepolia.id | typeof base.id,
          functionName: 'createCommunity',
          args: [metadata, census, guardians, createElectionPermission],
          value: price,
        })

        const receipt = await waitForTransactionReceipt(config as Config, {
          hash,
        })

        console.info('transaction hash & receipt:', hash, receipt)

        // Decode logs to find communityId
        const communityId: bigint | undefined = (() => {
          for (const log of receipt.logs) {
            try {
              const decodedLog = decodeEventLog({
                abi: communityHubAbi,
                eventName: 'CommunityCreated',
                data: log.data,
                topics: log.topics,
              })
              if (decodedLog) {
                const { communityId } = decodedLog.args
                setCid(communityId.toString())
                return communityId
              }
            } catch (e) {
              console.error('Error decoding log:', e)
            }
          }
        })()

        if (!communityId) {
          throw new Error('Could not retrieve community id from transaction logs')
        }

        // upload image
        const avatar = {
          communityID: Number(communityId),
          id: data.hash,
          data: data.src,
        }

        await bfetch(`${appUrl}/images/avatar`, {
          method: 'POST',
          body: JSON.stringify(avatar),
        })

        setTx(hash)
      } catch (e) {
        console.error('could not create community:', e)
        if ('shortMessage' in (e as { shortMessage: string })) {
          setError('Could not create community: ' + (e as { shortMessage: string }).shortMessage)
        } else if (e instanceof Error) {
          setError('Could not create community: ' + e.message)
        }
      } finally {
        setLoading(false)
      }
    },
    [walletClient, address, loading, price, profile]
  )

  return (
    <Box display='flex' flexDir='column' gap={1}>
      <FormProvider {...methods}>
        {tx ? (
          <CommunityDone id={cid} tx={tx} />
        ) : (
          <>
            <Heading size='md'>Create community</Heading>
            <Text color='gray.400' mb={4}>
              Create your Farcaster.vote community to start managing proposals, creating polls, notify users, etc.
            </Text>
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
                  <Text fontSize='sm' color={'purple.500'}>
                    Required information
                  </Text>
                  <Meta />
                  <CensusSelector />
                </VStack>
              </Box>
              <Flex direction={'column'} gap={4}>
                <Box bg='white' p={4} boxShadow='md' borderRadius='md'>
                  <VStack spacing={8} alignItems='left'>
                    <Text fontSize='sm' color={'purple.500'}>
                      Social information
                    </Text>
                    <ChannelsSelector />
                    <GroupChat />
                  </VStack>
                </Box>
                <Box bg='white' p={4} boxShadow='md' borderRadius='md'>
                  <Confirm isLoading={loading || isLoadingPrice} price={price} />
                  {error && (
                    <Alert status='error' mt={3}>
                      <AlertIcon />
                      <AlertDescription whiteSpace='collapse' overflowWrap='anywhere' maxW='100%'>
                        {error.toString()}
                      </AlertDescription>
                    </Alert>
                  )}
                </Box>
              </Flex>
            </Box>
          </>
        )}
      </FormProvider>
    </Box>
  )
}
