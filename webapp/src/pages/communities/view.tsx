import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'
import { Check } from '~components/Check'
import { CommunitiesView } from '~components/Communities/View'
import { fetchCommunity } from '~queries/communities'

const Community = () => {
  const { id, chain } = useParams()
  const { bfetch } = useAuth()
  const {
    data: community,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['community', chain, id],
    queryFn: fetchCommunity(bfetch, id as string),
    enabled: !!id,
  })

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  return <CommunitiesView chain={chain as string} community={community} refetch={refetch} />
}

export default Community
