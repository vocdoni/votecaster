import { inputAnatomy } from '@chakra-ui/anatomy'
import { createMultiStyleConfigHelpers } from '@chakra-ui/react'

const { definePartsStyle, defineMultiStyleConfig } = createMultiStyleConfigHelpers(inputAnatomy.keys)

const outline = definePartsStyle({
  field: {
    bg: 'transparent',
    borderColor: 'purple.border',
    _hover: {
      borderColor: 'purple.500',
    },
    _focus: {
      borderColor: 'purple.500',
      boxShadow: '0 0 0 2px #9f7aea',
    },
    _placeholder: {
      color: 'white',
    },
  },
})

export const Input = defineMultiStyleConfig({ variants: { outline } })
