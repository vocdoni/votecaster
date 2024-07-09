import { Alert, AlertDescription, AlertIcon } from '@chakra-ui/react'
import { getChain } from '~util/chain'
import { useHealthcheck } from './use-healthcheck'

type ChainStatusProps = {
  alias: ChainKey
}

export const ChainStatus = ({ alias }: ChainStatusProps) => {
  const health = useHealthcheck()
  const chain = getChain(alias)

  if (health[alias]) return null

  return (
    <Alert status='warning'>
      <AlertIcon />
      <AlertDescription>
        {chain.name} seems down right now. Because of this, some information might be missing or outdated, and the page
        may misbehave.
      </AlertDescription>
    </Alert>
  )
}
