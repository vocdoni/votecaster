import { defineStyleConfig, extendTheme } from '@chakra-ui/react'

export const theme = extendTheme({
  components: {
    Link: defineStyleConfig({
      baseStyle: {
        _hover: {
          color: 'purple.500',
          textDecoration: 'none',
        },
      },
    }),
  },
})
