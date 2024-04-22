import {Alert, Box, Heading, Text, VStack, AlertDescription, Flex} from '@chakra-ui/react'
import {FormProvider, SubmitHandler, useForm} from 'react-hook-form'
import {useAccount, useWalletClient, type UseWalletClientReturnType} from 'wagmi'
import {degenContractAddress, electionResultsContract} from '../../../util/constants'
import {CensusFormValues} from '../../CensusTypeSelector'
import {CensusSelector} from './CensusSelector'
import {Channels} from './Channels'
import {Confirm} from './Confirm'
import {CommunityMetaFormValues, Meta} from './Meta'
import {censusTypeToEnum} from "../../../util/types.ts";
import {useCallback, useState} from "react";
import {CommunityHub__factory} from '../../../typechain'
import {CommunityHubInterface, ICommunityHub} from "../../../typechain/src/CommunityHub.ts";
import {BrowserProvider} from "ethers";
import {id} from "@ethersproject/hash";
import {ContractTransactionReceipt} from "ethers";
import {GroupChat} from "./GroupChat.tsx";
import CommunityDone from "./Done.tsx";

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
  const [communityId, setCommunityId] = useState<string | null>(null)
  const {data: walletClient} = useWalletClient()
  const {address} = useAccount()

  const onSubmit: SubmitHandler<CommunityFormValues> = useCallback(async (data) => {
    if (isPending) return;
    setError(null)
    try {
      setIsPending(true)
      console.info('received form data:', data)

      const metadata: ICommunityHub.CommunityMetadataStruct = {
        name: data.name, // name
        imageURI: data.logo, // logo uri
        groupChatURL: data.groupChat ?? '', // groupChatURL
        channels: data.channels.map((chan) => chan.value) ?? [],  // channels
        notifications: true // notifications
      }

      const census: ICommunityHub.CensusStruct = {
        censusType: censusTypeToEnum(data.censusType), // Census type
        tokens: data.addresses?.filter(({_, address}) => address !== '')
          .map(({blockchain, address}) => [blockchain, address]) ?? [], // tokens
        channel: data.channel ?? '' // channel
      }

      const guardians = data.admins.map((admin) => admin.value)
      const createElectionPermission = BigInt(0)

      console.info('Degen contract address', degenContractAddress)
      console.info('mapped for contract write:', [
        metadata,
        census,
        guardians,
        electionResultsContract,
        createElectionPermission,
      ])

      // todo(kon): put this code on a provider and get the contract instance from there
      let signer: any
      if (walletClient && address && walletClient.account.address === address) {
        signer = await walletClientToSigner(walletClient)
      }
      if (!signer) throw Error("Can't get the signer")

      const communityHubContract = CommunityHub__factory.connect(degenContractAddress, signer)

      // todo(kon): can this be moved to a reactQuery?
      const tx = await communityHubContract.createCommunity(
        metadata, census, guardians, electionResultsContract, createElectionPermission, {value: BigInt("100000000000000000000")})

      const receipt = await tx.wait()

      if (!receipt) {
        throw Error("Cannot get receipt")
      }

      const communityHubInterface = CommunityHub__factory.createInterface();
      const log = findLog(
        receipt,
        communityHubInterface,
      );
      if (!log) {
        throw Error("Cannot get community log")
      }
      const parsedLog = communityHubInterface.parseLog(log)
      const communityId = parsedLog?.args['communityId']
      if (!communityId) {
        throw Error("Cannot get community id")
      }
      console.log("Commnuity id found", communityId, tx.hash)
      setTx(tx.hash)
      setCommunityId(communityId)
    } catch (e) {
      console.error('could not create community:', e)
      if (e instanceof Error) {
        setError('could not create community: ' + e.message)
      }
    } finally {
      setIsPending(false)

    }
  }, [walletClient, address, isPending])

  return (
    <Box display='flex' flexDir='column' gap={1}>
      <FormProvider {...methods}>
        {tx ? (<CommunityDone tx={tx}/>) : (
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
                  <Meta/>
                  <CensusSelector/>
                  <Channels/>
                </VStack>
              </Box>
              <Flex direction={'column'} gap={4}>
                <Box bg='white' p={4} boxShadow='md' borderRadius='md'>
                  <GroupChat/>
                </Box>
                <Box bg='white' p={4} boxShadow='md' borderRadius='md'>
                  <Confirm isLoading={isPending}/>
                </Box>
              </Flex>
            </Box>
            {error && <Alert maxW={'90vw'} status='error'>
              <AlertDescription whiteSpace="nowrap" overflow="hidden" textOverflow="ellipsis"
                                isTruncated>{error}</AlertDescription></Alert>}
          </>
        )}
      </FormProvider>
    </Box>
  )
}

export async function walletClientToSigner(walletClient: UseWalletClientReturnType['data']) {
  const {account, chain, transport} = walletClient!
  const network = {
    chainId: chain.id,
    name: chain.name,
    ensAddress: chain.contracts?.ensRegistry?.address,
  }
  const provider = new BrowserProvider(transport, network)
  const signer = await provider.getSigner(account.address)
  return signer
}

export function findLog(
  receipt: ContractTransactionReceipt,
  iface: CommunityHubInterface,
) {
  return receipt.logs.find(
    (log) => {
      return log.topics[0] ===
        id(
          iface.getEvent("CommunityCreated").format(
            "sighash",
          ),
        );
    }
  );
}