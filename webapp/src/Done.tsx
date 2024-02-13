import { Box, Button, Code, IconButton, Text, useClipboard } from '@chakra-ui/react'
import { FaArchway, FaCheck, FaRegCopy } from 'react-icons/fa6'

const appUrl = import.meta.env.APP_URL
const pollUrl = (pid: string) => `${appUrl}/${pid}`
const cast = (uri: string) => window.open(`https://warpcast.com/~/compose?embeds[]=${encodeURIComponent(pollUrl(uri))}`)

export const Done = ({ pid }: { pid: string }) => {
  const { hasCopied, onCopy } = useClipboard(pollUrl(pid))

  return (
    <>
      <Text display='inline'>Done! You can now cast it using this link:</Text>
      <Box display='flex' alignItems='center' gap={1}>
        <Code isTruncated maxW='95%'>
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
      <Button colorScheme='purple' rightIcon={<FaArchway />} onClick={() => cast(pid)}>
        Cast it!
      </Button>
    </>
  )
}
