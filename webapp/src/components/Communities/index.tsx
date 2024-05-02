import { SimpleGrid } from '@chakra-ui/react'
import { CommunityCard } from './Card'

export const CommunitiesList = ({ data }: { data: Community[] }) => (
  <SimpleGrid gap={4} w='full' alignItems='start' columns={{ base: 1, md: 2, lg: 3, xl: 4 }}>
    {data &&
      data.map((community: Community, k: number) => (
        <CommunityCard
          name={community.name}
          slug={community.id.toString()}
          key={k}
          pfpUrl={community.logoURL}
          admins={community.admins}
          disabled={community.disabled}
        />
      ))}
  </SimpleGrid>
)
