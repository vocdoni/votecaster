import { useParams } from 'react-router-dom'
import { Check } from '~components/Check'
import { useDegenHealthcheck } from '~components/Healthcheck/use-healthcheck'
import { PollView } from '~components/Poll'
import { useApiPollInfo, useContractPollInfo } from '~queries/polls'
import { contractDataToObject } from '~util/mappings'

const CommunityPoll = () => {
  const { pid: electionId, id: communityId } = useParams()
  const { connected } = useDegenHealthcheck()
  const contractQuery = useContractPollInfo(communityId, electionId)
  const apiQuery = useApiPollInfo(electionId)

  if (apiQuery.isLoading || apiQuery.error) {
    return <Check isLoading={apiQuery.isLoading} error={apiQuery.error} />
  }

  // Merge contract and API data
  let results = {
    electionId,
    ...apiQuery.data,
  } as PollInfo

  if (connected && contractQuery.data?.date) {
    results = {
      ...results,
      ...contractDataToObject(contractQuery.data),
    }
  }

  return <PollView loading={false} onChain={!!contractQuery.data} poll={results} />
}

export default CommunityPoll
