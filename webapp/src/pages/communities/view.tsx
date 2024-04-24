import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { useAuth } from '../../components/Auth/useAuth'
import { Check } from '../../components/Check'
import { CommunitiesView } from '../../components/Communities/View'
import { fetchCommunity } from '../../queries/communities'
import { Text } from '@chakra-ui/react'

const Community = () => {
  const { id } = useParams()
  const { bfetch } = useAuth()
  const { data: community, isLoading, error } = useQuery({
    queryKey: ['community', id],
    queryFn: fetchCommunity(bfetch, id),
    enabled: !!id,
  })

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }
  return <CommunitiesView community={community} />
}

export default Community
