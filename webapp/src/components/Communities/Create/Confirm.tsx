import {Box, Button, ButtonProps, Heading, Text} from '@chakra-ui/react'
import {useConnectModal} from '@rainbow-me/rainbowkit'
import {CiWallet} from 'react-icons/ci'
import {GiTopHat} from 'react-icons/gi'
import {MdOutlineRocketLaunch} from 'react-icons/md'
import {useAccount, useWalletClient} from 'wagmi'
import {SubmitHandler} from "react-hook-form";
import {useCallback, useEffect, useState} from "react";
import {ICommunityHub} from "../../../typechain/src/CommunityHub.ts";
import {censusTypeToEnum} from "../../../util/types.ts";
import {degenContractAddress, electionResultsContract} from "../../../util/constants.ts";
import {CommunityHub__factory} from "../../../typechain";
import {CommunityFormValues, findLog, walletClientToSigner} from "./Form.tsx";

export const Confirm = (props: ButtonProps) => {
  const {isConnected} = useAccount()
  const [isLoading, setIsLoading] = useState(false)
  const [price, setPrice] = useState<string | null>()
  const {openConnectModal} = useConnectModal()

  const {data: walletClient} = useWalletClient()
  const {address} = useAccount()

  useEffect(() => {
    if (!walletClient || !address) return
      ;
    (async () => {
      try {
        setIsLoading(true)
        // todo(kon): put this code on a provider and get the contract instance from there
        let signer: any
        if (walletClient && address && walletClient.account.address === address) {
          signer = await walletClientToSigner(walletClient)
        }
        if (!signer) throw Error("Can't get the signer")

        const communityHubContract = CommunityHub__factory.connect(degenContractAddress, signer)

        // todo(kon): move this to a reactQuery?
        const price = await communityHubContract.getCreateCommunityPrice()

        setPrice(price.toString())
      } catch (e) {
        console.error('could not create community:', e)
      } finally {
        setIsLoading(false)
      }
    })();
  }, [walletClient, address])

  return (
    <Box display='flex' gap={4} flexDir='column'>
      <Heading size='sm'>Create your community</Heading>
      <Text>Your community will be deployed on the Degenchain.</Text>
      <Text>
        As soon as it's created, you will be able to create and manage polls secured by the Vocdoni protocol for
        decentralized, censorship-resistant and gassless voting.
      </Text>
      {!!price && <Box display='flex' justifyContent='space-between' fontWeight='500' w='full'>
        <Text>Cost</Text>
        <Text>{price} $DEGEN</Text>
      </Box>}
      {isConnected ? (
        <Button mt={4} colorScheme='blue' type='submit' rightIcon={<GiTopHat/>}
                leftIcon={<MdOutlineRocketLaunch/>} {...props}>
          Deploy your community on Degenchain
        </Button>
      ) : (
        <Button onClick={openConnectModal} colorScheme='blue' leftIcon={<CiWallet/>} {...props}>
          Connect wallet first
        </Button>
      )}
    </Box>
  )
}
