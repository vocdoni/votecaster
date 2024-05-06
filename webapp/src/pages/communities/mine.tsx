import { Box } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useAuth } from '~components/Auth/useAuth'
import { Check } from '~components/Check'
import { CommunitiesList } from '~components/Communities'
import { Pagination } from '~components/Pagination'
import { fetchCommunitiesByAdmin } from '~queries/communities'

const MyCommunitiesList = () => {
  const { bfetch, profile } = useAuth()
  const [offset, setOffset] = useState<number>(0)
  const [total, setTotal] = useState<number>(0)

  const { data, error, isLoading } = useQuery({
    queryKey: ['communities', 'byAdmin', profile?.fid, offset],
    queryFn: fetchCommunitiesByAdmin(bfetch, profile!, { offset }),
    enabled: profile != null,
  })

  useEffect(() => {
    if (!data?.pagination.total) return
    setTotal(data.pagination.total)
  }, [data?.pagination.total])

  return (
    <Box w='full'>
      {(error || isLoading) && <Check error={error} isLoading={isLoading} />}
      <CommunitiesList data={data?.communities || []} />
      <Pagination total={total} offset={offset} setOffset={setOffset} />
    </Box>
  )
}

export default MyCommunitiesList
