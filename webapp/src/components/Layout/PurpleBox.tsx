import { Box, BoxProps } from '@chakra-ui/react'

export const PurpleBox = (props: BoxProps) => (
  <Box boxShadow='md' borderRadius='md' bg='purple.100' p={4} display='flex' flexDir='column' gap={2} {...props} />
)
