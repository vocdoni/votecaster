import { Avatar, Box, Button, Heading, HStack, Link, SimpleGrid, Text, VStack } from '@chakra-ui/react'
import { MdOutlineGroupAdd } from 'react-icons/md'
import { Link as RouterLink } from 'react-router-dom'

export const CommunitiesList = () => {
  return (
    <VStack spacing={4} w='full' alignItems='start'>
      <Heading size='md'>Communities</Heading>
      <SimpleGrid gap={4} w='full' alignItems='start' columns={{ base: 1, md: 2, lg: 4 }}>
        {[1, 2, 3, 4, 5, 6, 7, 8].map((i, k) => (
          <Link
            key={k}
            as={RouterLink}
            to={`/communities/${i}`}
            w='full'
            border='1px solid'
            borderColor='gray.200'
            borderRadius='md'
            p={2}
            boxShadow='sm'
            borderRadius='lg'
            bg='white'
            _hover={{ boxShadow: 'none', bg: 'purple.100' }}
          >
            <HStack>
              <Avatar src='https://i.imgur.com/Y3NHD20.jpg' />
              <Text fontWeight='500'>Community {i}</Text>
            </HStack>
          </Link>
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
