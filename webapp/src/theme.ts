import { defineStyleConfig, extendTheme } from '@chakra-ui/react'

export const theme = extendTheme({
  fonts: {
    heading: '"Inter", sans-serif',
    body: '"Inter", sans-serif',
  },
  components: {
    Link: defineStyleConfig({
      baseStyle: {
        _hover: {
          color: 'purple.500',
          textDecoration: 'none',
        },
      },
      variants: {
        primary: {
          color: 'purple.500',
        },
      },
    }),
  },
})
