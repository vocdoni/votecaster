import { Box, Button, Heading, Image, Link, SimpleGrid, Spinner, useBreakpointValue } from '@chakra-ui/react'
import { useEffect, useMemo, useState } from 'react'
import { FaDownload } from 'react-icons/fa6'
import { useParams } from 'react-router-dom'
import { CsvGenerator } from '../generator'

export const Poll = () => {
  const { pid } = useParams()
  const [voters, setVoters] = useState([])
  const [loaded, setLoaded] = useState<boolean>(false)
  const [loading, setLoading] = useState<boolean>(false)

  useEffect(() => {
    if (loaded || loading || !pid) return
    ;(async () => {
      try {
        setLoading(true)
        const response = await fetch(`${import.meta.env.APP_URL}/votersOf/${pid}`)
        const data = await response.json()
        setVoters(data.voters)
      } catch (e) {
        console.error(e)
      } finally {
        setLoaded(true)
        setLoading(false)
      }
    })()
  }, [])
  const columns = useBreakpointValue({
    base: 1, // default is for mobile devices
    sm: 2, // for medium screens and up
    md: 3, // for large screens and up
    lg: 4, // for extra large screens and up
  })

  const usersfile = useMemo(() => {
    if (!voters.length) return { url: '', filename: '' }

    return new CsvGenerator(
      ['Username'],
      voters.map((username) => [username])
    )
  }, [voters])

  return (
    <Box w={{ base: 'full', lg: '50%' }} textAlign='center'>
      <Image src={`${import.meta.env.APP_URL}/preview/${pid}`} mb={10} fallback={<Spinner />} />
      {voters.length > 0 && (
        <Box pt={5} px={{ base: 5, sm: 0 }} borderTop='1px solid gray' textAlign='initial'>
          <Heading display='flex' justifyContent='space-between'>
            Voters
            <Link href={usersfile.url} download={'voters-list.csv'}>
              <Button alignSelf='end' fontWeight='normal' variant='text' rightIcon={<FaDownload />}>
                download list
              </Button>
            </Link>
          </Heading>
          <SimpleGrid columns={columns}>
            {voters.map((username, index) => (
              <Box key={index}>
                <Link href={`https://warpcast.com/${username}`} isExternal>
                  {username}
                </Link>
              </Box>
            ))}
          </SimpleGrid>
        </Box>
      )}
    </Box>
  )
}
