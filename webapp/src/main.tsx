import { ChakraProvider } from '@chakra-ui/react'
import { AuthKitProvider } from '@farcaster/auth-kit'
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import { theme } from './theme'

const rootElement = document.getElementById('root')
ReactDOM.createRoot(rootElement).render(
  <React.StrictMode>
    <ChakraProvider theme={theme}>
      <AuthKitProvider>
        <App />
      </AuthKitProvider>
    </ChakraProvider>
  </React.StrictMode>
)
