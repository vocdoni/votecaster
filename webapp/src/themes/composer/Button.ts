import { defineStyle, defineStyleConfig } from '@chakra-ui/react'

const outline = defineStyle({
  color: 'white',
  borderColor: 'purple.border',
  _hover: {
    bgColor: 'purple.800',
  },
})

export const Button = defineStyleConfig({
  variants: { outline },
})
