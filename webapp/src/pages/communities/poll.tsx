import { ethers } from 'ethers'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'
import { PollView } from '~components/Poll'
import { degenChainRpc, degenContractAddress } from '~constants'
import { fetchPollInfo, fetchPollsVoters } from '~queries/polls'
import { CommunityHub__factory } from '~typechain'
import { toArrayBuffer } from '~util/hex'

const CommunityPoll = () => {
  const { pid: electionId, id: communityId } = useParams()
  const { bfetch } = useAuth()

  const [loaded, setLoaded] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState<boolean>(false)
  const [pollResults, setResults] = useState<PollInfo | null>(null)
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
        console.log('received contract data:', contractData.options)
        let voteCount = 0
        if (contractData.date !== '') {
          const participants = contractData.participants.map((p) => parseInt(p.toString()))
          const tally = contractData.tally.map((t) => t.map((v) => parseInt(v.toString())))
          const date = new Date(contractData.date.replace(/[UTC|CEST]+ m=\+[\d.]+$/, ''))

          setResults({
            ...contractData,
            electionId,
            participants: participants,
            tally: tally,
            turnout: parseFloat(contractData.turnout.toString()),
            voteCount: contractData.participants.length,
            finalized: true,
            endTime: date,
            // although it is already setted, we need to set it again to avoid type issues
            options: contractData.options,
            // TODO: get this from the contract or api
            totalWeight: contractData.participants.reduce((acc, p) => acc + parseInt(p.toString()), 0),
            lastVoteTime: date,
            createdByUsername: '',
            createdByDisplayname: '',
            createdTime: date,
            censusParticipantsCount: Number(contractData.totalVotingPower),
          })
          voteCount = contractData.participants.length
          console.info('results gathered from contract')
        } else {
          const apiData = await fetchPollInfo(bfetch, electionId)()
          const tally: number[][] = [[]]
          apiData.tally?.forEach((t) => {
            tally[0].push(parseInt(t))
          })
          setResults({
            ...apiData,
            totalWeight: Number(apiData.totalWeight),
            endTime: new Date(apiData.endTime),
            createdTime: new Date(apiData.createdTime),
            lastVoteTime: new Date(apiData.lastVoteTime),
            tally: tally,
          })
          voteCount = apiData.voteCount
          console.info('results gathered from api')
        }
        // get voters
        if (voteCount > 0) {
          try {
            setVoters(await fetchPollsVoters(bfetch, electionId)())
          } catch (e) {
            console.error('error fetching voters', e)
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
