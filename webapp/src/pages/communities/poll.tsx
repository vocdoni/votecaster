import { ethers } from 'ethers'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useAuth } from '../../components/Auth/useAuth'
import { PollView } from '../../components/Poll'
import { fetchPollInfo, fetchPollsVoters } from '../../queries/polls'
import { CommunityHub__factory } from '../../typechain'
import { degenChainRpc, degenContractAddress } from '../../util/constants'
import { toArrayBuffer } from '../../util/hex'
import type { PollResult } from '../../util/types'

const CommunityPoll = () => {
  const { pid: electionId, id: communityId } = useParams()
  const { bfetch } = useAuth()

  const [loaded, setLoaded] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState<boolean>(false)
  const [pollResults, setResults] = useState<PollResult | null>(null)
  const [voters, setVoters] = useState<string[]>([])

  useEffect(() => {
    if (loaded || loading || !electionId || !communityId) return
    ;(async () => {
      try {
        setLoading(true)
        // get results from the contract
        const provider = new ethers.JsonRpcProvider(degenChainRpc)
        const communityHubContract = CommunityHub__factory.connect(degenContractAddress, provider)
        const contractData = await communityHubContract.getResult(communityId, toArrayBuffer(electionId))
        let voteCount = 0
        if (contractData.date !== '') {
          const participants = contractData.participants.map((p) => parseInt(p.toString()))
          const tally = contractData.tally.map((t) => t.map((v) => parseInt(v.toString())))
          setResults({
            censusRoot: contractData.censusRoot,
            censusURI: contractData.censusURI,
            endTime: new Date(contractData.date.replace(/[UTC|CEST]+ m=\+[\d.]+$/, '')),
            options: contractData.options,
            participants: participants,
            question: contractData.question,
            tally: tally,
            turnout: parseFloat(contractData.turnout.toString()),
            voteCount: contractData.participants.length,
            finalized: true,
            censusParticipantsCount: Number(contractData.totalVotingPower), // TODO: get this from the contract or api
          })
          voteCount = contractData.participants.length
          console.log('results from contract')
        } else {
          const apiData = await fetchPollInfo(bfetch)(electionId)
          const tally: number[][] = [[]]
          apiData.tally?.forEach((t) => {
            tally[0].push(parseInt(t))
          })
          setResults({
            censusRoot: '',
            censusURI: '',
            endTime: new Date(apiData.endTime),
            options: apiData.options,
            participants: apiData.participants,
            question: apiData.question,
            tally: tally,
            turnout: apiData.turnout,
            voteCount: apiData.voteCount,
            finalized: apiData.finalized,
            censusParticipantsCount: apiData.censusParticipantsCount,
          })
          voteCount = apiData.voteCount
          console.log('results from api')
        }
        // get voters
        if (voteCount > 0) {
          try {
            setVoters(await fetchPollsVoters(bfetch)(electionId))
          } catch (e) {
            console.log('error fetching voters', e)
          }
        }
      } catch (e) {
        setError('Error fetching poll results')
        console.error(e)
      } finally {
        setLoaded(true)
        setLoading(false)
      }
    })()
  }, [])

  return (
    <PollView
      loaded={loaded}
      loading={loading}
      onChain={true}
      poll={pollResults}
      voters={voters}
      errorMessage={error}
      electionId={electionId}
    />
  )
}

export default CommunityPoll
