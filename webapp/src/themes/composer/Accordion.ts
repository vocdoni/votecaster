import { accordionAnatomy } from '@chakra-ui/anatomy'
import { createMultiStyleConfigHelpers } from '@chakra-ui/react'

const { definePartsStyle, defineMultiStyleConfig } = createMultiStyleConfigHelpers(accordionAnatomy.keys)

const composer = definePartsStyle({
  root: {
    border: '0px solid transparent',
    w: 'full',
  },
  button: {
    p: 0,
  },
  panel: {
    p: 0,
    mt: 3,
  },
})

export const Accordion = defineMultiStyleConfig({
  variants: { composer },
  defaultProps: {
    variant: 'composer',
  },
})
