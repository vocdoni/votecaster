import { Alert, AlertDescription, Box, Flex, Heading, Text, VStack } from '@chakra-ui/react'
import { id } from '@ethersproject/hash'
import { BrowserProvider, ContractTransactionReceipt } from 'ethers'
import { useCallback, useEffect, useState } from 'react'
import { FormProvider, SubmitHandler, useForm } from 'react-hook-form'
import { useAccount, useBalance, useWalletClient, type UseWalletClientReturnType } from 'wagmi'
import { CensusFormValues } from '~components/CensusTypeSelector'
import { degenContractAddress } from '~constants'
import { CommunityHub__factory, ICommunityHub } from '~typechain'
import { CommunityHubInterface } from '~typechain/src/CommunityHub'
import { censusTypeToEnum } from '~util/types'
import { CensusSelector } from './CensusSelector'
import { Channels } from './Channels'
import { Confirm } from './Confirm'
import CommunityDone from './Done'
import { GroupChat } from './GroupChat'
import { CommunityMetaFormValues, Meta } from './Meta'

export type CommunityFormValues = Pick<CensusFormValues, 'addresses' | 'censusType' | 'channel'> &
  CommunityMetaFormValues & {
    channels: { label: string; value: string }[]
    enableNotifications: boolean // todo(kon): not for mvp
  }

export const CommunitiesCreateForm = () => {
  const methods = useForm<CommunityFormValues>()
  const [isPending, setIsPending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [tx, setTx] = useState<string | null>(null)
  const { data: walletClient } = useWalletClient()
  const { address } = useAccount()
  const [isLoadingPrice, setIsLoadingPrice] = useState(false)
  const [price, setPrice] = useState<bigint | null>()
  const calcPrice = price ? (Number(price) / 10 ** 18).toString() : ''

  const { data: balanceResult, isLoading: isBalanceLoading, error: balanceError } = useBalance({ address })
  const [userBalance, setUserBalance] = useState<string | null>(null)

  useEffect(() => {
    if (!walletClient || !address) {
      return
    }
    ;(async () => {
      try {
        setIsLoadingPrice(true)
        // todo(kon): put this code on a provider and get the contract instance from there
        let signer: any
        if (walletClient && address && walletClient.account.address === address) {
          signer = await walletClientToSigner(walletClient)
        }
        if (!signer) throw Error("Can't get the signer")

        if (!isBalanceLoading) {
          setUserBalance(balanceResult ? (Number(balanceResult.value) / 10 ** balanceResult.decimals).toString() : '0')
        }
        const communityHubContract = CommunityHub__factory.connect(degenContractAddress, signer)

        // todo(kon): move this to a reactQuery?
        const price = await communityHubContract.getCreateCommunityPrice()

        setPrice(price)
      } catch (e) {
        console.error('could not create community:', e)
      } finally {
        setIsLoadingPrice(false)
      }
    })()
  }, [walletClient, address, isBalanceLoading, balanceResult, balanceError])

  const onSubmit: SubmitHandler<CommunityFormValues> = useCallback(
    async (data) => {
      if (isPending) return
      setError(null)
      try {
        if (!price) throw Error('Price is not calculated yet')
        setIsPending(true)
        const metadata: ICommunityHub.CommunityMetadataStruct = {
          name: data.name, // name
          imageURI: data.logo, // logo uri
          groupChatURL: data.groupChat ?? '', // groupChatURL
          channels: data.channels?.map((chan) => chan.value) ?? [], // channels
          notifications: true, // notifications
        }

        const census: ICommunityHub.CensusStruct = {
          censusType: censusTypeToEnum(data.censusType), // Census type
          tokens:
            data.addresses
              ?.filter(({ address }) => address !== '')
              .map(({ blockchain, address: contractAddress }) => {
                return {
                  blockchain,
                  contractAddress,
                } as ICommunityHub.TokenStruct
              }) ?? ([] as ICommunityHub.TokenStruct[]), // tokens
          channel: data.channel ?? '', // channel
        }

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

        // todo(kon): put this code on a provider and get the contract instance from there
        let signer: any
        if (walletClient && address && walletClient.account.address === address) {
          signer = await walletClientToSigner(walletClient)
        }
        if (!signer) throw Error("Can't get the signer")

        const communityHubContract = CommunityHub__factory.connect(degenContractAddress, signer)

        // todo(kon): can this be moved to a reactQuery?
        const tx = await communityHubContract.createCommunity(metadata, census, guardians, createElectionPermission, {
          value: price,
        })

        const receipt = await tx.wait()

        if (!receipt) {
          throw Error('Cannot get receipt')
        }

        const communityHubInterface = CommunityHub__factory.createInterface()
        const log = findLog(receipt, communityHubInterface)
        if (!log) {
          throw Error('Cannot get community log')
        }
        // this stop working after last contract update
        // const parsedLog = communityHubInterface.parseLog(log)
        // const communityId = parsedLog?.args['communityId']
        // if (!communityId) {
        //   throw Error("Cannot get community id")
        // }
        // console.log("Commnuity id found", communityId, tx.hash)
        // setCommunityId(communityId)

        setTx(tx.hash)
      } catch (e) {
        console.error('could not create community:', e)
        if ('shortMessage' in (e as any)) {
          setError('Could not create community: ' + (e as any).shortMessage)
        } else if (e instanceof Error) {
          setError('Could not create community: ' + e.message)
        }
      } finally {
        setIsPending(false)
      }
    },
    [walletClient, address, isPending, price]
  )

  return (
    <Box display='flex' flexDir='column' gap={1}>
      <FormProvider {...methods}>
        {tx ? (
          <CommunityDone tx={tx} />
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
                    price={calcPrice}
                    balance={userBalance as string}
                  />
                  {error && (
                    <Alert status='error' mt={3}>
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

export async function walletClientToSigner(walletClient: UseWalletClientReturnType['data']) {
  const { account, chain, transport } = walletClient!
  const network = {
    chainId: chain.id,
    name: chain.name,
    ensAddress: chain.contracts?.ensRegistry?.address,
  }
  const provider = new BrowserProvider(transport, network)
  const signer = await provider.getSigner(account.address)
  return signer
}

export function findLog(receipt: ContractTransactionReceipt, iface: CommunityHubInterface) {
  return receipt.logs.find((log) => {
    return log.topics[0] === id(iface.getEvent('CommunityCreated').format('sighash'))
  })
}
