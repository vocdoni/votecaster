import { Box, BoxProps, Link, Stack, Text } from '@chakra-ui/react'
import axios from 'axios'
import { useEffect, useState } from 'react'

const appUrl = import.meta.env.APP_URL

export const TopTenPolls = (props: BoxProps) => {
  const [polls, setPolls] = useState([])

  useEffect(() => {
    ;(async () => {
      try {
        const response = await axios.get(`${appUrl}/rankings/pollsByVotes`)
        const { polls } = response.data
        if (polls && polls.length) {
          setPolls(polls)
        }
      } catch (e) {
        console.error('error fetching polls:', e)
      }
    })()
  }, [])

  if (!polls.length) return null

  return (
    <Box bg={'gray.800'} p={5} borderRadius='lg' {...props}>
      <Text fontSize='xl' mb={4} fontWeight='600' color='white'>
        Top 10 polls (by votes)
      </Text>
      <Stack spacing={3}>
        {polls.map((poll, index) => (
          <Box
            key={index}
            p={3}
            bg={'gray.700'}
            borderRadius='md'
            display='flex'
            justifyContent='space-between'
            flexDir={{ base: 'column', sm: 'row' }}
            gap={{ base: 0, sm: 2 }}
          >
            <Link
              href={`https://warpcast.com/${poll.createdByUsername}`}
              isExternal
              color='teal.300'
              fontWeight='medium'
            >
              {poll.title} — by {poll.createdByUsername}
            </Link>
            <Text color='gray.200' alignSelf='end'>
              {poll.voteCount} votes
            </Text>
          </Box>
        ))}
      </Stack>
    </Box>
  )
}
