import { ChakraProvider } from '@chakra-ui/react'
import { RainbowKitProvider } from '@rainbow-me/rainbowkit'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React from 'react'
import ReactDOM from 'react-dom/client'
import { WagmiProvider } from 'wagmi'
import { AuthProvider } from './components/Auth/AuthContext'
import { Router } from './router'
import { theme } from './theme'
import { config } from './util/rainbow'

import '@rainbow-me/rainbowkit/styles.css'

const queryClient = new QueryClient()
const rootElement = document.getElementById('root')

if (!rootElement) {
  throw new Error('could not find root element :\\')
}

ReactDOM.createRoot(rootElement).render(
  <React.StrictMode>
    <ChakraProvider theme={theme}>
      <WagmiProvider config={config}>
        <QueryClientProvider client={queryClient}>
          <RainbowKitProvider>
            <AuthProvider>
              <Router />
            </AuthProvider>
          </RainbowKitProvider>
        </QueryClientProvider>
      </WagmiProvider>
    </ChakraProvider>
  </React.StrictMode>
)
