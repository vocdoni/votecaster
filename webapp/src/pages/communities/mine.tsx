import { useQuery } from '@tanstack/react-query'
import { useAuth } from '~components/Auth/useAuth'
import { Check } from '~components/Check'
import { CommunitiesList } from '~components/Communities'
import { fetchCommunitiesByAdmin } from '~queries/communities'

const MyCommunitiesList = () => {
  const { bfetch, profile } = useAuth()
  const { data, error, isLoading } = useQuery({
    queryKey: ['communities', 'byAdmin'],
    queryFn: () => fetchCommunitiesByAdmin(bfetch, profile!),
    enabled: profile != null,
  })

  if (error || isLoading) {
    return <Check error={error} isLoading={isLoading} />
  }

  if (!data) return

  return <CommunitiesList data={data} />
}

export default MyCommunitiesList
