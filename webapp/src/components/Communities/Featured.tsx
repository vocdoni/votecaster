import { Heading, SimpleGrid, VStack } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useAuth } from '~components/Auth/useAuth'
import { Check } from '~components/Check'
import { CommunityCard } from '~components/Communities/Card'
import { fetchFeatured } from '~queries/communities'

export const FeaturedCommunities = () => {
  const { bfetch } = useAuth()
  const { data, isLoading, error } = useQuery({
    queryKey: ['communities', 'featured'],
    queryFn: fetchFeatured(bfetch, { limit: 4 }),
  })

  return (
    <VStack spacing={8} w='full'>
      <Heading as='h2'>Featured communities</Heading>
      {(isLoading || error) && <Check isLoading={isLoading} error={error} />}
      <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} w='full' gap={{ base: 4, lg: 20 }}>
        {data?.communities.map((community) => (
          <CommunityCard
            name={community.name}
            slug={community.id.toString()}
            pfpUrl={community.logoURL}
            key={community.id}
          />
        ))}
      </SimpleGrid>
    </VStack>
  )
}
