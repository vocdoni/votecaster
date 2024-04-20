import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { useAuth } from '../../components/Auth/useAuth'
import { Check } from '../../components/Check'
import { CommunitiesView } from '../../components/Communities/View'
import { fetchCommunity } from '../../queries/communities'

const Community = () => {
  const { id } = useParams()
  const { bfetch } = useAuth()
  const { data, isLoading, error } = useQuery({
    queryKey: ['community'],
    queryFn: fetchCommunity(bfetch, id),
  })

  if (isLoading || error) {
    return <Check isLoading={isLoading} error={error} />
  }

  return <CommunitiesView community={data} />
}

export default Community