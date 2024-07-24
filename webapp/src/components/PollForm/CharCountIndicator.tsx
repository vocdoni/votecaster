import { Box, BoxProps, Text } from '@chakra-ui/react'
import { FC } from 'react'

type CharCountIndicatorProps = BoxProps & {
  currentLength: number
  maxLength: number
}

export const CharCountIndicator: FC<CharCountIndicatorProps> = ({ currentLength, maxLength, ...rest }) => {
  return (
    <Box position='absolute' bottom='8px' right='8px' color='gray.500' fontSize='sm' {...rest}>
      <Text>{`${currentLength}/${maxLength}`}</Text>
    </Box>
  )
}
