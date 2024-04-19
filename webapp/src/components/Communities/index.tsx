import {Box, Button, Heading, Link, SimpleGrid, Text, VStack} from '@chakra-ui/react'
import {useQuery} from '@tanstack/react-query'
import {MdOutlineGroupAdd} from 'react-icons/md'
import {Link as RouterLink} from 'react-router-dom'
import {fetchCommunities} from '../../queries/communities'
import {useAuth} from '../Auth/useAuth'
import {Check} from '../Check'
import {CommunityCard} from './Card'

export const CommunitiesList = () => {
  const {bfetch} = useAuth()
  const {data, error, isLoading} = useQuery({
    queryKey: ['communities'],
    queryFn: fetchCommunities(bfetch),
  })

  return (
    <VStack spacing={4} w='full' alignItems='start'>
      <Heading size='md'>Communities</Heading>
      <SimpleGrid gap={4} w='full' alignItems='start' columns={{base: 1, md: 2, lg: 4}}>
        {data &&
          data.map((community, k) => (
            <CommunityCard key={k} community={community}/>
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
          <Button leftIcon={<MdOutlineGroupAdd/>}>Create a community</Button>
        </Link>
      </Box>
    </VStack>
  )
}
