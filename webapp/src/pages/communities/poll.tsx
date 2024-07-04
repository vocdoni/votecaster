import { useParams } from 'react-router-dom'
import { merge } from 'ts-deepmerge'
import { Check } from '~components/Check'
import { PollView } from '~components/Poll'
import { useApiPollInfo, useContractPollInfo } from '~queries/polls'
import { contractDataToObject } from '~util/mappings'

const CommunityPoll = () => {
  const { electionId, communityId, chainAlias } = useParams<CommunityPollParams>()
  const { data, isLoading, error } = useContractPollInfo(
    chainAlias as ChainKey,
    communityId as string,
    electionId as string
  )
  const api = useApiPollInfo(electionId as string)

  // Merge contract and API data
  let results = {
    electionId,
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
      <PollView loading={api.isLoading && isLoading} onChain={!!data} poll={results} />
    </>
  )
}

export default CommunityPoll
