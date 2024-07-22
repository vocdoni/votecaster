import { extendTheme } from '@chakra-ui/react'
import { Button } from './composer/Button'
import { Input } from './composer/Input'

export const composer = extendTheme({
  styles: {
    global: {
      body: {
        bg: window.location.href.includes('light') ? '#232323' : 'transparent',
        color: 'white',
      },
    },
  },
  colors: {
    purple: {
      500: '#9f7aea',
      border: '#412e43',
    },
  },
  components: {
    Input,
    Button,
  },
})
