import { useParams } from 'react-router-dom'
import { merge } from 'ts-deepmerge'
import { Check } from '~components/Check'
import { useDegenHealthcheck } from '~components/Healthcheck/use-healthcheck'
import { PollView } from '~components/Poll'
import { useApiPollInfo, useContractPollInfo } from '~queries/polls'
import { contractDataToObject } from '~util/mappings'

const CommunityPoll = () => {
  const { pid: electionId, id: communityId, chain: chainAlias } = useParams()
  const { connected } = useDegenHealthcheck()
  const { data } = useContractPollInfo(chainAlias, communityId, electionId)
  const apiQuery = useApiPollInfo(electionId)

  if (apiQuery.isLoading || apiQuery.error) {
    return <Check isLoading={apiQuery.isLoading} error={apiQuery.error} />
  }

  // Merge contract and API data
  let results = {
    electionId,
    ...apiQuery.data,
  } as PollInfo

  if (connected && data?.date) {
    results = merge.withOptions(
      {
        mergeArrays: false,
      },
      results,
      contractDataToObject(data)
    ) as PollInfo
  }

  return <PollView loading={false} onChain={!!data} poll={results} />
}

export default CommunityPoll
