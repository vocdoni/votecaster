import {Box, Button, Heading, Link, SimpleGrid, Text, VStack, Stack, useRadio} from '@chakra-ui/react'
import {useQuery} from '@tanstack/react-query'
import {Link as RouterLink} from 'react-router-dom'
import {fetchCommunities} from '../../queries/communities'
import {useAuth} from '../Auth/useAuth'
import {Check} from '../Check'
import {CommunityCard} from './Card'
import {FaUsers, FaRegStar} from "react-icons/fa";
import {useCallback, useState} from "react";

export const CommunitiesList = () => {
  const {bfetch} = useAuth()
  const {profile} = useAuth()

  // state to show only the communities the user is part of
  const [showMyCommunities, setShowMyCommunities] = useState(false)

  // callback to toggle showMyCommunities
  const toggleMyCommunities = useCallback(() => {
    setShowMyCommunities(!showMyCommunities)
  }, [showMyCommunities])

  const {data, error, isLoading} = useQuery({
    queryKey: ['communities'],
    queryFn: fetchCommunities(bfetch),
  })

  // Filter by community admins fid in case showMyCommunities is true
  const filteredData = showMyCommunities ?
    data?.filter(community => community.admins.map((admin) => admin.fid).includes(profile?.fid)) : data

  return (
    <VStack spacing={4} w='full' alignItems='start'>
      <Heading size='md'>Communities</Heading>
      <Stack direction='row' align='center'>
        <Button onClick={() => {
          if (showMyCommunities) toggleMyCommunities()
        }} colorScheme='gray' size='xs' leftIcon={<FaUsers/>}
                variant={!showMyCommunities ? 'solid' : 'outline'}>
          All communities
        </Button>
        <Button onClick={() => {
          if (!showMyCommunities) toggleMyCommunities()
        }} colorScheme='gray' size='xs' leftIcon={<FaRegStar/>}
                variant={showMyCommunities ? 'solid' : 'outline'}>
          My communities
        </Button>
      </Stack>
      <SimpleGrid gap={4} w='full' alignItems='start' columns={{base: 1, md: 2, lg: 4}}>
        {filteredData &&
          filteredData.map((community, k) => (
            <CommunityCard name={community.name} slug={community.id} key={k} pfpUrl={community.logoURL}
                           admins={community.admins}/>
          ))}
      </SimpleGrid>
      <Check error={error} isLoading={isLoading}/>
      <Box
        w='full'
        boxShadow='sm'
        borderRadius='lg'
        minHeight={300}
        display='flex'
        flexDir='column'
        alignItems='center'
        justifyContent='center'
        bg='white'
        p={10}
        textAlign='center'
        gap={4}
      >
        <Text fontSize='larger' fontWeight='500'>
          Create your own community and start managing its governance
        </Text>
        <Link as={RouterLink} to='/communities/new'>
          <Button>Create a community on 🎩 Degenchain</Button>
        </Link>
      </Box>
    </VStack>
  )
}
