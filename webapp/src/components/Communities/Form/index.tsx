import { Alert, AlertDescription, AlertIcon, Box, Flex, Heading, Text, VStack } from '@chakra-ui/react'
import { JsonRpcSigner } from 'ethers'
import { useCallback, useEffect, useState } from 'react'
import { FormProvider, SubmitHandler, useForm } from 'react-hook-form'
import { useAccount, useBalance, useWalletClient } from 'wagmi'
import { useAuth } from '~components/Auth/useAuth'
import { CensusFormValues } from '~components/CensusTypeSelector'
import { appUrl, degenContractAddress } from '~constants'
import { CommunityHub__factory, ICommunityHub } from '~typechain'
import { walletClientToSigner } from '~util/rainbow'
import { cleanChannel } from '~util/strings'
import { censusTypeToEnum, ContractCensusType } from '~util/types'
import { CensusSelector } from './CensusSelector'
import { Channels } from './Channels'
import { Confirm } from './Confirm'
import CommunityDone from './Done'
import { GroupChat } from './GroupChat'
import { CommunityMetaFormValues, Meta } from './Meta'

export type CommunityFormValues = Pick<CensusFormValues, 'addresses' | 'censusType' | 'channel'> &
  CommunityMetaFormValues & {
    channels: { label: string; value: string }[]
    enableNotifications: boolean
  }

export const CommunitiesCreateForm = () => {
  const { profile, bfetch } = useAuth()
  const methods = useForm<CommunityFormValues>()
  const [isPending, setIsPending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [tx, setTx] = useState<string | null>(null)
  const [cid, setCid] = useState<string | null>(null)
  const { data: walletClient } = useWalletClient()
  const { address, chainId } = useAccount()
  const [isLoadingPrice, setIsLoadingPrice] = useState(false)
  const [price, setPrice] = useState<bigint | null>()

  const { data: balanceResult, isLoading: isBalanceLoading, error: balanceError } = useBalance({ address })
  const [userBalance, setUserBalance] = useState<string | null>(null)

  useEffect(() => {
    if (!walletClient || !address) {
      return
    }
    ;(async () => {
      try {
        setIsLoadingPrice(true)
        let signer: JsonRpcSigner | undefined
        if (walletClient && address && walletClient.account.address === address) {
          signer = await walletClientToSigner(walletClient)
        }
        if (!signer) throw Error("Can't get the signer")

        if (!isBalanceLoading) {
          setUserBalance(balanceResult ? (Number(balanceResult.value) / 10 ** balanceResult.decimals).toString() : '0')
        }
        const communityHubContract = CommunityHub__factory.connect(degenContractAddress, signer)

        const price = await communityHubContract.getCreateCommunityPrice()

        setPrice(price)
      } catch (e) {
        console.error('could not create community:', e)
      } finally {
        setIsLoadingPrice(false)
      }
    })()
  }, [walletClient, chainId, address, isBalanceLoading, balanceResult, balanceError])

  const onSubmit: SubmitHandler<CommunityFormValues> = useCallback(
    async (data) => {
      if (isPending) return
      setError(null)
      try {
        if (!price) throw Error('Price is not calculated yet')
        setIsPending(true)
        const metadata: ICommunityHub.CommunityMetadataStruct = {
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
          case ContractCensusType.ERC20, ContractCensusType.NFT:
            if (data.addresses?.length === 0) throw Error('Tokens is not set')
            break
          default:
            throw Error('Census type is not allowed')
        }

        const census: ICommunityHub.CensusStruct = {
          censusType: cencusType,
          tokens:
            data.addresses
              ?.filter(({ address }) => address !== '')
              .map(({ blockchain, address: contractAddress }) => {
                return {
                  blockchain,
                  contractAddress,
                } as ICommunityHub.TokenStruct
              }) ?? ([] as ICommunityHub.TokenStruct[]), // tokens
          channel: data.channel ? cleanChannel(data.channel) : '', // channel
        }

        console.log(census)

        const guardians = data.admins.map((admin) => BigInt(admin.value))
        const createElectionPermission = 0

        console.info('Degen contract address', degenContractAddress)
        console.info('mapped for contract write:', [
          metadata,
          census,
          guardians,
          createElectionPermission,
          'price: ' + price,
        ])

        let signer: JsonRpcSigner | undefined
        if (walletClient && address && walletClient.account.address === address) {
          signer = await walletClientToSigner(walletClient)
        }
        if (!signer) throw Error("Can't get the signer")

        const communityHubContract = CommunityHub__factory.connect(degenContractAddress, signer)

        const tx = await communityHubContract.createCommunity(metadata, census, guardians, createElectionPermission, {
          value: price,
        })

        const receipt = await tx.wait()
        if (!receipt) {
          throw Error('Cannot get receipt')
        }

        const communityHubInterface = CommunityHub__factory.createInterface()
        // get just created community id
        const { communityID } = communityHubInterface.decodeEventLog('CommunityCreated', tx.data) as unknown as {
          communityID: string
        }
        // upload image
        const avatar = {
          communityID,
          id: data.hash,
          data: data.src,
        }

        await bfetch(`${appUrl}/images/avatar`, {
          method: 'POST',
          body: JSON.stringify(avatar),
        })

        setTx(tx.hash)
        setCid(communityID)
      } catch (e) {
        console.error('could not create community:', e)
        if ('shortMessage' in (e as { shortMessage: string })) {
          setError('Could not create community: ' + (e as { shortMessage: string }).shortMessage)
        } else if (e instanceof Error) {
          setError('Could not create community: ' + e.message)
        }
      } finally {
        setIsPending(false)
      }
    },
    [walletClient, address, isPending, price, profile]
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
                    <Channels />
                    <GroupChat />
                  </VStack>
                </Box>
                <Box bg='white' p={4} boxShadow='md' borderRadius='md'>
                  <Confirm
                    isLoading={isPending || isLoadingPrice || isBalanceLoading}
                    price={price}
                    balance={userBalance as string}
                  />
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
