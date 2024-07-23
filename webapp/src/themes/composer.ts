import { extendTheme, ThemeConfig } from '@chakra-ui/react'
import { Accordion } from './composer/Accordion'
import { Button } from './composer/Button'
import { Input } from './composer/Input'

const config: ThemeConfig = {
  initialColorMode: 'dark',
  useSystemColorMode: false,
}

type ThemeProps = {
  colorMode: Exclude<typeof config.initialColorMode, undefined>
}

export const composer = extendTheme({
  config,
  styles: {
    global: ({ colorMode }: ThemeProps) => ({
      body: {
        bg: colorMode === 'dark' ? 'transparent' : '#232323',
      },
      // this is required in order to override the overflow: hidden on the chakra-collapse component
      // otherwise the select menus inside the collapse are hidden
      ':not(.chakra-dont-set-collapse) > .chakra-collapse': {
        overflow: 'initial !important',
      },
    }),
  },
  colors: {
    purple: {
      500: '#9f7aea',
      border: '#412e43',
    },
  },
  components: {
    Accordion,
    Button,
    Input,
  },
})
