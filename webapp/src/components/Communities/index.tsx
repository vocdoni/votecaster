import {Box, Button, Heading, Link, SimpleGrid, Text, VStack, Stack} from '@chakra-ui/react'
import {useQuery} from '@tanstack/react-query'
import {Link as RouterLink} from 'react-router-dom'
import {fetchCommunities, fetchCommunitiesByAdmin} from '../../queries/communities'
import {Community} from '../../util/types'

import {useAuth} from '../Auth/useAuth'
import {Check} from '../Check'
import {CommunityCard} from './Card'
import {FaUsers, FaRegStar} from "react-icons/fa";
import {useCallback, useState} from "react";

export const CommunitiesList = () => {
  const {bfetch, profile, isAuthenticated} = useAuth()

  // state to show only the communities the user is part of
  const [showMyCommunities, setShowMyCommunities] = useState(false)

  // callback to toggle showMyCommunities
  const toggleMyCommunities = useCallback(() => {
    setShowMyCommunities(!showMyCommunities)
  }, [showMyCommunities])

  const {data: allCommunities, error: allCommunitiesError, isLoading: isAllCommunitiesLoading} = useQuery({
    queryKey: ['communities'],
    queryFn: fetchCommunities(bfetch),
  })
  const {data: myCommunities, error: myCommunitiesError, isLoading: isMyCommunitiesLoading} = useQuery({
    queryKey: ['communities', 'byAdmin'],
    queryFn: () => fetchCommunitiesByAdmin(bfetch, profile!),
    enabled: profile != null,
  })
  // Filter by community admins fid in case showMyCommunities is true
  const filteredData = showMyCommunities ? myCommunities : allCommunities

  return (
    <VStack spacing={4} w='full' alignItems='start'>
      <Heading size='md'>Communities</Heading>

      {isAuthenticated &&
        <ToggleStateComponent state={showMyCommunities} toggleState={toggleMyCommunities} state1text={"All communities"}
                              state2text={"My communities"}/>}
      <SimpleGrid gap={4} w='full' alignItems='start' columns={{base: 1, md: 2, lg: 4}}>
        {filteredData && filteredData.map((community: Community, k: number) => (
            <CommunityCard name={community.name} slug={community.id.toString()} key={k} pfpUrl={community.logoURL} admins={community.admins} disabled={community.disabled}/>
        ))}
      </SimpleGrid>
      <Check error={allCommunitiesError || myCommunitiesError} isLoading={isAllCommunitiesLoading || isMyCommunitiesLoading}/>
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

interface IToggleStateComponentProps {
  state: boolean
  toggleState: () => void
  state1text: string
  state2text: string
}

export const ToggleStateComponent = ({state, toggleState, state1text, state2text}: IToggleStateComponentProps) => {
  return (
    <Stack direction='row' align='center'>
      <Button onClick={() => {
        if (state) toggleState()
      }} colorScheme='gray' size='xs' leftIcon={<FaUsers/>} variant={!state ? 'solid' : 'outline'}>
        {state1text}
      </Button>
      <Button onClick={() => {
        if (!state) toggleState()
      }} colorScheme='gray' size='xs' leftIcon={<FaRegStar/>} variant={state ? 'solid' : 'outline'}>
        {state2text}
      </Button>
    </Stack>
  )

}
