import { useParams } from 'react-router-dom'
import { merge } from 'ts-deepmerge'
import { Check } from '~components/Check'
import { PollView } from '~components/Poll'
import { useApiPollInfo, useContractPollInfo } from '~queries/polls'
import { contractDataToObject } from '~util/mappings'

const CommunityPoll = () => {
  const { chain, community, poll } = useParams<CommunityPollParams>()
  const { data, isLoading, error } = useContractPollInfo(chain as ChainKey, Number(community), poll as string)
  const api = useApiPollInfo(poll as string)

  // Merge contract and API data
  let results = {
    ...api.data,
  } as PollInfo

  if (data?.date) {
    results = merge.withOptions(
      {
        mergeArrays: false,
      },
      results,
      contractDataToObject(data)
    ) as PollInfo
  }
  return (
    <>
      <Check isLoading={api.isLoading || isLoading} error={api.error || error} />
      {results.electionId && <PollView loading={api.isLoading && isLoading} poll={results} />}
    </>
  )
}

export default CommunityPoll
