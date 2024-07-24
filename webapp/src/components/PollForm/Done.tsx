import { Box, Button, Code, Icon, IconButton, Image, Link, Text, useClipboard } from '@chakra-ui/react'
import { useMemo } from 'react'
import { useFormContext } from 'react-hook-form'
import { FaCheck, FaDownload, FaRegCopy } from 'react-icons/fa6'
import { FarcasterLogo } from '~components/FarcasterLogo'
import { appUrl } from '~constants'
import { CsvGenerator } from '~src/generator'
import { usePollForm } from './usePollForm'

const pollUrl = (pid: string) => `${appUrl}/${pid}`
const cast = (uri: string) => window.open(`https://warpcast.com/~/compose?embeds[]=${encodeURIComponent(uri)}`)

export const Done = () => {
  const { shortened, pid, usernames, setUsernames, censusRecords, setPid } = usePollForm()
  const { hasCopied, onCopy } = useClipboard(shortened ?? pollUrl(pid as string))
  const { reset } = useFormContext()

  const usersfile = useMemo(() => {
    if (!usernames.length) return { url: '', filename: '' }

    return new CsvGenerator(
      ['Username'],
      usernames.map((username) => [username])
    )
  }, [usernames])

  return (
    <>
      <Text display='inline'>Done! You can now cast it using this link:</Text>
      <Box display='flex' alignItems='center' justifyContent='space-between' overflow='hidden'>
        <Code overflowX='auto' whiteSpace='nowrap' flex={1} isTruncated>
          {shortened ?? pollUrl(pid as string)}
        </Code>
        <IconButton
          aria-label='Copy to clipboard'
          colorScheme='purple'
          icon={hasCopied ? <FaCheck /> : <FaRegCopy />}
          size='xs'
          onClick={onCopy}
          cursor='pointer'
          p={1.5}
          title={hasCopied ? 'Copied!' : 'Copy to clipboard'}
        />
      </Box>
      <Image src={`${appUrl}/preview/${pid}`} alt='poll preview' />
      <Button
        colorScheme='purple'
        rightIcon={
          <Icon height={24} mt='px'>
            <FarcasterLogo fill='white' />
          </Icon>
        }
        onClick={() => cast(shortened ?? pollUrl(pid as string))}
      >
        Cast it!
      </Button>
      <Box fontSize='xs' textAlign='right'>
        or{' '}
        <Button
          variant='text'
          size='xs'
          p={0}
          height='auto'
          onClick={() => {
            reset()
            setPid(null)
            setUsernames([])
          }}
        >
          create a new one
        </Button>
      </Box>
      {usernames.length > 0 && (
        <Box>
          <Text>
            You created a census for a total of {usernames.length} farcaster users, containing{' '}
            {Math.round((usernames.length / censusRecords) * 1000) / 10}% of the specified census.{` `}
            <Link download={'census-usernames.csv'} href={usersfile.url}>
              Download usernames list <Icon as={FaDownload} />
            </Link>
          </Text>
        </Box>
      )}
    </>
  )
}
