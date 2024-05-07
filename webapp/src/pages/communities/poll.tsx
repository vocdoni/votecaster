import { useParams } from 'react-router-dom'
import { Check } from '~components/Check'
import { PollView } from '~components/Poll'
import { useApiPollInfo, useContractPollInfo } from '~queries/polls'
import { contractDataToObject } from '~util/mappings'

const CommunityPoll = () => {
  const { pid: electionId, id: communityId } = useParams()
  const contractQuery = useContractPollInfo(communityId, electionId)
  const apiQuery = useApiPollInfo(electionId)

  if (contractQuery.isLoading || apiQuery.isLoading || contractQuery.error || apiQuery.error) {
    return (
      <Check isLoading={contractQuery.isLoading || apiQuery.isLoading} error={contractQuery.error || apiQuery.error} />
    )
  }

  // Merge contract and API data
  const results = {
    electionId,
    ...contractDataToObject(contractQuery.data),
    ...apiQuery.data,
  } as PollInfo

  return <PollView loading={false} onChain={!!contractQuery.data} poll={results} />
}

export default CommunityPoll
