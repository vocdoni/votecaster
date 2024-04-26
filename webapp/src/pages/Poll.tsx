import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'
import { PollView } from '~components/Poll'
import { fetchPollInfo, fetchPollsVoters } from '~queries/polls'

const Poll = () => {
  const { pid: electionId } = useParams()
  const { bfetch } = useAuth()

  const [loaded, setLoaded] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState<boolean>(false)
  const [pollResults, setResults] = useState<PollResult | null>(null)
  const [voters, setVoters] = useState<string[]>([])

  useEffect(() => {
    if (loaded || loading || !electionId) return
    ;(async () => {
      try {
        setLoading(true)
        const apiData = await fetchPollInfo(bfetch)(electionId)
        const tally: number[][] = [[]]
        apiData.tally?.forEach((t) => {
          tally[0].push(parseInt(t))
        })
        setResults({
          ...apiData,
          censusRoot: '',
          censusURI: '',
          endTime: new Date(apiData.endTime),
          tally: tally,
        })
        // get voters
        if (apiData.voteCount > 0) {
          try {
            setVoters(await fetchPollsVoters(bfetch)(electionId))
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
      loaded={loaded}
      onChain={false}
      loading={loading}
      poll={pollResults}
      voters={voters}
      errorMessage={error}
      electionId={electionId}
    />
  )
}

export default Poll
