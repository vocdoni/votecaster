
import { ethers } from 'ethers'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useAuth } from '../../components/Auth/useAuth'
import { CommunityPollView } from '../../components/Communities/Poll'
import type { PollResult } from '../../util/types'

import { toArrayBuffer } from '../../util/hex'
import { CommunityHub__factory } from '../../typechain'
import { degenChainRpc, degenContractAddress } from '../../util/constants'
import { fetchPollInfo } from '../../queries/polls'

const mockedResults: PollResult = {
  censusRoot: 'a989f2e94f9f7954c96ba2cef784525c5ce5c3cba90f0b3da14349a93f3e7dde',
  censusURI: 'https://census.com',
  endTime: new Date("2024-04-20T14:28:51.228+00:00"),
  options: ['Option 1', 'Option 2'],
  participants: [237855, 308972, 10080],
  question: 'Whats your favorite love movie?',
  tally: [[1, 2], [], [], []],
  voteCount: 3,
  turnout: 100,
  finalized: true,
  censusParticipantsCount: 3,
}

const CommunityPoll = () => {
  const { pid: electionId, id: communityId } = useParams()
  const { bfetch } = useAuth()

  const [loaded, setLoaded] = useState<boolean>(false)
  const [loading, setLoading] = useState<boolean>(false)
  const [pollResults, setResults] = useState<PollResult | null>(null)

  useEffect(() => {
    if (loaded || loading || !electionId || !communityId) return
      ; (async () => {
        try {
          setLoading(true)
          // get results from the contract
          const provider = new ethers.JsonRpcProvider(degenChainRpc)
          const communityHubContract = CommunityHub__factory.connect(degenContractAddress, provider)
          const contractData = await communityHubContract.getResult(communityId, toArrayBuffer(electionId))
          let results: PollResult
          if (contractData.date !== "") {
            const participants = contractData.participants.map((p) => parseInt(p.toString()))
            const tally = contractData.tally.map((t) => t.map((v) => parseInt(v.toString())))
            results = {
              censusRoot: contractData.censusRoot,
              censusURI: contractData.censusURI,
              endTime: new Date(contractData.date.replace(/m=\+[\d.]+$/, '')),
              options: contractData.options,
              participants: participants,
              question: contractData.question,
              tally: tally,
              turnout: parseFloat(contractData.turnout.toString()),
              voteCount: contractData.participants.length,
              finalized: true,
              censusParticipantsCount: Number(contractData.totalVotingPower), // TODO: get this from the contract or api
            }
            console.log("results from contract")
          } else {
            try {
              const apiData = await fetchPollInfo(bfetch)(electionId)
              const tally: number[][] = [[]]
              apiData.tally?.forEach((t) => {
                tally[0].push(parseInt(t))
              })
              results = {
                censusRoot: "",
                censusURI: "",
                endTime: new Date(apiData.endTime),
                options: apiData.options,
                participants: apiData.participants,
                question: apiData.question,
                tally: tally,
                turnout: apiData.turnout,
                voteCount: apiData.voteCount,
                finalized: apiData.finalized,
                censusParticipantsCount: apiData.censusParticipantsCount,
              }
              console.log("results from api")
            } catch (e) {
              console.error(e)
              results = mockedResults
              console.log("mocked results")
            }
          }
          setResults(results)
        } catch (e) {
          console.error(e)
        } finally {
          setLoaded(true)
          setLoading(false)
        }
      })()
  }, [])

  return <CommunityPollView 
          loaded={loaded}
          loading={loading}
          poll={pollResults} 
          electionId={electionId} 
          communityId={communityId}/>
}

export default CommunityPoll
