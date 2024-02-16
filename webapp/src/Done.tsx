import { Box, Button, Code, IconButton, Image, Text, useClipboard } from '@chakra-ui/react'
import { Dispatch, SetStateAction } from 'react'
import { useFormContext } from 'react-hook-form'
import { FaArchway, FaCheck, FaRegCopy } from 'react-icons/fa6'

const appUrl = import.meta.env.APP_URL
const pollUrl = (pid: string) => `${appUrl}/${pid}`
const cast = (uri: string) => window.open(`https://warpcast.com/~/compose?embeds[]=${encodeURIComponent(pollUrl(uri))}`)

export const Done = ({ pid, setPid }: { pid: string; setPid: Dispatch<SetStateAction<string | null>> }) => {
  const { hasCopied, onCopy } = useClipboard(pollUrl(pid))
  const { reset } = useFormContext()

  return (
    <>
      <Text display='inline'>Done! You can now cast it using this link:</Text>
      <Box display='flex' alignItems='center' justifyContent='space-between' overflow='hidden'>
        <Code overflowX='auto' whiteSpace='nowrap' flex={1} isTruncated>
          {pollUrl(pid)}
        </Code>
        <IconButton
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
      <Button colorScheme='purple' rightIcon={<FaArchway />} onClick={() => cast(pid)}>
        Cast it!
      </Button>
      <Box fontSize='xs' align='right'>
        or{' '}
        <Button
          variant='text'
          size='xs'
          p={0}
          height='auto'
          onClick={() => {
            reset()
            setPid(null)
          }}
        >
          create a new one
        </Button>
      </Box>
    </>
  )
}
