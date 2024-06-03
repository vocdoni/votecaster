import { Box } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { useAuth } from '~components/Auth/useAuth'
import { Check } from '~components/Check'
import { CommunitiesList } from '~components/Communities'
import { Pagination } from '~components/Pagination'
import { fetchCommunitiesByAdmin } from '~queries/communities'
import { pageToOffset } from '~util/mappings'

const MyCommunitiesList = () => {
  const { bfetch, profile } = useAuth()
  const { page }: { page?: string } = useParams()
  const p = Number(page || 1)

  const { data, error, isLoading } = useQuery({
    queryKey: ['communities', 'byAdmin', profile?.fid, p],
    queryFn: fetchCommunitiesByAdmin(bfetch, profile!, { offset: pageToOffset(p) }),
    enabled: profile != null,
  })

  return (
    <Box w='full'>
      {(error || isLoading) && <Check error={error} isLoading={isLoading} />}
      <CommunitiesList data={data?.communities || []} />
      <Pagination total={data?.pagination.total || 0} page={p} path='/communities/mine/:page?' />
    </Box>
  )
}

export default MyCommunitiesList
