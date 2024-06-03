import { ChakraProvider } from '@chakra-ui/react'
import { RainbowKitProvider } from '@rainbow-me/rainbowkit'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { degen, mainnet } from 'viem/chains'
import { WagmiProvider } from 'wagmi'
import { BlockchainProvider } from '~components/Blockchains/BlockchainProvider'
import { BlockchainRegistryProvider } from '~components/Blockchains/BlockchainRegistry'
import { HealthcheckProvider } from '~components/Healthcheck/HealthcheckProvider'
import { AuthProvider } from './components/Auth/AuthContext'
import { Router } from './router'
import { theme } from './theme'
import { config } from './util/rainbow'

import '@rainbow-me/rainbowkit/styles.css'

const queryClient = new QueryClient()

const Providers = () => (
  <ChakraProvider theme={theme}>
    <WagmiProvider config={config}>
      <QueryClientProvider client={queryClient}>
        <ReactQueryDevtools initialIsOpen={false} />
        <RainbowKitProvider>
          <HealthcheckProvider>
            <BlockchainRegistryProvider>
              <BlockchainProvider chain={mainnet}>
                <BlockchainProvider chain={degen}>
                  <AuthProvider>
                    <Router />
                  </AuthProvider>
                </BlockchainProvider>
              </BlockchainProvider>
            </BlockchainRegistryProvider>
          </HealthcheckProvider>
        </RainbowKitProvider>
      </QueryClientProvider>
    </WagmiProvider>
  </ChakraProvider>
)

export default Providers
