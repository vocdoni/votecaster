import { ethers } from 'ethers'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'
import { Check } from '~components/Check'
import { PollView } from '~components/Poll'
import { degenChainRpc, degenContractAddress } from '~constants'
import { fetchPollInfo } from '~queries/polls'
import { CommunityHub__factory } from '~typechain'
import { toArrayBuffer } from '~util/hex'

const CommunityPoll = () => {
  const { pid: electionId, id: communityId } = useParams()
  const { bfetch } = useAuth()

  const [loaded, setLoaded] = useState<boolean>(false)
  const [error, setError] = useState<Error | null>(null)
  const [loading, setLoading] = useState<boolean>(false)
  const [pollResults, setResults] = useState<PollInfo | null>(null)

  useEffect(() => {
    if (loaded || loading || !electionId || !communityId) return
    ;(async () => {
      try {
        setLoading(true)
        // get results from the contract
        const provider = new ethers.JsonRpcProvider(degenChainRpc)
        const communityHubContract = CommunityHub__factory.connect(degenContractAddress, provider)
        const contractData = await communityHubContract.getResult(communityId, toArrayBuffer(electionId))
        console.info('received contract data:', contractData.options)
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
            totalWeight: contractData.participants.reduce((acc, p) => acc + parseInt(p.toString()), 0),
            // TODO: get this from the contract or api
            lastVoteTime: date,
            createdByUsername: '',
            createdByDisplayname: '',
            createdTime: date,
            censusParticipantsCount: 0,
          })
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
          console.info('results gathered from api')
        }
      } catch (e) {
        setError(new Error('Error fetching poll results'))
        console.error(e)
      } finally {
        setLoaded(true)
        setLoading(false)
      }
    })()
  }, [])

  if (error || loading) {
    return <Check error={error} isLoading={loading} />
  }

  if (!pollResults) {
    return <Check error={new Error('No poll results found')} isLoading={false} />
  }

  return <PollView loading={loading} onChain={true} poll={pollResults} />
}

export default CommunityPoll
