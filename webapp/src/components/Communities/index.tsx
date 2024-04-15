import { Box, Button, Heading, Link, SimpleGrid, Text, VStack } from '@chakra-ui/react'
import { MdOutlineGroupAdd } from 'react-icons/md'
import { Link as RouterLink } from 'react-router-dom'
import { CommunityCard } from './Card'

export const CommunitiesList = () => {
  return (
    <VStack spacing={4} w='full' alignItems='start'>
      <Heading size='md'>Communities</Heading>
      <SimpleGrid gap={4} w='full' alignItems='start' columns={{ base: 1, md: 2, lg: 4 }}>
        {[1, 2, 3, 4, 5, 6, 7, 8].map((i, k) => (
          <CommunityCard name={`Community ${i}`} slug={i} key={k} pfpUrl='https://i.imgur.com/Y3NHD20.jpg' />
        ))}
      </SimpleGrid>
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
          <Button leftIcon={<MdOutlineGroupAdd />}>Create a community</Button>
        </Link>
      </Box>
    </VStack>
  )
}
