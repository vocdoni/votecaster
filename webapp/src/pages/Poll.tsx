import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'

import { useAuth } from '../components/Auth/useAuth'
import { PollView } from '../components/Poll'
import type { PollResult } from '../util/types'
import { fetchPollInfo } from '../queries/polls'

const Poll = () => {
  const { pid: electionId } = useParams()
  const { bfetch } = useAuth()

  const [loaded, setLoaded] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState<boolean>(false)
  const [pollResults, setResults] = useState<PollResult | null>(null)

  useEffect(() => {
    if (loaded || loading || !electionId ) return
      ; (async () => {
        try {
          setLoading(true)
            const apiData = await fetchPollInfo(bfetch)(electionId)
            const tally: number[][] = [[]]
            apiData.tally?.forEach((t) => {
              tally[0].push(parseInt(t))
            })
            setResults({
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
            })
        } catch (e) {
          setError("Error fetching poll results")
          console.error(e)
        } finally {
          setLoaded(true)
          setLoading(false)
        }
      })()
  }, [])

  return <PollView 
          loaded={loaded}
          onChain={false}
          loading={loading}
          poll={pollResults}
          errorMessage={error}
          electionId={electionId} />
}

export default Poll
